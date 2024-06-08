package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	isSimulationFromFlags = flag.Bool("simulation", false, "Run the simulation")
)

func main() {
	log.Println("Starting metrics server on port 80 /metrics")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":19998", nil)
	}()

	flag.Parse()

	worldMap := NewMap()
	server := NewServer(worldMap)

	server.Start()
}
