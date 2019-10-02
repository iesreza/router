# Simple router
This is simple web router for golang supporting lazy (optional) variables
You may check for samples inside the router_test.go

## Quick Guide
#### Domain Filter
Domain filter is used to filter domain. it can parse wildcards or even regex.
```
    handler := router.GetInstance()
    
	handler.Domain("*.test.com,test.com,domain2.com", func(req router.Request) {
		//Any thing you want to do if domain matches
		fmt.Println(req.Req())
	}, func(handle *router.Route) {
		//Sub routers goes here
		handle.Match("index.html","GET", func(req router.Request){
			fmt.Println("Matched")
		})
	})
	
```

#### Group And Match Filters
Group filter is used to filter first part of url. in group filter router does not care to anything but the first part of url and ignore the rest.

Match filter is used to match whole url. it means it will check all parts of url not only part of it. match can be combined with group filter for easier routing decision and cleaner code.
```
    handler := router.GetInstance()
    
	handler.Group("articles", func(req router.Request) {
	    //You will reach to here if domain.tld/articles/* matched 
		//If you dont need this just pass nil!
	}, func(handle *router.Route) {
		//Sub routers goes here
		handle.Match("article/[i:id]","GET", func(req router.Request){
			fmt.Println("Matched article id"+req.Parameters["id"])
		})
	})
	
```

### Parameters
You might want to read url parameters it can simply done by few simple switches
```
	handle.Match("article/[i:id]/[a:title]/[s:str]","GET", func(req router.Request){
			fmt.Println("Matched article id:"+req.Parameters["id"])
			fmt.Println("Matched article title only alpha num :"+req.Parameters["title"])
			fmt.Println("Matched some strings:"+req.Parameters["str"])
	})
		
``` 

#### Lazy parameter
Lazy or optional parameter is kind of parameters that you might want to pass it to url or simply ignore it. the router will match and extract the parameter if exist but still will match if the lazy parameter is not matched.

The syntax is same as common parameters with and additional ~ appended to first part of parameter. the syntax for url of lazy parameters should be like domain.tld/lazy:123
```
    //Following urls should match the router
    //domain.tld/article/1/optional_id:18/title:xyz123
    //domain.tld/article/1/optional_id:18/
    //domain.tld/article/1
   	handle.Match("article/[i:id]/[i:optional_id]/~[a:title]/","GET", func(req router.Request){
   			fmt.Println(req.Parameters)
   	})
	

