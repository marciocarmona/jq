package jq

import (
	"bytes"
	"fmt"
	"net/http"
)

type bufferedResponseWriter struct {
	header *http.Header
	buffer *bytes.Buffer
	status *int
}

func (w bufferedResponseWriter) Header() http.Header {
	return *w.header
}

func (w bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w bufferedResponseWriter) WriteHeader(status int) {
	w.status = &status
}

type jqHandler struct {
	handler http.Handler
	pattern string
}

func NewJqHandler(handler http.Handler) (*jqHandler) {
	return &jqHandler{handler, "%s"}
}

func NewJqHandlerWithPattern(handler http.Handler, pattern string) (*jqHandler) {
	return &jqHandler{handler, pattern}
}

func (h jqHandler) applyPattern(jqExpression string) string {
	return fmt.Sprintf(h.pattern, jqExpression)
}

func (h jqHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	jqExpression := req.FormValue("jq")
	if jqExpression == "" {
		h.handler.ServeHTTP(w, req)
		return
	}
	
	j, err := NewJq(h.applyPattern(jqExpression))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	
	header := make(http.Header)
	buffer := bytes.Buffer{}
	bw := bufferedResponseWriter {
		header: &header,
		buffer: &buffer,
	}
	h.handler.ServeHTTP(bw, req)
	
	if bw.status != nil && *bw.status < 200 && *bw.status > 299 {
		w.WriteHeader(*bw.status)
		w.Write(bw.buffer.Bytes())
		return
	}
	
	if bw.status != nil {
		w.WriteHeader(*bw.status)
	}
	
	for _, val := range j.Execute(bw.buffer.String()) {
		fmt.Fprint(w, val)
	}
	
}