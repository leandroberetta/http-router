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
		parameters := Parameters(req)
		namespace := parameters["namespace"]
		deployment := parameters["deployment"]
		w.Write([]byte(namespace + "/" + deployment))
	})
	http.ListenAndServe(":8080", r.Handler())
}

```