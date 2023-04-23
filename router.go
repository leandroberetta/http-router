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
	IsStatic   bool
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

func (r *Router) AddStaticRoute(path, dir string) {
	fs := http.FileServer(http.Dir(dir))
	regexp, _ := regexp.Compile(path + "/*")
	route := Route{
		Path: path,
		Handler: func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, path)
			fs.ServeHTTP(w, req)
		},
		IsStatic: true,
		Regexp:   regexp,
		Method:   http.MethodGet,
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

func (r *Router) Static(path, dir string) {
	r.AddStaticRoute(path, dir)
}

func (r *Router) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.Routes {
			// TODO: Support multiple static routes
			if route.matchPath(req.URL.Path) && route.Method == req.Method {
				if route.IsStatic {
					route.Handler(w, req)
				} else {
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
}

func (r *Route) matchPath(path string) bool {
	matched := r.Regexp.MatchString(path)
	if !r.IsStatic {
		return matched && r.Segments == len(strings.Split(path, "/"))-1
	}
	return matched
}

func Parameters(req *http.Request) map[string]string {
	return req.Context().Value(ParametersKey("parameters")).(map[string]string)
}
