package router

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStaticRoute(t *testing.T) {
	r := NewRouter()
	r.Static("/static", "test/static")
	req, err := http.NewRequest("GET", "/static/hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(r.Handler())
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code error: got %v want %v", status, http.StatusOK)
	}
	body, _ := io.ReadAll(rr.Body)
	fmt.Println((string(body)))
	containsHelloWorld := strings.Contains(string(body), "hello")
	if !containsHelloWorld {
		t.Errorf("Body error, expected hello string")
	}
}

func TestStaticRouteIndex(t *testing.T) {
	r := NewRouter()
	r.Static("/static", "test/static")
	req, err := http.NewRequest("GET", "/static/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(r.Handler())
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Status code error: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestAddRoute(t *testing.T) {
	r := NewRouter()
	r.AddParametersRoute("/namespaces/:namespace/deployments/:deployment", http.MethodGet, func(w http.ResponseWriter, req *http.Request) {})
	route := r.ParametersRoutes[0]
	if route.Method != http.MethodGet {
		t.Errorf("Method error, got: %s, want: %s.", route.Method, http.MethodGet)
	}
	if route.Segments != 4 {
		t.Errorf("Segments error, got: %d, want: %d.", route.Segments, 4)
	}
	match := route.matchPath("/namespaces/bookinfo/deployments/ratings")
	if match == false {
		t.Errorf("Match error")
	}
	expectedParameters := []string{"namespace", "deployment"}
	for i, parameter := range route.Parameters {
		if parameter != expectedParameters[i] {
			t.Errorf("Parameter error, got: %s, want: %s.", parameter, expectedParameters[i])
		}
	}
}

func TestGetRequest(t *testing.T) {
	r := NewRouter()
	r.Static("/static", "test/static")
	r.Get("/namespaces/:namespace/deployments/:deployment", func(w http.ResponseWriter, req *http.Request) {
		parameters := Parameters(req)
		namespace := parameters["namespace"]
		deployment := parameters["deployment"]
		w.Write([]byte(namespace + "/" + deployment))
	})
	req, err := http.NewRequest("GET", "/namespaces/bookinfo/deployments/ratings", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(r.Handler())
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code error: got %v want %v", status, http.StatusOK)
	}
	expectedBody := "bookinfo/ratings"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}
