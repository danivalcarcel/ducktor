package healthcheck

import (
	"fmt"
	"strings"
	"time"
)

type HealthCheck struct {
	Name               string
	Interface          string
	HealthyThreshold   int
	UnHealthyThreshold int
	Interval           int
	Timeout            time.Duration

	Host     string
	Port     int
	Endpoint string
	Match    int
}

type HealthChecker interface {
	CheckHealth() HealthCheckResult
}

type HealthCheckResult struct {
	IsHealthy    bool
	ResponseTime time.Duration
	Error        error
}

func HealthCheckerFrom(config HealthCheck) (HealthChecker, error) {
	switch strings.ToLower(config.Interface) {
	case "http", "https":
		return &HTTPChecker{
			Endpoint: config.Endpoint,
			Host:     config.Host,
			Port:     config.Port,
			Protocol: config.Interface,
			Match:    config.Match,
			Timeout:  config.Timeout,
		}, nil
	case "tcp":
		return &TCPChecker{
			Host:    config.Host,
			Port:    config.Port,
			Timeout: config.Timeout,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported interface: %s", config.Interface)
	}
}
