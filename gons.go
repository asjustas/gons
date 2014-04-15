package main

import (
	"github.com/codegangsta/martini"
	"runtime"
	"log"
	"net/http"
)

func main() {
	conf, err := goini.Load("config.ini")

	if err != nil {
        panic(err)
    }

	m := martini.Classic()
	log.Fatal(http.ListenAndServe(":8080", m))
	runtime.GOMAXPROCS(runtime.NumCPU())
}