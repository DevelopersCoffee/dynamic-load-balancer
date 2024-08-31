package strategy

import "com.developerscoffee/dynamic-load-balancer/common"

// BalancingStrategy defines the interface for load balancing strategies
type BalancingStrategy interface {
	Init([]*common.Backend)
	GetNextBackend(common.IncomingReq) *common.Backend
	RegisterBackend(*common.Backend)
	PrintTopology()
}
