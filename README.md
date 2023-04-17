# HTTP Router

Basic HTTP router built on top of Go's *net/http* library.

## Usage

```go
package main

import (
	"fmt"
	"net/http"

	router "github.com/leandroberetta/http-router"
)

func main() {
	r := router.NewRouter()
	r.Get("/api/namespaces/:namespace/deployments/:deployment", func(w http.ResponseWriter, req *http.Request) {
		namespace := req.Context().Value(router.ParameterName("namespace")).(string)
		deployment := req.Context().Value(router.ParameterName("deployment")).(string)

		fmt.Printf("Namespace: %s / Application: %s", namespace, deployment)
	})
	http.ListenAndServe(":8080", r.Handler())
}

```