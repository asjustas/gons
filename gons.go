package main

import (
	"github.com/codegangsta/martini"
	"runtime"
	"log"
	"net/http"
)

func main() {
	m := martini.Classic()
	log.Fatal(http.ListenAndServe(":8080", m))
	runtime.GOMAXPROCS(runtime.NumCPU())
}