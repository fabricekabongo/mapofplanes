package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type Cluster struct {
	Nodes      map[string]Node
	Broadcast  chan []byte
	clusterDNS string
}

func NewCluster(clusterDNS string) *Cluster {
	return &Cluster{
		Nodes:      make(map[string]Node),
		Broadcast:  make(chan []byte),
		clusterDNS: clusterDNS,
	}
}

func (c *Cluster) extractIP(addr net.Addr) string {
	switch v := addr.(type) {
	case *net.TCPAddr:
		return v.IP.String()
	case *net.UDPAddr:
		return v.IP.String()
	default:
		return ""
	}
}

func (c *Cluster) AddNode(host string, port int, broadcast chan []byte) {
	node := Node{
		Host:      host,
		Port:      port,
		Broadcast: broadcast,
	}

	c.Nodes[host] = node

	err := node.Connect()
	if err != nil {
		fmt.Println("Error connecting to node: ", node.Address())
		return
	}
}

func (c *Cluster) RemoveNode(host string) {
	delete(c.Nodes, host)
}

func (c *Cluster) Start() {
	listener, err := net.Listen("tcp", ":20001")
	if err != nil {
		panic(err)
	}
	go c.DiscoverNodes()
	for {
		conn, err := listener.Accept()
		log.Println("New clustering connection from: ", conn.RemoteAddr())
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		node := Node{
			Conn:   conn,
			Reader: bufio.NewReader(conn),
			Writer: bufio.NewWriter(conn),
		}
		if _, ok := c.Nodes[c.extractIP(conn.RemoteAddr())]; ok {
			fmt.Println("Node already exists in cluster: ", node.Address())
			err := conn.Close()
			if err != nil {
				return
			}
			continue
		}
		c.Nodes[c.extractIP(conn.RemoteAddr())] = node

		go node.StartListening()
	}
}
func (c *Cluster) DiscoverNodes() {
	// Periodically resolve DNS and attempt to connect to each pod
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ips, err := net.LookupIP(c.clusterDNS)
		if err != nil {
			fmt.Println("Error resolving DNS: ", err)
			continue
		}

		for _, ip := range ips {
			if _, ok := c.Nodes[ip.String()]; ok {
				continue
			}

			log.Println("Discovered node: ", ip.String())
			c.AddNode(ip.String(), 20001, c.Broadcast)
		}
	}
}

func (c *Cluster) BroadcastMessage(data []byte) {
	for _, node := range c.Nodes {
		err := node.Send(data)
		if err != nil {
			fmt.Println("Error sending message to node: ", node.Address())
			c.RemoveNode(node.Host) // Remove node from cluster if it fails to send message it will be added back in the next iteration
		}
	}
}

func (c *Cluster) Close() {
	for _, node := range c.Nodes {
		err := node.Close()
		if err != nil {
			fmt.Println("Error closing connection to node: ", node.Address())
		}
	}
}

type Node struct {
	Host      string
	Port      int
	Reader    *bufio.Reader
	Writer    *bufio.Writer
	Conn      net.Conn
	Broadcast chan []byte
}

func (n *Node) Connect() error {
	conn, err := net.Dial("tcp", n.Address())
	if err != nil {
		return err
	}
	n.Conn = conn
	n.Reader = bufio.NewReader(conn)
	n.Writer = bufio.NewWriter(conn)

	return nil
}

func (n *Node) Close() error {
	return n.Conn.Close()
}

func (n *Node) Send(data []byte) error {
	_, err := n.Writer.Write(data)
	if err != nil {
		return err
	}

	return n.Writer.Flush()
}

func (n *Node) StartListening() {
	defer func(n *Node) {
		err := n.Close()
		if err != nil {
			fmt.Println("Error closing connection to node: ", n.Address())
		}
	}(n)
	for {
		data, err := n.Reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Error reading from node: ", n.Address())
			return
		}

		n.Broadcast <- data
	}
}

func (n *Node) Address() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
}
