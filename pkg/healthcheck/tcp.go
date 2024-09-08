package healthcheck

import (
	"fmt"
	"net"
	"time"
)

type TCPChecker struct {
	Host    string
	Port    int
	Timeout time.Duration
}

func (t *TCPChecker) CheckHealth() HealthCheckResult {
	var netClient = &net.Dialer{
		Timeout: time.Second * t.Timeout,
	}

	start := time.Now()
	conn, err := netClient.Dial("tcp", fmt.Sprintf("%s:%d", t.Host, t.Port))
	response_time := time.Since(start)
	isHealthy := false

	if err == nil {
		conn.Close()
		isHealthy = true
	}

	return HealthCheckResult{
		IsHealthy:    isHealthy,
		ResponseTime: response_time,
		Error:        err,
	}
}
