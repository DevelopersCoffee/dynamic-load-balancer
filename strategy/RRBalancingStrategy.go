package strategy

import (
	"fmt"
	"sync"

	"com.developerscoffee/dynamic-load-balancer/common"
)

type RRBalancingStrategy struct {
	Index    int
	Backends []*common.Backend
	mu       sync.Mutex
}

func (s *RRBalancingStrategy) Init(backends []*common.Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *RRBalancingStrategy) GetNextBackend(_ common.IncomingReq) *common.Backend {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Index = (s.Index + 1) % len(s.Backends)
	return s.Backends[s.Index]
}

func (s *RRBalancingStrategy) RegisterBackend(backend *common.Backend) {
	s.Backends = append(s.Backends, backend)
}

func (s *RRBalancingStrategy) PrintTopology() {
	for index, backend := range s.Backends {
		fmt.Printf("[%d] %s:%d\n", index, backend.Host, backend.Port)
	}
}

func NewRRBalancingStrategy(backends []*common.Backend) *RRBalancingStrategy {
	strategy := new(RRBalancingStrategy)
	strategy.Init(backends)
	return strategy
}
