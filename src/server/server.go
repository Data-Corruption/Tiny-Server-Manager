package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tsm/src/files"
	"tsm/src/server/middleware"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
)

const (
	ShutdownTimeout = 30 * time.Second
)

var Instance Server // Global server instance

// all members are set at init and never changed
type Server struct {
	// "{protocol}://{host}{port}" or "{protocol}://{host}" if port is left to default (":80" or ":443")
	BaseURL        string
	Protocol       string
	Host           string
	Port           string
	LocalIP        string
	UsingTLS       bool
	ShutdownRoute  chan bool      // Channel to trigger a shutdown from the route
	ShutdownSignal chan os.Signal // Channel to listen for OS signals
	router         *chi.Mux
	server         *http.Server // The http or https server
}

// Returns the first non-loopback ipv4 address
func getLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		} // interface down
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		} // loopback interface

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("cannot find local IP address")
}

func (s *Server) initInfo() {
	// Check if using TLS
	s.Protocol = "http"
	if files.FilesExist([]string{files.Config.TLSKeyPath, files.Config.TLSCertPath}) {
		s.UsingTLS = true
		s.Protocol = "https"
		blog.Info("Using TLS")
	} else {
		blog.Warn("TLS files not found, using HTTP")
	}

	// Get the local IP
	var err error
	s.LocalIP, err = getLocalIP()
	if err != nil {
		panic(err)
	}
	middleware.LocalIP = s.LocalIP

	// Set the host
	if files.Config.Host == "" {
		s.Host = s.LocalIP
		if s.Host == "" {
			panic("could not determine host")
		}
	} else {
		s.Host = files.Config.Host
	}

	// Set the port
	if files.Config.Port == 0 {
		if s.UsingTLS {
			s.Port = ":443"
		} else {
			s.Port = ":80"
		}
	} else {
		s.Port = fmt.Sprintf(":%d", files.Config.Port)
	}

	// Set the base URL
	if s.Port == ":80" || s.Port == ":443" {
		s.BaseURL = fmt.Sprintf("%s://%s", s.Protocol, s.Host)
	} else {
		s.BaseURL = fmt.Sprintf("%s://%s%s", s.Protocol, s.Host, s.Port)
	}

	blog.Debug(fmt.Sprintf("Initialized server: %+v", s))
}

func (s *Server) Start() {
	// Initialize the server
	s.initInfo()
	s.ShutdownRoute = make(chan bool, 1)
	s.router = NewRouter()

	// Configure the server
	if s.UsingTLS {
		s.server = &http.Server{
			Addr:    s.Port,
			Handler: s.router,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
			},
		}
	} else {
		s.server = &http.Server{
			Addr:    s.Port,
			Handler: s.router,
		}
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for shutdown signals
	s.ShutdownSignal = make(chan os.Signal, 1)
	signal.Notify(s.ShutdownSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-s.ShutdownRoute:
			blog.Info("Shutdown initiated via route...")
		case <-s.ShutdownSignal:
			blog.Info("Shutdown initiated via OS signal...")
		}

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, ShutdownTimeout)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				panic("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := s.server.Shutdown(shutdownCtx)
		if err != nil {
			panic(err)
		}
		serverStopCtx()
	}()

	// Start the server
	blog.Info(fmt.Sprintf("Starting server on %s", s.BaseURL))
	log.Printf("Starting server on %s", s.BaseURL)
	if s.UsingTLS {
		if err := s.server.ListenAndServeTLS(files.Config.TLSCertPath, files.Config.TLSKeyPath); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failure in ListenAndServer: %v", err))
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failure in ListenAndServer: %v", err))
		}
	}

	// Wait for server to shutdown
	<-serverCtx.Done()
	blog.Info("Server shutdown gracefully")
}
