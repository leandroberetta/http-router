package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	Routes []Route
}

type Route struct {
	Path       string
	Parameters []string
	Segments   int
	Method     string
	Handler    http.HandlerFunc
	Regexp     *regexp.Regexp
}

type ParametersKey string

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) AddRoute(path, method string, handler http.HandlerFunc) {
	parameters := []string{}
	for _, p := range strings.Split(path, "/") {
		if parameter, ok := strings.CutPrefix(p, ":"); ok {
			parameters = append(parameters, parameter)
		}
	}
	pathRegexp := path
	for _, parameter := range parameters {
		pathRegexp = strings.Replace(pathRegexp, ":"+parameter, "([a-z]+)", 1)
	}
	regexp, _ := regexp.Compile(pathRegexp)
	route := Route{
		Path:       path,
		Method:     method,
		Segments:   len(strings.Split(path, "/")) - 1,
		Parameters: parameters,
		Handler:    handler,
		Regexp:     regexp,
	}
	r.Routes = append(r.Routes, route)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.AddRoute(path, http.MethodGet, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.AddRoute(path, http.MethodPost, handler)
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.AddRoute(path, http.MethodPut, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.AddRoute(path, http.MethodDelete, handler)
}

func (r *Router) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.Routes {
			if route.matchPath(req.URL.Path) && route.Method == req.Method {
				parameters := route.Regexp.FindStringSubmatch(req.URL.Path)
				parametersMap := make(map[string]string, len(route.Parameters))
				for i, parameter := range route.Parameters {
					parametersMap[parameter] = parameters[i+1]
				}
				ctx := context.WithValue(req.Context(), ParametersKey("parameters"), parametersMap)
				route.Handler(w, req.WithContext(ctx))
			}
		}
	}
}

func (r *Route) matchPath(path string) bool {
	return r.Regexp.MatchString(path) && r.Segments == len(strings.Split(path, "/"))-1
}

func Parameters(req *http.Request) map[string]string {
	return req.Context().Value(ParametersKey("parameters")).(map[string]string)
}
