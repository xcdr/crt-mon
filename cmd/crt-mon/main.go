package main

import (
	"context"
	"crt-mon/pkg/certexp"
	"crt-mon/pkg/config"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version = "dev"
	build   = "none"
	author  = "undefined"
)

type metricsServer struct {
	elapsedDays *prometheus.GaugeVec
	checkError  *prometheus.GaugeVec
	options     *config.Options
	domains     []certexp.Domain
	httpServer  http.Server
}

func newMetricsServer(address string, port int, options *config.Options) *metricsServer {
	srv := metricsServer{
		httpServer: http.Server{Addr: fmt.Sprintf("%s:%d", address, port)},
	}

	srv.options = options
	// srv.hosts = make([]certexp.HostInfo, 0)

	srv.elapsedDays = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "crt_mon_elapsed_days",
		Help: "The total number of elapsed days",
	}, []string{"host", "address"})

	srv.checkError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "crt_mon_check_error",
		Help: "Last check error code",
	}, []string{"host", "address"})

	registry := prometheus.NewRegistry()
	registry.MustRegister(srv.elapsedDays)
	registry.MustRegister(srv.checkError)

	metricsHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", metricsHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &srv
}

func (s *metricsServer) startHTTP() {
	log.Printf("HTTP server starting at: %s\n", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Unable to start HTTP server: %v\n", err)
	}
}

func (s *metricsServer) stopHTTP() {
	log.Printf("Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.httpServer.SetKeepAlivesEnabled(false)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown HTTP server: %v\n", err)
	}
}

func (s *metricsServer) loadConfig() error {
	domains, err := config.Parse(*s.options.ConfigFile)

	if err == nil {
		s.domains = *domains
		log.Printf("Loaded %d hosts from: %s", cap(s.domains), *s.options.ConfigFile)
	}

	return err
}

func (s *metricsServer) startWorker(interval int) error {
	if err := s.loadConfig(); err != nil {
		return err
	}

	go func() {
		for {
			counter := 0

			// Reset counters because this is simpliest way to remove unused domains or addresses
			s.elapsedDays.Reset()
			s.checkError.Reset()

			for elem, domain := range s.domains {
				domain.Resolve(*s.options.CheckIPv6)

				for _, addr := range domain.Addresses {
					var check *certexp.Check = certexp.NewCheck(certexp.HostInfo{Name: domain.Name, Address: addr, Port: domain.Port})

					if err := check.Expiration(); err != nil {
						log.Printf("Expiration check error: %v", err)
					}

					for _, res := range check.Result {
						s.elapsedDays.WithLabelValues(check.Host.Name, res.Address.String()).Set(float64(res.Expiry.Days))
						s.checkError.WithLabelValues(check.Host.Name, res.Address.String()).Set(float64(res.Error.Code))
					}

					counter = elem + 1
				}
			}

			log.Printf("Processed: %d hosts, sleeping for %v minutes\n", counter, interval)
			time.Sleep(time.Minute * time.Duration(interval))
		}
	}()

	return nil
}

func signalHandler(signalChan chan os.Signal, server *metricsServer) {
	for {
		sig := <-signalChan
		log.Printf("Handled signal: %s", sig)

		switch sig {
		case syscall.SIGHUP:
			if err := server.loadConfig(); err != nil {
				log.Printf("Not loaded new config because of unexpected error: %v", err)
			}
		default:
			server.stopHTTP()
		}
	}
}

func main() {
	options := config.NewOptions()
	options.CommonFlags()

	port := flag.Int("port", 2112, "Listen port")
	flag.Parse()

	signalChan := make(chan os.Signal, 1)

	program := filepath.Base(os.Args[0])

	log.Printf("Process %s started, version: %s+%s, author: %s\n", program, version, build, author)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	srv := newMetricsServer("0.0.0.0", *port, options)

	go signalHandler(signalChan, srv)

	if err := srv.startWorker(10); err != nil {
		log.Printf("Unexpected error: %v", err)
		log.Printf("Process %s stopped\n", program)
		os.Exit(1)
	}

	srv.startHTTP()

	log.Printf("Process %s stopped\n", program)
	os.Exit(0)
}
