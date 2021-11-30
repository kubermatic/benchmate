package main

import (
	"flag"
	bmServer "github.com/kubermatic/benchmate/server"
	"log"
	"net/http"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on")
	flag.Parse()

	http.HandleFunc("/bm", bmServer.StatsHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
