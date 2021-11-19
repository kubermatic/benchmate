package main

import (
	bmServer "github.com/kubermatic/benchmate/server"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/stats", bmServer.StatsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
