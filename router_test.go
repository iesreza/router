package router_test

import (
	"fmt"
	"github.com/iesreza/router"
	"net/http"
	"testing"
)



func TestRouter(t *testing.T) {
	handler := router.GetInstance()
	handler.Domain("*.test.com,test.com,ihwc.ir", func(req router.Request) {
		fmt.Println(req.Req().Host)
	}, func(handle *router.Route) {
		handle.Group("abcd", nil ,func(handle *router.Route) {
			handle.Match("efg/[i:id]/~[i:id2]", "GET", func(req router.Request) {
				req.WriteString( fmt.Sprint(req.Parameters))

			})
		})
	})


	handler.Group("articles", func(req router.Request) {
		//You will reach to here if domain.tld/articles/* matched
		//If you dont need this just pass nil!
	}, func(handle *router.Route) {
		//Sub routers goes here
		handle.Match("article/[i:id]","GET", func(req router.Request){
			fmt.Println("Matched article id"+req.Parameters["id"])
		})
	})


	http.ListenAndServe("0.0.0.0:80", &handler)


}
