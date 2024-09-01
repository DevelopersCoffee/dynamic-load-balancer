package strategy

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"com.developerscoffee/dynamic-load-balancer/common"
	"github.com/spaolacci/murmur3"
)

type ConsistentHashingStrategy struct {
	HashRing   []uint32
	BackendMap map[uint32]*common.Backend
	mu         sync.Mutex
	vnodes     int
}

func NewConsistentHashingStrategy(backends []*common.Backend, vnodes int) *ConsistentHashingStrategy {
	strategy := &ConsistentHashingStrategy{
		BackendMap: make(map[uint32]*common.Backend),
		vnodes:     vnodes,
	}

	for _, backend := range backends {
		strategy.addBackend(backend)
	}

	return strategy
}

func (s *ConsistentHashingStrategy) Init(backends []*common.Backend) {
	s.HashRing = []uint32{}
	s.BackendMap = make(map[uint32]*common.Backend)
	for _, backend := range backends {
		s.addBackend(backend)
	}
}

func (s *ConsistentHashingStrategy) addBackend(backend *common.Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < s.vnodes; i++ {
		virtualNodeKey := fmt.Sprintf("%s-%d", backend.String(), i)
		hashValue := murmur3.Sum32([]byte(virtualNodeKey)) // Use Murmur3 to generate the hash value
		s.HashRing = append(s.HashRing, hashValue)
		s.BackendMap[hashValue] = backend
	}
	sort.Slice(s.HashRing, func(i, j int) bool { return s.HashRing[i] < s.HashRing[j] })

	s.PrintTopology()
}

func (s *ConsistentHashingStrategy) RegisterBackend(backend *common.Backend) {
	s.addBackend(backend)
}

func (s *ConsistentHashingStrategy) removeBackend(backend *common.Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < s.vnodes; i++ {
		virtualNodeKey := fmt.Sprintf("%s-%d", backend.String(), i)
		hashValue := murmur3.Sum32([]byte(virtualNodeKey))
		index := s.findHashIndex(hashValue)
		s.HashRing = append(s.HashRing[:index], s.HashRing[index+1:]...)
		delete(s.BackendMap, hashValue)
	}

	s.PrintTopology()
}

func (s *ConsistentHashingStrategy) GetNextBackend(req common.IncomingReq) *common.Backend {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.HashRing) == 0 {
		log.Println("No healthy backends available")
		return nil
	}

	reqHash := murmur3.Sum32([]byte(req.ReqId)) // Use Murmur3 to generate the hash value
	index := s.findHashIndex(reqHash)

	for i := 0; i < len(s.HashRing); i++ {
		currentIndex := (index + i) % len(s.HashRing)
		backend := s.BackendMap[s.HashRing[currentIndex]]
		if backend.IsHealthy {
			return backend
		}
	}

	log.Println("No healthy backends found after checking all backends")
	return nil
}

func (s *ConsistentHashingStrategy) findHashIndex(hashValue uint32) int {
	index := sort.Search(len(s.HashRing), func(i int) bool {
		return s.HashRing[i] >= hashValue
	})

	if index == len(s.HashRing) {
		index = 0
	}

	return index
}

func (s *ConsistentHashingStrategy) StartHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, backend := range s.BackendMap {
			go func(backend *common.Backend) {
				resp, err := http.Get(fmt.Sprintf("http://%s:%d/health", backend.Host, backend.Port))
				if err != nil || resp.StatusCode != 200 {
					if backend.IsHealthy {
						log.Printf("Backend %s is marked as down", backend.String())
						backend.IsHealthy = false
						s.removeBackend(backend)
					}
				} else {
					if !backend.IsHealthy {
						log.Printf("Backend %s is back online", backend.String())
						backend.IsHealthy = true
						s.addBackend(backend)
					}
				}
			}(backend)
		}
	}
}

func (s *ConsistentHashingStrategy) PrintTopology() {
	fmt.Println("=== Consistent Hash Ring Topology ===")
	totalNodes := len(s.HashRing)
	vnodeDistribution := make(map[string]int)
	hashRanges := make(map[string][]uint32)

	// Count the number of virtual nodes for each backend
	for _, hash := range s.HashRing {
		backend := s.BackendMap[hash]
		vnodeDistribution[backend.String()]++
		hashRanges[backend.String()] = append(hashRanges[backend.String()], hash)
	}

	// Print details for each backend
	for backend, count := range vnodeDistribution {
		fmt.Printf("Backend %s\n", backend)
		fmt.Printf("  - Number of vNodes: %d\n", count)
		fmt.Printf("  - Proportion of Hash Space: %.2f%%\n", float64(count)/float64(totalNodes)*100)
		fmt.Printf("  - Hash Range: %d to %d\n", hashRanges[backend][0], hashRanges[backend][count-1])
		//fmt.Println("  - vNode Hashes: ", hashRanges[backend])
		fmt.Println()
	}

	// Display overall distribution
	fmt.Println("=== Summary ===")
	fmt.Printf("Total virtual nodes in ring: %d\n", totalNodes)
	fmt.Printf("Total backends: %d\n", len(vnodeDistribution))
}
