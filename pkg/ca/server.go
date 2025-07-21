// Package ca provides HTTP server functionality for the Certificate Authority
package ca

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/nzions/sharedgolibs/pkg/middleware"
)

// Server wraps the CA with HTTP server functionality
type Server struct {
	ca        *CA
	port      string
	enableGUI bool
	guiAPIKey string
	gui       *GUIHandler
}

// ServerConfig holds configuration for the CA server
type ServerConfig struct {
	Port       string
	CAConfig   *CAConfig
	EnableGUI  bool   // Enable the web GUI interface
	GUIAPIKey  string // API key required for GUI access (if set)
	PersistDir string // Directory to persist CA data (empty = RAM only)
}

// DefaultServerConfig returns sensible defaults for server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:       "8090",
		CAConfig:   DefaultCAConfig(),
		EnableGUI:  true, // GUI enabled by default
		GUIAPIKey:  "",   // No API key by default
		PersistDir: "",   // RAM only by default
	}
}

// NewServer creates a new CA server
func NewServer(config *ServerConfig) (*Server, error) {
	if config == nil {
		config = DefaultServerConfig()
	}

	// Pass PersistDir from ServerConfig to CAConfig
	if config.CAConfig != nil {
		config.CAConfig.PersistDir = config.PersistDir
	} else {
		config.CAConfig = DefaultCAConfig()
		config.CAConfig.PersistDir = config.PersistDir
	}

	ca, err := NewCA(config.CAConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA: %w", err)
	}

	server := &Server{
		ca:        ca,
		port:      config.Port,
		enableGUI: config.EnableGUI,
		guiAPIKey: config.GUIAPIKey,
	}

	// Initialize GUI handler if enabled
	if config.EnableGUI {
		gui, err := NewGUIHandler(ca, config.GUIAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create GUI handler: %w", err)
		}
		server.gui = gui
	}

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up HTTP handlers with API key protection if configured
	var caHandler, certHandler, healthHandler http.Handler
	caHandler = http.HandlerFunc(s.handleCARequest)
	certHandler = http.HandlerFunc(s.handleCertRequest)
	healthHandler = http.HandlerFunc(s.handleHealth)

	// Apply API key middleware to API endpoints if API key is configured
	if s.guiAPIKey != "" {
		caHandler = middleware.WithAPIKey(s.guiAPIKey, caHandler)
		certHandler = middleware.WithAPIKey(s.guiAPIKey, certHandler)
		healthHandler = middleware.WithAPIKey(s.guiAPIKey, healthHandler)
	}

	http.Handle("/ca", caHandler)
	http.Handle("/cert", certHandler)
	http.Handle("/health", healthHandler)

	// Web UI handlers (only if GUI is enabled)
	if s.enableGUI && s.gui != nil {
		// Apply API key middleware if configured
		var dashboardHandler, certsHandler, generateHandler, certDetailsHandler, downloadCAHandler, downloadCAKeyHandler, downloadCertHandler, downloadCertKeyHandler, certsTableHandler, logStreamHandler http.Handler
		dashboardHandler = http.HandlerFunc(s.gui.HandleDashboard)
		certsHandler = http.HandlerFunc(s.gui.HandleCertificates)
		generateHandler = http.HandlerFunc(s.gui.HandleGenerate)
		certDetailsHandler = http.HandlerFunc(s.gui.HandleCertDetails)
		downloadCAHandler = http.HandlerFunc(s.gui.HandleDownloadCA)
		downloadCAKeyHandler = http.HandlerFunc(s.gui.HandleDownloadCAKey)
		downloadCertHandler = http.HandlerFunc(s.gui.HandleDownloadCert)
		downloadCertKeyHandler = http.HandlerFunc(s.gui.HandleDownloadCertKey)
		certsTableHandler = http.HandlerFunc(s.gui.HandleCertsTable)
		logStreamHandler = http.HandlerFunc(s.gui.HandleLogStream)

		if s.guiAPIKey != "" {
			dashboardHandler = middleware.WithAPIKey(s.guiAPIKey, dashboardHandler)
			certsHandler = middleware.WithAPIKey(s.guiAPIKey, certsHandler)
			generateHandler = middleware.WithAPIKey(s.guiAPIKey, generateHandler)
			certDetailsHandler = middleware.WithAPIKey(s.guiAPIKey, certDetailsHandler)
			downloadCAHandler = middleware.WithAPIKey(s.guiAPIKey, downloadCAHandler)
			downloadCAKeyHandler = middleware.WithAPIKey(s.guiAPIKey, downloadCAKeyHandler)
			downloadCertHandler = middleware.WithAPIKey(s.guiAPIKey, downloadCertHandler)
			downloadCertKeyHandler = middleware.WithAPIKey(s.guiAPIKey, downloadCertKeyHandler)
			certsTableHandler = middleware.WithAPIKey(s.guiAPIKey, certsTableHandler)
			logStreamHandler = middleware.WithAPIKey(s.guiAPIKey, logStreamHandler)
		}

		http.Handle("/", dashboardHandler)
		http.Handle("/ui/", dashboardHandler)
		http.Handle("/ui/certs", certsHandler)
		http.Handle("/ui/generate", generateHandler)
		http.Handle("/ui/cert-details/", certDetailsHandler)
		http.Handle("/ui/download-ca", downloadCAHandler)
		http.Handle("/ca-key", downloadCAKeyHandler)
		http.Handle("/ui/certs-table", certsTableHandler)
		http.Handle("/ui/logs", logStreamHandler)

		// Certificate download routes - these need special handling
		http.HandleFunc("/cert/", func(w http.ResponseWriter, r *http.Request) {
			if s.guiAPIKey != "" {
				// Check API key
				apiKey := r.Header.Get("X-API-Key")
				if apiKey == "" {
					apiKey = r.URL.Query().Get("api_key")
				}
				if apiKey != s.guiAPIKey {
					http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
					return
				}
			}

			// Check if it's a key request
			if strings.HasSuffix(r.URL.Path, "/key") {
				s.gui.HandleDownloadCertKey(w, r)
			} else {
				s.gui.HandleDownloadCert(w, r)
			}
		})
	}

	log.Printf("[ca] Certificate Authority listening on port %s", s.port)
	log.Printf("[ca] Endpoints:")
	log.Printf("[ca]   GET  /ca    - Download CA certificate")
	log.Printf("[ca]   POST /cert  - Request service certificate")
	log.Printf("[ca]   GET  /health - Health check")

	if s.guiAPIKey != "" {
		log.Printf("[ca]   Note: All endpoints require API key authentication")
		log.Printf("[ca]   Use X-API-Key header or ?api_key= query parameter")
	}

	if s.enableGUI {
		log.Printf("[ca]   GUI Interface:")
		log.Printf("[ca]     GET  /ui/   - Web UI dashboard")
		log.Printf("[ca]     GET  /ui/certs - List issued certificates")
		log.Printf("[ca]     GET  /ui/generate - Generate new certificate")
	} else {
		log.Printf("[ca]   GUI interface is disabled")
	}

	return http.ListenAndServe(":"+s.port, nil)
}

// GetCA returns the underlying CA instance
func (s *Server) GetCA() *CA {
	return s.ca
}

func (s *Server) handleCARequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[ca] CA certificate requested from %s", r.RemoteAddr)

	caCertPEM := s.ca.CertificatePEM()

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=sharedgolibs-ca.crt")
	w.Write(caCertPEM)
}

func (s *Server) handleCertRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ca] Invalid certificate request from %s: %v", r.RemoteAddr, err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate the request using the same validation as IssueServiceCertificate
	if req.ServiceName == "" {
		log.Printf("[ca] Invalid certificate request from %s: missing service_name", r.RemoteAddr)
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}

	if len(req.Domains) == 0 {
		log.Printf("[ca] Invalid certificate request from %s: missing domains", r.RemoteAddr)
		http.Error(w, "domains are required", http.StatusBadRequest)
		return
	}

	log.Printf("[ca] Certificate request from %s for service: %s, IP: %s, domains: %v", r.RemoteAddr, req.ServiceName, req.ServiceIP, req.Domains)

	// Issue certificate using the CA
	response, err := s.ca.IssueServiceCertificate(req)
	if err != nil {
		log.Printf("[ca] Failed to generate certificate for %s: %v", req.ServiceName, err)
		http.Error(w, "Certificate generation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[ca] ✅ Certificate issued for %s", req.ServiceName)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caInfo := s.ca.GetCAInfo()
	response := map[string]interface{}{
		"status":  "healthy",
		"version": Version,
		"ca_info": caInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
