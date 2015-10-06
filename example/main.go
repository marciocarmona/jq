package main

import (
	"fmt"
	"io"
	"runtime"
	"log"
	"time"
	"net/http"
	"github.com/marciocarmona/jq"
)

func proxyGitHub(rw http.ResponseWriter, req *http.Request) {
//	resp, err := http.Get("https://api.github.com/repos/google/acai")
	resp, err := http.Get("https://api.github.com/users/google/repos?per_page=10")
	if err == nil {
		io.Copy(rw, resp.Body)
	} else {
		fmt.Fprint(rw, err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

type jqHandlerWrapper struct {
	handler http.Handler
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	var p http.HandlerFunc = proxyGitHub
	s := &http.Server{
		Addr:           ":8080",
		Handler:        jq.NewJqHandlerWithPattern(p, `if type == "array" then [.[]|%[1]s] else %[1]s end`),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
