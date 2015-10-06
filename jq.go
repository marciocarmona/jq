package jq

/*
#cgo LDFLAGS: -ljq -all-static
#include <jq.h>
#include <stdlib.h>
*/
import "C"
import (
	"strings"
	"errors"
	"unsafe"
	"runtime"
	"github.com/hashicorp/golang-lru"
)

var (
	errInvalidProgram = errors.New("invalid program")
	errJqStateNil     = errors.New("jq state nil")
)

var cache *lru.Cache

type message struct {
	input []string
	output chan []string
}

type Jq struct {
	jqExecutors []*jqExecutor
	channel 	chan *message
}

func NewJq(program string) (*Jq, error) {
	if (cache == nil) {
		cache, _ = lru.New(128)
	}
	value, ok := cache.Get(program)
	if ok {
		return value.(*Jq), nil
	}
	
	size := runtime.NumCPU()
	jq := &Jq{
		jqExecutors: make([]*jqExecutor, size),
		channel: make(chan *message),
	}
	runtime.SetFinalizer(jq, jqFinalizer)
	
	for i := 0; i < size; i++ {
		jqe, err := newJqExecutor(program)
		if err != nil {
			return nil, err
		}
		jq.jqExecutors[i] = jqe
		go func(jqe *jqExecutor) {
			defer jqe.free()
			for msg := range jq.channel {
				msg.output <- jqe.execute(msg.input...)
			}
		}(jq.jqExecutors[i])
	}
	
	cache.Add(program, jq)
	return jq, nil
}

func (jq *Jq) Execute(input... string) []string {
	msg := &message{
		input: input,
		output: make(chan []string),
	}
	jq.channel <- msg
	return <- msg.output
}

// Jq used to hold state
type jqExecutor struct {
	state *C.jq_state
}

// Free will release all memory from Jq state
func (j *jqExecutor) free() {
	C.jq_teardown(&j.state)
}

func (j *jqExecutor) execute(input... string) []string {
	jsons := strings.Join(input, "")
	parser := C.jv_parser_new(0)
	defer C.jv_parser_free(parser)

	// Make a simple input then convert to CString.
	cjsons := C.CString(jsons)
	defer C.free(unsafe.Pointer(cjsons))
	C.jv_parser_set_buf(parser, cjsons, C.int(len(jsons)), 0)
	
	var response []string
	// Check if v is valid.
	for v := C.jv_parser_next(parser); C.jv_is_valid(v) == 1; v = C.jv_parser_next(parser) {
		C.jq_start(j.state, v, 0)

		// Check if res is valid.
		for res := C.jq_next(j.state); C.jv_is_valid(res) == 1; res = C.jq_next(j.state) {
			response = append(response, toString(res))
		}
		
	}
	return response
}

func toString(jv C.jv) string {
	// Dump it!
	dumped := C.jv_dump_string(jv, 0)
	defer C.jv_free(dumped)

	// Convert dump to string!
	strval := C.jv_string_value(dumped)
	return C.GoString(strval)
}

func jqFinalizer(jq *Jq) {
	close(jq.channel)
}

// NewJq returns an initialized state with a compiled program
// program should be a valid jq program/filter
// see http://stedolan.github.io/jq/manual/#Basicfilters for more info
func newJqExecutor(program string) (*jqExecutor, error) {
	jq := &jqExecutor{
		state: C.jq_init(),
	}

	if jq.state == nil {
		return nil, errJqStateNil
	}

	pgm := C.CString(program)
	defer C.free(unsafe.Pointer(pgm))

	// Compiles a program into jq_state.
	if ok := C.jq_compile(jq.state, pgm); ok != 1 {
		jq.free()
		return nil, errInvalidProgram
	}

	return jq, nil
}