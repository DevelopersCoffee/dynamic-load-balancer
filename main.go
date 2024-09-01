package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"com.developerscoffee/dynamic-load-balancer/common"
	STRATEGY "com.developerscoffee/dynamic-load-balancer/strategy"
)

type LB struct {
	backends []*common.Backend
	strategy STRATEGY.BalancingStrategy
}

func InitLB(strategyType string) *LB {
	log.Println("Initializing backends")

	backends := []*common.Backend{
		{Host: "localhost", Port: 8081, IsHealthy: true},
		{Host: "localhost", Port: 8082, IsHealthy: true},
		{Host: "localhost", Port: 8083, IsHealthy: true},
		{Host: "localhost", Port: 8084, IsHealthy: true},
	}

	var strategy STRATEGY.BalancingStrategy

	log.Printf("Selected strategy: %s", strategyType)

	if strategyType == "consistent" {
		log.Println("Using Consistent Hashing Strategy")
		consistentStrategy := STRATEGY.NewConsistentHashingStrategy(backends, 100) // 100 vNodes per backend
		consistentStrategy.Init(backends)
		go consistentStrategy.StartHealthCheck()
		strategy = consistentStrategy
		log.Println("Consistent Hashing Strategy initialized successfully")
	} else if strategyType == "round-robin" {
		log.Println("Using Round-Robin Strategy")
		strategy = STRATEGY.NewRRBalancingStrategy(backends)
		log.Println("Round-Robin Strategy initialized successfully")
	} else {
		log.Fatalf("Unknown strategy type: %s", strategyType)
	}

	lb := &LB{
		backends: backends,
		strategy: strategy,
	}
	log.Println("Load Balancer initialized successfully")
	return lb
}

func (lb *LB) proxyHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	incomingReq := common.IncomingReq{
		ReqId: req.RequestURI,
	}

	backend := lb.strategy.GetNextBackend(incomingReq)
	if backend == nil {
		http.Error(w, "No healthy backends available", http.StatusServiceUnavailable)
		log.Println("No healthy backends available to handle the request")
		return
	}

	backendURL := fmt.Sprintf("http://%s:%d%s", backend.Host, backend.Port, req.RequestURI)
	resp, err := http.Get(backendURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Error connecting to backend %s: %s", backend.String(), err.Error())
		backend.IsHealthy = false            // Mark the backend as unhealthy
		lb.strategy.RegisterBackend(backend) // Re-register backend for possible future use
		http.Error(w, "Backend is currently unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	// Define a command-line flag for the strategy
	strategyFlag := flag.String("strategy", "round-robin", "Load balancing strategy (consistent or round-robin)")
	flag.Parse()

	// Get the strategy from the environment variable if the flag is not set
	strategyType := *strategyFlag
	if strategyType == "" {
		strategyType = os.Getenv("LOAD_BALANCER_STRATEGY")
	}

	// Initialize the load balancer with the selected strategy
	log.Println("Initializing Load Balancer")
	lb := InitLB(strategyType)
	log.Println("Load Balancer initialized successfully")

	// Start the HTTP server
	http.HandleFunc("/", lb.proxyHTTP)
	log.Println("Load Balancer is listening on HTTP port 9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
