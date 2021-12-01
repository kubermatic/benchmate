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

	http.HandleFunc("/benchmate", bmServer.BenchmateHandler)
	http.HandleFunc("/exit", bmServer.BenchmateHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
