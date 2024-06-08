package main

import (
	"encoding/gob"
	"fabricekabongo.com/geodb/clustering"
	server2 "fabricekabongo.com/geodb/server"
	"fabricekabongo.com/geodb/world"
	"flag"
	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	clusterDNS = os.Getenv("CLUSTER_DNS")
)

func main() {

	populateEnv()
	log.Println("Starting metrics server on port 80 /metrics and 20001 for clustering")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":19998", nil)
		if err != nil {
			return
		}
	}()
	flag.Parse()

	gob.Register(&world.LocationEntity{})
	gob.Register(&world.Grid{})
	gob.Register(&world.Map{})

	worldMap := world.NewMap()

	mList, broadcasts, err := createClustering(clusterDNS, worldMap)
	if err != nil {
		log.Fatal("Failed to create cluster: ", err)
	}

	defer func(mList *memberlist.Memberlist, timeout time.Duration) {
		err := mList.Leave(timeout)
		if err != nil {
			log.Println("Failed to leave cluster: ", err)
		}
	}(mList, 0)

	writer := server2.NewWriteHandler(worldMap, broadcasts)
	reader := server2.NewReadHandler(worldMap)

	server := server2.NewServer(*writer, *reader)

	server.Start()
}

func createClustering(clusterDNS string, world *world.Map) (*memberlist.Memberlist, *memberlist.TransmitLimitedQueue, error) {
	broadcasts := &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return 1 // Replace with the actual number of nodes
		},
		RetransmitMult: 3,
	}

	delegate := clustering.NewBroadcastDelegate(world, broadcasts)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get hostname: ", err)
	}

	config := memberlist.DefaultLocalConfig()
	config.Name = hostname
	config.BindPort = 20001
	config.AdvertisePort = 20001
	config.Delegate = delegate

	if err != nil {
		log.Println("Failed to get hostname: ", err)
	}

	mList, err := memberlist.Create(config)
	if err != nil {
		log.Println("Failed to create cluster: ", err)
		return nil, nil, err
	}

	broadcasts.NumNodes = func() int {
		return mList.NumMembers()
	}

	clusterIPs, err := getClusterIPs(clusterDNS)
	if err != nil {
		log.Println("Failed to get cluster IPs: ", err)
		return nil, nil, err
	}

	_, err = mList.Join(clusterIPs)
	if err != nil {
		log.Println("Failed to join cluster: ", err)
		return nil, nil, err
	}

	return mList, broadcasts, nil
}

func getClusterIPs(clusterDNS string) ([]string, error) {
	ips, err := net.LookupIP(clusterDNS)
	if err != nil {
		return nil, err
	}

	// map addresses to strings
	var clusterIPs []string
	for _, ip := range ips {
		clusterIPs = append(clusterIPs, ip.String())
	}

	return clusterIPs, nil
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
