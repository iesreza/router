package router

import (
	"net/http"
	"os"
	"regexp"
	"strings"
)

var defaultRouter = handler{
	routes: []*Route{},
}

type handler struct {
	routes     []*Route
	middleware []func(req Request) bool
	Fallback   func(req Request)
}

type Route struct {
	match       string
	method      string
	domain      string
	static      bool
	staticDir   string
	domainMatch *regexp.Regexp
	routes      []*Route
	tokens      []token
	group       bool
	middleware  []func(req Request) bool
	Callback    func(req Request)
	Fallback    func(req Request)
}

func GetInstance() handler {
	return defaultRouter
}
func (handle *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	uri := strings.Trim(request.RequestURI, "/")
	uriChunks := strings.Split(uri, "?")
	uriTokens := strings.Split(uriChunks[0], "/")
	req := Request{
		writer:     writer,
		request:    request,
		Parameters: map[string]value{},
		Get:        map[string]value{},
		Post:       map[string]value{},
		Matched:    false,
	}
	for _, r := range handle.routes {

		if recursiveMatch(uriTokens, r, &req) {
			return
		}
	}
	if !req.Matched && handle.Fallback != nil {
		req.Matched = true
		handle.Fallback(req)
	}

}

func recursiveMatch(uriTokens []string, handle *Route, req *Request) bool {
	for _, item := range handle.middleware {
		if !item(*req) {
			return false
		}
	}
	if handle.domainMatch != nil && !handle.domainMatch.MatchString(req.Req().Host) {
		return false
	}
	if handle.method != "" && strings.ToLower(req.request.Method) != strings.ToLower(handle.method) {
		return false
	}
	if !handle.group && len(uriTokens) > len(handle.tokens) {

		return false
	}

	p := min(len(uriTokens), len(handle.tokens))
	temp := map[string]value{}
	matched := 0
	lazyMatched := 0
	var i int
	pointer := -1

	for i = 0; i < len(handle.tokens); i++ {
		if handle.tokens[i].lazy {

			index := strings.Index(req.request.RequestURI, handle.tokens[i].varName+":")
			if index > 0 {

				variable := ""
				for i := index + len(handle.tokens[i].varName+":"); i < len(req.request.RequestURI); i++ {
					if req.request.RequestURI[i] == '/' {
						break
					}
					variable += string(req.request.RequestURI[i])
				}
				if handle.tokens[i].match.(*regexp.Regexp).MatchString(variable) {
					temp[handle.tokens[i].varName] = value(variable)
					lazyMatched++
				}
			}
			continue
		}
	}

	i = 0
	for i = 0; i < len(handle.tokens); i++ {
		if handle.tokens[i].lazy {
			continue
		}
		pointer++
		if pointer == len(uriTokens) {

			return false
		}
		if !handle.tokens[i].isMatch(uriTokens[pointer]) {

			return false
		} else {
			matched++
			if handle.tokens[i].matchType == 1 {
				req.Parameters[handle.tokens[i].varName] = value(uriTokens[pointer])
			}
		}
	}

	if !handle.group && matched+lazyMatched != len(uriTokens) {

		return false
	}

	//restore lazy vars
	for k, v := range temp {
		req.Parameters[k] = v
	}

	if handle.static {
		path := handle.staticDir + "/" + strings.Join(uriTokens[p:], "/")
		if fileExists(path) {
			if handle.Callback != nil {
				handle.Callback(*req)
			}
			http.ServeFile(req.writer, req.request, path)
			req.Matched = true
			return true
		} else {
			req.writer.WriteHeader(404)
		}

		return false
	}
	if handle.Callback != nil {
		handle.Callback(*req)
	}

	if !handle.group {
		req.Matched = true
	}
	if handle.group && len(handle.routes) > 0 {
		for _, r := range handle.routes {
			if recursiveMatch(uriTokens[p:], r, req) {
				return true && !r.group
			}
		}
		if !req.Matched && handle.Fallback != nil {
			req.Matched = true
			handle.Fallback(*req)
		}
	}
	return false
}

func min(i int, i2 int) int {
	if i < i2 {
		return i
	}
	return i2
}

func (handle *handler) Group(match string, onMatch func(req Request), onSubRouter func(handle *Route)) *Route {
	newRoute := Route{
		match:    strings.Trim(match, "/"),
		routes:   []*Route{},
		tokens:   tokenize(strings.Trim(match, "/")),
		group:    true,
		Callback: onMatch,
	}
	handle.routes = append(handle.routes, &newRoute)
	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
	return &newRoute
}

func (handle *Route) Group(match string, onMatch func(req Request), onSubRouter func(handle *Route)) *Route {
	newRoute := Route{
		match:    strings.Trim(match, "/"),
		routes:   []*Route{},
		tokens:   tokenize(strings.Trim(match, "/")),
		group:    true,
		Callback: onMatch,
	}
	handle.routes = append(handle.routes, &newRoute)

	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
	return &newRoute
}

func (handle *Route) Match(match string, method string, onMatch func(req Request)) *Route {

	newRoute := Route{
		match:    strings.Trim(match, "/"),
		routes:   []*Route{},
		method:   method,
		tokens:   tokenize(strings.Trim(match, "/")),
		Callback: onMatch,
	}

	handle.routes = append(handle.routes, &newRoute)
	return &newRoute
}

func (handle *handler) Match(match string, method string, onMatch func(req Request)) *Route {
	newRoute := Route{
		match:    strings.Trim(match, "/"),
		routes:   []*Route{},
		method:   method,
		tokens:   tokenize(strings.Trim(match, "/")),
		Callback: onMatch,
	}

	handle.routes = append(handle.routes, &newRoute)
	return &newRoute
}

func (handle *handler) Domain(match string, onMatch func(req Request), onSubRouter func(handle *Route)) *Route {
	chunks := split(match, ",; ")
	regex := ""
	for _, item := range chunks {
		regex += "(" + item + ")|"
	}
	regex = strings.Trim(regex, "|")
	regex = strings.Replace(regex, ".", "\\.", -1)
	regex = strings.Replace(regex, "*", "[a-zA-Z0-9\\-]*", -1)

	newRoute := Route{
		match:       strings.Trim(match, "/"),
		routes:      []*Route{},
		group:       true,
		domain:      match,
		domainMatch: regexp.MustCompile(regex),
		Callback:    onMatch,
	}
	handle.routes = append(handle.routes, &newRoute)
	if onSubRouter != nil {
		onSubRouter(&newRoute)
	}
	return &newRoute
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

func (route *Route) Middleware(middlewares ...func(req Request) bool) *Route {
	for _, item := range middlewares {
		route.middleware = append(route.middleware, item)
	}
	return route
}
func (route *handler) Middleware(middlewares ...func(req Request) bool) *handler {
	for _, item := range middlewares {
		route.middleware = append(route.middleware, item)
	}
	return route
}

func (handle *handler) Static(match string, dir string, onMatch func(req Request)) *Route {
	newRoute := Route{
		match:     strings.Trim(match, "/"),
		routes:    []*Route{},
		static:    true,
		staticDir: strings.Trim(dir, "/"),
		tokens:    tokenize(strings.Trim(match, "/")),
		group:     true,
		Callback:  onMatch,
	}
	handle.routes = append(handle.routes, &newRoute)
	return &newRoute
}

func (handle *Route) Static(match string, dir string, onMatch func(req Request)) *Route {
	newRoute := Route{
		match:     strings.Trim(match, "/"),
		routes:    []*Route{},
		static:    true,
		staticDir: strings.Trim(dir, "/"),
		tokens:    tokenize(strings.Trim(match, "/")),
		group:     true,
		Callback:  onMatch,
	}
	handle.routes = append(handle.routes, &newRoute)
	return &newRoute
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
