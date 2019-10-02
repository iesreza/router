package router

import (
	"net/http"
	"strconv"
)

type Request struct {
	writer http.ResponseWriter
	request *http.Request
	Parameters map[string]value
	Get map[string]value
	Post map[string]value
}

func NewRequest(writer http.ResponseWriter,request *http.Request) Request  {
	req := Request{
		writer:writer,
		request:request,
	}

	return req
}

func (r *Request)Header(statusCode int)  {
	r.writer.WriteHeader(statusCode)
}

func (r *Request)Write(bytes []byte)  {
	r.writer.Write(bytes)
}

func (r *Request)WriteString(s string)  {
	r.writer.Write([]byte(s))
}

type value string


func (r *Request)Req() *http.Request {
	return r.request
}

func (v value)Int() int  {
	i,_ := strconv.Atoi(string(v))
	return i
}

func (v value)Float() float64  {
	f,_ := strconv.ParseFloat(string(v),32)
	return f
}

func (v value)String() string  {
	return string(v)
}
