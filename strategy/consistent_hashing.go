package strategy

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"com.developerscoffee/dynamic-load-balancer/common"
)

type ConsistentHashingStrategy struct {
	HashRing   []int
	BackendMap map[int]*common.Backend
	mu         sync.Mutex
}

func NewConsistentHashingStrategy(backends []*common.Backend) *ConsistentHashingStrategy {
	strategy := &ConsistentHashingStrategy{
		BackendMap: make(map[int]*common.Backend),
	}

	for _, backend := range backends {
		strategy.addBackend(backend)
	}

	return strategy
}

func (s *ConsistentHashingStrategy) Init(backends []*common.Backend) {
	s.HashRing = []int{}
	s.BackendMap = make(map[int]*common.Backend)
	for _, backend := range backends {
		s.addBackend(backend)
	}
}

func (s *ConsistentHashingStrategy) addBackend(backend *common.Backend) {
	s.mu.Lock()
	defer s.mu.Unlock() // This will automatically unlock at the end of the function

	hash := hash(backend.String())
	s.HashRing = append(s.HashRing, hash)
	s.BackendMap[hash] = backend
	sort.Ints(s.HashRing)

	s.PrintTopology() // This is safe because Unlock will be called after PrintTopology returns
}

func (s *ConsistentHashingStrategy) RegisterBackend(backend *common.Backend) {
	s.addBackend(backend)
}

func (s *ConsistentHashingStrategy) removeBackend(backend *common.Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := hash(backend.String())
	for i, h := range s.HashRing {
		if h == hash {
			s.HashRing = append(s.HashRing[:i], s.HashRing[i+1:]...)
			break
		}
	}
	delete(s.BackendMap, hash)

	s.PrintTopology()
}

func (s *ConsistentHashingStrategy) GetNextBackend(req common.IncomingReq) *common.Backend {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqHash := hash(req.ReqId)

	index := sort.Search(len(s.HashRing), func(i int) bool {
		return s.HashRing[i] >= reqHash
	})

	if index == len(s.HashRing) {
		index = 0
	}

	backend := s.BackendMap[s.HashRing[index]]

	if !backend.IsHealthy {
		log.Printf("Backend %s is down, rehashing...", backend.String())
		s.removeBackend(backend)
		return s.GetNextBackend(req)
	}

	return backend
}

func hash(s string) int {
	h := md5.New()
	var sum int
	io.WriteString(h, s)
	for _, b := range h.Sum(nil) {
		sum += int(b)
	}
	return sum % 1024 // Adjusted mod value for more even distribution
}

func (s *ConsistentHashingStrategy) StartHealthCheck() {
	for {
		for _, backend := range s.BackendMap {
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/health", backend.Host, backend.Port))
			if err != nil || resp.StatusCode != 200 {
				backend.IsHealthy = false
				log.Printf("Backend %s is marked as down", backend.String())
				s.removeBackend(backend)
			} else {
				backend.IsHealthy = true
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *ConsistentHashingStrategy) PrintTopology() {

	fmt.Println("Current Consistent Hash Ring:")
	for _, hash := range s.HashRing {
		backend := s.BackendMap[hash]
		fmt.Printf("Backend %s is at hash %d\n", backend.String(), hash)
	}
}
