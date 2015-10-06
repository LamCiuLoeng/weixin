package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world, from docker")
	})
	http.ListenAndServe("0.0.0.0:5000", nil)
}
