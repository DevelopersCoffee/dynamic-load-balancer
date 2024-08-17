package strategy

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"

	"com.developerscoffee/dynamic-load-balancer/common"
)

type BalancingStrategy interface {
	Init([]*common.Backend)
	GetNextBackend(common.IncomingReq) *common.Backend
	RegisterBackend(*common.Backend)
	PrintTopology()
}

// Round-Robin Balancing
type RRBalancingStrategy struct {
	Index    int
	Backends []*common.Backend
}

func (s *RRBalancingStrategy) Init(backends []*common.Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *RRBalancingStrategy) GetNextBackend(_ common.IncomingReq) *common.Backend {
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

// Hashed Balancing
type HashedBalancingStrategy struct {
	OccupiedSlots []int
	Backends      []*common.Backend
}

func (s *HashedBalancingStrategy) Init(backends []*common.Backend) {
	s.OccupiedSlots = []int{}
	s.Backends = []*common.Backend{}
	for _, backend := range backends {
		key := hash(backend.String())
		if len(s.OccupiedSlots) == 0 {
			s.OccupiedSlots = append(s.OccupiedSlots, key)
			s.Backends = append(s.Backends, backend)
			continue
		}

		index := sort.Search(len(s.OccupiedSlots), func(i int) bool {
			return s.OccupiedSlots[i] >= key
		})

		if index == len(s.OccupiedSlots) {
			s.OccupiedSlots = append(s.OccupiedSlots, key)
			s.Backends = append(s.Backends, backend)
		}
	}
}

func hash(s string) int {
	h := md5.New()
	var sum int = 0
	io.WriteString(h, s)
	for _, b := range h.Sum(nil) {
		sum += int(b)
	}
	return sum % 19
}
