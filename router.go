package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	ParametersRoutes []ParametersRoute
	StaticRoutes     []StaticRoute
}

type BaseRoute struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
	Regexp  *regexp.Regexp
}

type ParametersRoute struct {
	BaseRoute
	Parameters []string
	Segments   int
	IsStatic   bool
}

type StaticRoute struct {
	BaseRoute
}

type ParametersKey string

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) AddParametersRoute(path, method string, handler http.HandlerFunc) {
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
	route := ParametersRoute{
		BaseRoute: BaseRoute{
			Path:    path,
			Method:  method,
			Handler: handler,
			Regexp:  regexp,
		},
		Parameters: parameters,
		Segments:   len(strings.Split(path, "/")) - 1,
	}
	r.ParametersRoutes = append(r.ParametersRoutes, route)
}

func (r *Router) AddStaticRoute(path, dir string) {
	fs := http.FileServer(http.Dir(dir))
	regexp, _ := regexp.Compile(path + "/*")
	route := StaticRoute{
		BaseRoute: BaseRoute{
			Path: path,
			Handler: func(w http.ResponseWriter, req *http.Request) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, path)
				fs.ServeHTTP(w, req)
			},
			Regexp: regexp,
			Method: http.MethodGet,
		},
	}
	r.StaticRoutes = append(r.StaticRoutes, route)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.AddParametersRoute(path, http.MethodGet, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.AddParametersRoute(path, http.MethodPost, handler)
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.AddParametersRoute(path, http.MethodPut, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.AddParametersRoute(path, http.MethodDelete, handler)
}

func (r *Router) Static(path, dir string) {
	r.AddStaticRoute(path, dir)
}

func (r *Router) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.ParametersRoutes {
			if route.matchPath(req.URL.Path) && route.Method == req.Method {
				parameters := route.Regexp.FindStringSubmatch(req.URL.Path)
				parametersMap := make(map[string]string, len(route.Parameters))
				for i, parameter := range route.Parameters {
					parametersMap[parameter] = parameters[i+1]
				}
				ctx := context.WithValue(req.Context(), ParametersKey("parameters"), parametersMap)
				route.Handler(w, req.WithContext(ctx))
				return
			}
		}
		for _, route := range r.StaticRoutes {
			if route.matchPatch(req.URL.Path) {
				route.Handler(w, req)
				return
			}
		}
	}
}

func (r *ParametersRoute) matchPath(path string) bool {
	return r.Regexp.MatchString(path) && r.Segments == len(strings.Split(path, "/"))-1
}

func (r *StaticRoute) matchPatch(path string) bool {
	return r.Regexp.MatchString(path)
}

func Parameters(req *http.Request) map[string]string {
	return req.Context().Value(ParametersKey("parameters")).(map[string]string)
}
