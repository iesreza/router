# Simple router
This is a simple web router for golang supporting lazy (optional) parameters
You may check for samples inside the router_test.go

## Quick Guide
#### Domain Filter
Domain filter is used to filter domains. it can parse wildcards or even regex.
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
Group filter is used to filter first part of url. Group filter router does not care about anything but the first part of url and ignore the rest.

Match filter is used to match whole url. It matches whole url as same as filter. match can be combined with group filter for easier routing decision and cleaner code.
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
	//a => alpha num
	//i or id => integer
	//s => string
	handle.Match("article/[i:id]/[a:title]/[s:str]","GET", func(req router.Request){
			fmt.Println("Matched article id:"+req.Parameters["id"])
			fmt.Println("Matched article title only alpha num :"+req.Parameters["title"])
			fmt.Println("Matched some strings:"+req.Parameters["str"])
	})
		
``` 

#### Lazy parameter
Lazy or optional parameter is kind of parameters that you might want to pass it to url or simply ignore it. the router will match and extract the parameter if exist but still will match if the corresponding lazy parameter is not exist.

The syntax is same as common parameters with and additional ~ appended to first part of parameter. the syntax for url of lazy parameters should be like domain.tld/lazy:123

```
    //Following urls should match the router
    //domain.tld/article/1/optional_id:18/title:xyz123
    //domain.tld/article/1/optional_id:18/
    //domain.tld/article/1
   	handle.Match("article/[i:id]/[i:optional_id]/~[a:title]/","GET", func(req router.Request){
   			fmt.Println(req.Parameters)
   	})
```


#### MiddleWares
You may use middlewares to intervene in url matching procedure. middlewares can override matching by return true or false.
```
        handle.Match("article/[i:id]","GET", func(req router.Request){
			fmt.Println("Matched article id"+req.Parameters["id"])
		}).Middleware(func(req router.Request) bool {
          				if req.Req().Host == "test.com"{
          					return false
          				}
          
          				return true
          			})
		
```


#### Static files
You may want to serve static files as assets or simple html files and etc. you may simply create a customized router for this reason.
```
	handler.Group("static",nil, func(handle *router.Route) {
		handle.Static("subdir","./assets/",nil)
	})
	//or
	handler.Static("subdir","./assets/", func(req router.Request) {
    	fmt.Println("User request for static file:"+req.Req().RequestURI)
    })
```