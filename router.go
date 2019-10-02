package router

import (
	"net/http"
	"regexp"
	"strings"
)

var defaultRouter = handler{
	routes:[]*Route{},
}
type handler struct {
	routes []*Route
}

type Route struct {
	match  string
	method string
	domain string
	domainMatch *regexp.Regexp
	routes []*Route
	tokens []token
	group  bool
	callback func(req Request)
}

func GetInstance() handler {
	return defaultRouter
}
func (handle *handler)ServeHTTP(writer http.ResponseWriter,request *http.Request){
	uri := strings.Trim(request.RequestURI,"/")
	uriTokens := strings.Split(uri,"/")
	req := Request{
		writer:writer,
		request:request,
		Parameters: map[string]value{},
		Get: map[string]value{},
		Post: map[string]value{},
	}
	for _,r := range handle.routes{

		if recursiveMatch(uriTokens,r,&req){
			return
		}
	}


}

func recursiveMatch(uriTokens []string,handle *Route,req *Request) bool{

	if handle.domainMatch != nil && !handle.domainMatch.MatchString(req.Req().Host){
		return false
	}
	if handle.method != "" && strings.ToLower(req.request.Method) != strings.ToLower(handle.method){
		return false
	}
	if !handle.group && len(uriTokens) > len(handle.tokens){

		return false
	}

	p := min(len(uriTokens),len(handle.tokens))
	temp := map[string]value{}
	matched := 0
	lazyMatched := 0
	var i int
	pointer := -1

	for i = 0; i < len(handle.tokens); i++{
		if handle.tokens[i].lazy{

			index := strings.Index(req.request.RequestURI,handle.tokens[i].varName+":")
			if index > 0{

				variable := ""
				for i:=index+len(handle.tokens[i].varName+":"); i < len(req.request.RequestURI); i++{
					if req.request.RequestURI[i] == '/'{
						break
					}
					variable += string(req.request.RequestURI[i])
				}
				if handle.tokens[i].match.(*regexp.Regexp).MatchString(variable){
					temp[handle.tokens[i].varName] = value(variable)
					lazyMatched++
				}
			}
			continue
		}
	}

	i = 0

	for i = 0; i < len(handle.tokens); i++{
		if handle.tokens[i].lazy{
			continue
		}
		pointer++
		if pointer == len(uriTokens){

			return false
		}
		if !handle.tokens[i].isMatch(uriTokens[pointer]){

			return false
		}else{
			matched++
			if handle.tokens[i].matchType == 1{
				req.Parameters[handle.tokens[i].varName] = value(uriTokens[pointer])
			}
		}
	}


	if !handle.group && matched+lazyMatched != len(uriTokens){

		return false
	}


	//restore lazy vars
	for k,v := range temp{
		req.Parameters[k] = v
	}

	if handle.callback != nil {
		handle.callback(*req)
	}

	if len(handle.routes) > 0{
		for _,r := range handle.routes{
			if recursiveMatch(uriTokens[p:],r,req){
				return true && !r.group
			}
		}
	}
	return false
}





func min(i int, i2 int) int {
	if i < i2{
		return i
	}
	return i2
}

func (handle *handler)Group(match string,onMatch func(req Request),onSubRouter func(handle *Route) ){
	newRoute := Route{
		match: strings.Trim(match,"/"),
		routes:[]*Route{},
		tokens:tokenize(strings.Trim(match,"/")),
		group:true,
		callback:onMatch,
	}
	handle.routes = append(handle.routes,&newRoute)
	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
}


func (handle *Route)Group(match string,onMatch func(req Request),onSubRouter func(handle *Route) ){
	newRoute := Route{
		match: strings.Trim(match,"/"),
		routes:[]*Route{},
		tokens:tokenize(strings.Trim(match,"/")),
		group:true,
		callback:onMatch,
	}
	handle.routes = append(handle.routes,&newRoute)

	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
}


func (handle *Route)Match(match string,method string,onMatch func(req Request) ){

	newRoute := Route{
		match: strings.Trim(match,"/"),
		routes:[]*Route{},
		method:method,
		tokens:tokenize(strings.Trim(match,"/")),
		callback:onMatch,
	}


	handle.routes = append(handle.routes,&newRoute)
}

func (handle *handler)Match(match string,method string,onMatch func(req Request) ){
	newRoute := Route{
		match: strings.Trim(match,"/"),
		routes:[]*Route{},
		method:method,
		tokens:tokenize(strings.Trim(match,"/")),
		callback:onMatch,
	}

	handle.routes = append(handle.routes,&newRoute)
}

func (handle *handler)Domain(match string,onMatch func(req Request),onSubRouter func(handle *Route) ){
	chunks := split(match,",; ")
	regex := ""
	for _,item := range chunks{
		regex += "("+item+")|"
	}
	regex = strings.Trim(regex,"|")
	regex = strings.Replace(regex,".","\\.",-1)
	regex = strings.Replace(regex,"*","[a-zA-Z0-9\\-]*",-1)

	newRoute := Route{
		match: strings.Trim(match,"/"),
		routes:[]*Route{},
		group:true,
		domain:match,
		domainMatch:regexp.MustCompile(regex),
		callback:onMatch,
	}
	handle.routes = append(handle.routes,&newRoute)
	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
}


func split(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}
