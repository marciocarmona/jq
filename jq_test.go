package jq

import (
	"bufio"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

type errorable interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

var largeInput []string = make([]string, 0)

func TestMain(m *testing.M) {
	f, e := os.Open("_testdata/repos.json")
	if e != nil {
        panic(e)
    }
	defer f.Close()
	
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		largeInput = append(largeInput, scanner.Text())
    }
	
	os.Exit(m.Run())
}

func TestNewJqValid(t *testing.T) {
	jq, err := NewJq(".")

	if err != nil {
		t.Errorf("expected error to be nil, got: %s", err)
	}

	if jq == nil {
		t.Error("expected jq to not be nil")
	}
	if jq.jqExecutors == nil {
		t.Error("expected jq.jqExecutors to not be nil")
	}
	if len(jq.jqExecutors) == 0 {
		t.Error("expected jq.jqExecutors to not be empty")
	}
	
	if jq.jqExecutors[0].state == nil {
		t.Error("expected jq.state to not be empty")
	}

}

func TestNewJqInvalid(t *testing.T) {

	jq, err := NewJq("INVALID")
	if err != errInvalidProgram {
		t.Errorf("expected error to be %s, got: %s", errInvalidProgram, err)
	}

	if jq != nil {
		t.Error("expected jq to be nil")
	}
}

func executeJq(t errorable, expression, input string) []string {
	jq, err := NewJq(expression)

	if err != nil {
		t.Errorf("expected error to be nil, got: %s", err)
	}

	if jq == nil {
		t.Error("expected jq to not be nil")
	}
	if jq.jqExecutors == nil {
		t.Error("expected jq.jqExecutors to not be nil")
	}
	if len(jq.jqExecutors) == 0 {
		t.Error("expected jq.jqExecutors to not be empty")
	}
	
	if jq.jqExecutors[0].state == nil {
		t.Error("expected jq.state to not be empty")
	}
	
	return jq.Execute(input)
}

func TestNewJqExecute(t *testing.T) {

	out := executeJq(t, `.args[] | select(.name == "b", .name == "d") | .value`,
		`{
			"args": [
				{"name": "a", "value": 1},
				{"name": "b", "value": 2},
				{"name": "c", "value": 3},
				{"name": "d", "value": 4}
			]
		}`)
	
	if len(out) != 2 {
		t.Error("expected out to have exactly 1 element")
	}
	if out[0] != "2" {
		t.Error("expected to return the value 2")
	}
	if out[1] != "4" {
		t.Error("expected to return the value 4")
	}
}

func BenchmarkJqExecuteSimpleCompile(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
			NewJq(`{name}`)
        }
    })
}

func BenchmarkJqExecuteSimpleCompileExecute(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
			jq, _ := NewJq(`.name`)
            x := jq.Execute(`{
					"name": "aaaa" 
				}`)
            if x[0] != `"aaaa"` {
            	b.Error("Invalid value, should be aaaa", x[0])
            }
        }
    })
}

func BenchmarkJqExecuteSimpleExecutePreCompiled(b *testing.B) {
	jq, _ := NewJq(`.name`)
	b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            x := jq.Execute(`{
					"name": "aaaa" 
				}`)
            if x[0] != `"aaaa"` {
            	b.Error("Invalid value, should be aaaa", x[0])
            }
	    }
    })
}

func BenchmarkJqExecuteSimpleLargeInput(b *testing.B) {
	jq, _ := NewJq(`.name`)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
        for pb.Next() {
            x := jq.Execute(largeInput[i%len(largeInput)])
            if x[0] == "" {
            	b.Error("Invalid value, should be not empty or null", x[0])
            }
        	i++
	    }
    })
}

func BenchmarkJqExecuteComplexLargeInput(b *testing.B) {
	jq, _ := NewJq(`[[[to_entries[] | select(((.value | type) == "string") and (.value[0:5] == "https")) | .value] | sort[] | if (. | contains("api.")) then "OK" else "NOK" end] | group_by(.)[] | {(.[0]): length}] | add`)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
        for pb.Next() {
            x := jq.Execute(largeInput[i%len(largeInput)])
            if !strings.Contains(x[0], `"OK":36`) {
            	b.Error(`Invalid value, should contain ["OK":36]`, x[0])
            }
        	i++
	    }
    })
}

func BenchmarkJqExecuteMappingLargeInput(b *testing.B) {
	jq, _ := NewJq(`{id, name, full_name, language, owner: (.owner | {id, login, type})}`)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
        for pb.Next() {
            x := jq.Execute(largeInput[i%len(largeInput)])
            if !strings.Contains(x[0], `"login":"google"`) {
            	b.Error(`Invalid value, should contain ["login":"google"]`, x[0])
            }
        	i++
	    }
    })
}

func BenchmarkJqCompileLRU(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
			NewJq(`{name` + strconv.Itoa(rand.Int()) + `}`)
        }
    })
}

func BenchmarkJqCompileExecuteLRU(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
			jq, _ := NewJq(`{name,x` + strconv.Itoa(rand.Int()) + `} | .name`)
            x := jq.Execute(`{
					"name": "aaaa" 
				}`)
            if x[0] != `"aaaa"` {
            	b.Error("Invalid value, should be aaaa", x[0])
            }
        }
    })
}