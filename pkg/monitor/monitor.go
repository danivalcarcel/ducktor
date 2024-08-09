package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ducktor/pkg/healthcheck"
	"ducktor/pkg/service"
)

type Monitor struct {
	Services []service.Service
}

var (
	healthMu sync.Mutex
)

func NewMonitor(configs []healthcheck.HealthCheck) (*Monitor, error) {
	services := make([]service.Service, len(configs))

	for i, config := range configs {
		checker, err := healthcheck.NewHealthChecker(config)
		if err != nil {
			return nil, fmt.Errorf("error while creating HealthChecker: %s", err)
		}

		services[i] = service.Service{
			Name:               config.Name,
			Checker:            checker,
			Interval:           time.Duration(config.Interval) * time.Second,
			UnHealthyThreshold: config.UnHealthyThreshold,
			HealthyThreshold:   config.HealthyThreshold,
			IsHealthy:          false,
		}
	}

	return &Monitor{Services: services}, nil
}

func health(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK

		w.Header().Set("Content-Type", "application/json")

		response := make(map[string]string)

		healthMu.Lock()
		defer healthMu.Unlock()

		for _, svc := range m.Services {
			if !svc.IsHealthy {
				status = http.StatusServiceUnavailable
				response[svc.Name] = "KO"
			} else {
				response[svc.Name] = "OK"
			}
		}

		w.WriteHeader(status)

		json.NewEncoder(w).Encode(response)
	}
}

func serve(m *Monitor, port int) {
	http.HandleFunc("/health", health(*m))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (m *Monitor) Run(port int) {

	go serve(m, port)

	for idx := range m.Services {
		svc := &m.Services[idx]

		go func(s *service.Service) {
			for {
				result := s.Check()

				healthMu.Lock()

				if result.IsHealthy {
					s.UnhealthyCount = 0
					s.HealthyCount++

					if s.HealthyCount >= s.HealthyThreshold {
						log.Printf("Service %s is healthy (%d consecutive successes)\n", s.Name, s.HealthyCount)
						s.IsHealthy = true
					}

				} else {
					s.HealthyCount = 0
					s.UnhealthyCount++

					if s.UnhealthyCount >= s.UnHealthyThreshold {
						log.Printf("Service %s is unhealthy (%d consecutive failures)\n", s.Name, s.UnhealthyCount)
						s.IsHealthy = false
					}
				}

				healthMu.Unlock()

				time.Sleep(s.Interval)
			}
		}(svc)
	}

	// Keep the main function alive
	select {}
}
