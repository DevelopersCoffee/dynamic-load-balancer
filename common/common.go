package common

import (
	"fmt"
	"net"
)

// Backend structure
type Backend struct {
	Host        string
	Port        int
	IsHealthy   bool
	NumRequests int
}

// IncomingReq structure
type IncomingReq struct {
	SrcConn net.Conn
	ReqId   string
}

// Implement the String method for Backend
func (b *Backend) String() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}
