// Package ca provides HTTP server functionality for the Certificate Authority
package ca

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
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
	Port      string
	CAConfig  *CAConfig
	EnableGUI bool   // Enable the web GUI interface
	GUIAPIKey string // API key required for GUI access (if set)
}

// DefaultServerConfig returns sensible defaults for server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:      "8090",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: true, // GUI enabled by default
		GUIAPIKey: "",   // No API key by default
	}
}

// NewServer creates a new CA server
func NewServer(config *ServerConfig) (*Server, error) {
	if config == nil {
		config = DefaultServerConfig()
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
		var dashboardHandler, certsHandler, generateHandler, certDetailsHandler, downloadCAHandler http.Handler
		dashboardHandler = http.HandlerFunc(s.gui.HandleDashboard)
		certsHandler = http.HandlerFunc(s.gui.HandleCertificates)
		generateHandler = http.HandlerFunc(s.gui.HandleGenerate)
		certDetailsHandler = http.HandlerFunc(s.gui.HandleCertDetails)
		downloadCAHandler = http.HandlerFunc(s.handleDownloadCA)

		if s.guiAPIKey != "" {
			dashboardHandler = middleware.WithAPIKey(s.guiAPIKey, dashboardHandler)
			certsHandler = middleware.WithAPIKey(s.guiAPIKey, certsHandler)
			generateHandler = middleware.WithAPIKey(s.guiAPIKey, generateHandler)
			certDetailsHandler = middleware.WithAPIKey(s.guiAPIKey, certDetailsHandler)
			downloadCAHandler = middleware.WithAPIKey(s.guiAPIKey, downloadCAHandler)
		}

		http.Handle("/", dashboardHandler)
		http.Handle("/ui/", dashboardHandler)
		http.Handle("/ui/certs", certsHandler)
		http.Handle("/ui/generate", generateHandler)
		http.Handle("/ui/cert-details/", certDetailsHandler)
		http.Handle("/ui/download-ca", downloadCAHandler)
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

func (s *Server) handleWebUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>GoogleEmu Certificate Authority</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; background: #f5f5f7; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        h1 { color: #1d1d1f; margin-bottom: 8px; }
        .subtitle { color: #86868b; margin-bottom: 32px; }
        .card { border: 1px solid #d2d2d7; border-radius: 8px; padding: 24px; margin: 16px 0; }
        .button { background: #007aff; color: white; padding: 12px 24px; border: none; border-radius: 8px; text-decoration: none; display: inline-block; margin: 8px 8px 8px 0; }
        .button:hover { background: #0056cc; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin: 24px 0; }
        .stat { background: #f6f6f6; padding: 16px; border-radius: 8px; text-align: center; }
        .stat-value { font-size: 24px; font-weight: bold; color: #007aff; }
        .stat-label { color: #86868b; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>SharedGoLibs Certificate Authority</h1>
        <p class="subtitle">Dynamic SSL certificate issuance for development services</p>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">{{.CertCount}}</div>
                <div class="stat-label">Certificates Issued</div>
            </div>
            <div class="stat">
                <div class="stat-value">{{.CAValidUntil}}</div>
                <div class="stat-label">CA Valid Until</div>
            </div>
        </div>

        <div class="card">
            <h3>Quick Actions</h3>
            <a href="/ui/certs" class="button">📋 View All Certificates</a>
            <a href="/ui/generate" class="button">🔐 Generate New Certificate</a>
            <a href="/ui/download-ca" class="button">⬇️ Download Root CA</a>
        </div>

        <div class="card">
            <h3>API Endpoints</h3>
            <ul>
                <li><code>GET /ca</code> - Download CA certificate</li>
                <li><code>POST /cert</code> - Request service certificate</li>
                <li><code>GET /health</code> - Health check</li>
            </ul>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("dashboard").Parse(html)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	certCount := s.ca.GetCertificateCount()

	data := struct {
		CertCount    int
		CAValidUntil string
	}{
		CertCount:    certCount,
		CAValidUntil: s.ca.Certificate().NotAfter.Format("Jan 2, 2006"),
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (s *Server) handleCertsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Issued Certificates - SharedGoLibs CA</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; background: #f5f5f7; }
        .container { max-width: 1000px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        h1 { color: #1d1d1f; margin-bottom: 32px; }
        .back-link { color: #007aff; text-decoration: none; margin-bottom: 24px; display: inline-block; }
        .cert-table { width: 100%; border-collapse: collapse; margin: 24px 0; }
        .cert-table th, .cert-table td { padding: 12px; text-align: left; border-bottom: 1px solid #d2d2d7; }
        .cert-table th { background: #f6f6f6; font-weight: 600; }
        .domains { font-family: Monaco, monospace; font-size: 12px; color: #007aff; }
        .date { color: #86868b; font-size: 14px; }
        .serial { font-family: Monaco, monospace; font-size: 12px; }
        .empty { text-align: center; color: #86868b; padding: 40px; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/ui/" class="back-link">← Back to Dashboard</a>
        <h1>Issued Certificates</h1>
        
        {{if .Certs}}
        <table class="cert-table">
            <thead>
                <tr>
                    <th>Service Name</th>
                    <th>Domains</th>
                    <th>Issued</th>
                    <th>Expires</th>
                    <th>Serial Number</th>
                </tr>
            </thead>
            <tbody>
                {{range .Certs}}
                <tr>
                    <td><strong>{{.ServiceName}}</strong></td>
                    <td class="domains">{{.DomainsStr}}</td>
                    <td class="date">{{.IssuedAtStr}}</td>
                    <td class="date">{{.ExpiresAtStr}}</td>
                    <td class="serial">{{.SerialNumber}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{else}}
        <div class="empty">
            <p>No certificates have been issued yet.</p>
            <a href="/ui/generate" style="color: #007aff;">Generate your first certificate</a>
        </div>
        {{end}}
    </div>
</body>
</html>`

	tmpl, err := template.New("certs").Parse(html)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Get certificates and prepare for display
	certList := s.ca.GetIssuedCertificates()

	// Sort by issued date (newest first)
	sort.Slice(certList, func(i, j int) bool {
		return certList[i].IssuedAt.After(certList[j].IssuedAt)
	})

	var certs []struct {
		*IssuedCert
		DomainsStr   string
		IssuedAtStr  string
		ExpiresAtStr string
	}

	for _, cert := range certList {
		certs = append(certs, struct {
			*IssuedCert
			DomainsStr   string
			IssuedAtStr  string
			ExpiresAtStr string
		}{
			IssuedCert:   cert,
			DomainsStr:   strings.Join(cert.Domains, ", "),
			IssuedAtStr:  cert.IssuedAt.Format("Jan 2, 2006 15:04"),
			ExpiresAtStr: cert.ExpiresAt.Format("Jan 2, 2006"),
		})
	}

	data := struct {
		Certs []struct {
			*IssuedCert
			DomainsStr   string
			IssuedAtStr  string
			ExpiresAtStr string
		}
	}{
		Certs: certs,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (s *Server) handleGenerateForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Generate Certificate - SharedGoLibs CA</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; background: #f5f5f7; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        h1 { color: #1d1d1f; margin-bottom: 32px; }
        .back-link { color: #007aff; text-decoration: none; margin-bottom: 24px; display: inline-block; }
        .form-group { margin: 24px 0; }
        label { display: block; margin-bottom: 8px; font-weight: 600; color: #1d1d1f; }
        input, textarea { width: 100%; padding: 12px; border: 1px solid #d2d2d7; border-radius: 8px; font-size: 16px; }
        textarea { height: 120px; font-family: Monaco, monospace; }
        .button { background: #007aff; color: white; padding: 12px 24px; border: none; border-radius: 8px; cursor: pointer; font-size: 16px; }
        .button:hover { background: #0056cc; }
        .help { color: #86868b; font-size: 14px; margin-top: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/ui/" class="back-link">← Back to Dashboard</a>
        <h1>Generate New Certificate</h1>
        
        <form method="POST">
            <div class="form-group">
                <label for="service_name">Service Name</label>
                <input type="text" id="service_name" name="service_name" required placeholder="my-service">
                <div class="help">A friendly name for this service</div>
            </div>
            
            <div class="form-group">
                <label for="domains">Domain Names (SANs)</label>
                <textarea id="domains" name="domains" required placeholder="example.com&#10;api.example.com&#10;192.168.1.100&#10;localhost"></textarea>
                <div class="help">One domain/IP per line. Supports hostnames and IP addresses.</div>
            </div>
            
            <button type="submit" class="button">🔐 Generate Certificate</button>
        </form>
    </div>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	if r.Method == http.MethodPost {
		serviceName := r.FormValue("service_name")
		domainsText := r.FormValue("domains")

		if serviceName == "" || domainsText == "" {
			http.Error(w, "Service name and domains are required", http.StatusBadRequest)
			return
		}

		domains := []string{}
		for _, domain := range strings.Split(domainsText, "\n") {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}

		if len(domains) == 0 {
			http.Error(w, "At least one domain is required", http.StatusBadRequest)
			return
		}

		// Generate certificate
		certPEM, keyPEM, err := s.ca.GenerateCertificate(serviceName, "internal", domains)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate certificate: %v", err), http.StatusInternalServerError)
			return
		}

		html := `<!DOCTYPE html>
<html>
<head>
    <title>Certificate Generated - SharedGoLibs CA</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; background: #f5f5f7; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        h1 { color: #1d1d1f; margin-bottom: 32px; }
        .back-link { color: #007aff; text-decoration: none; margin-bottom: 24px; display: inline-block; }
        .success { background: #d4edda; color: #155724; padding: 16px; border-radius: 8px; margin: 24px 0; }
        .cert-section { margin: 24px 0; }
        .cert-section h3 { color: #1d1d1f; margin-bottom: 12px; }
        .cert-output { font-family: Monaco, monospace; font-size: 12px; background: #f6f6f6; padding: 16px; border-radius: 8px; white-space: pre-wrap; word-break: break-all; border: 1px solid #d2d2d7; }
        .button { background: #007aff; color: white; padding: 8px 16px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; margin: 8px 8px 8px 0; }
        .button:hover { background: #0056cc; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/ui/" class="back-link">← Back to Dashboard</a>
        <h1>Certificate Generated Successfully</h1>
        
        <div class="success">
            ✅ Certificate for <strong>{{.ServiceName}}</strong> has been generated with {{.DomainCount}} domain(s).
        </div>
        
        <div class="cert-section">
            <h3>Certificate (PEM format)</h3>
            <div class="cert-output">{{.Certificate}}</div>
            <button class="button" onclick="copyToClipboard('cert')">📋 Copy Certificate</button>
        </div>
        
        <div class="cert-section">
            <h3>Private Key (PEM format)</h3>
            <div class="cert-output">{{.PrivateKey}}</div>
            <button class="button" onclick="copyToClipboard('key')">📋 Copy Private Key</button>
        </div>
        
        <div class="cert-section">
            <h3>CA Certificate (PEM format)</h3>
            <div class="cert-output">{{.CACert}}</div>
            <button class="button" onclick="copyToClipboard('ca')">📋 Copy CA Certificate</button>
        </div>
    </div>
    
    <script>
        function copyToClipboard(type) {
            const elements = document.querySelectorAll('.cert-output');
            let text = '';
            if (type === 'cert') text = elements[0].textContent;
            else if (type === 'key') text = elements[1].textContent;
            else if (type === 'ca') text = elements[2].textContent;
            
            navigator.clipboard.writeText(text).then(() => {
                alert('Copied to clipboard!');
            });
        }
    </script>
</body>
</html>`

		tmpl, err := template.New("result").Parse(html)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		// Get CA cert PEM
		caCertPEM := s.ca.CertificatePEM()

		data := struct {
			ServiceName string
			DomainCount int
			Certificate string
			PrivateKey  string
			CACert      string
		}{
			ServiceName: serviceName,
			DomainCount: len(domains),
			Certificate: certPEM,
			PrivateKey:  keyPEM,
			CACert:      string(caCertPEM),
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, data)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleDownloadCA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caCertPEM := s.ca.CertificatePEM()

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=sharedgolibs-ca.crt")
	w.Write(caCertPEM)
}
