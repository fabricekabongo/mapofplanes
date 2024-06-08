package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

var (
	clusterDNS = os.Getenv("CLUSTER_DNS")
)

func main() {

	populateEnv()
	log.Println("Starting metrics server on port 80 /metrics")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":19998", nil)
	}()

	flag.Parse()

	clusters := NewCluster(clusterDNS)

	worldMap := NewMap()
	server := NewServer(worldMap, *clusters)

	server.Start()
}

func populateEnv() {
	if clusterDNS == "" {
		log.Println("Please set the following environment variables:")
		log.Println("CLUSTER_DNS")
		log.Println("Reverting to flags...")

		flag.StringVar(&clusterDNS, "cluster-dns", "", "Cluster DNS")
		flag.Parse()

		if clusterDNS == "" {
			log.Println("No flags set. Please set the following flags:")
			log.Println("cluster-dns")
			os.Exit(1)
		}
	}
}
