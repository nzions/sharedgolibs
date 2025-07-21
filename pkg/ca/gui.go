package ca

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"
)

//go:embed gui/templates/*.html gui/static/*
var templateFS embed.FS

//go:embed gui/static/*
var staticFS embed.FS

// GUIHandler handles the web GUI requests
type GUIHandler struct {
	ca        *CA
	templates *template.Template
	apiKey    string
}

// CertificateViewModel represents a certificate for the GUI
type CertificateViewModel struct {
	*IssuedCert
	IsExpired      bool
	IsExpiringSoon bool
}

// DashboardData holds data for the dashboard template
type DashboardData struct {
	Title                string
	Page                 string
	Version              string
	CertCount            int
	CAValidUntil         string
	CAValidFrom          string
	CASerialNumber       string
	CASubject            string
	CAKeyAlgorithm       string
	CASignatureAlgorithm string
	RecentCerts          []CertificateViewModel
	AllCerts             []CertificateViewModel
	RequireAPIKey        bool
	BaseURL              string
}

// CertificatesData holds data for the certificates template
type CertificatesData struct {
	Title         string
	Page          string
	Version       string
	Certificates  []CertificateViewModel
	RequireAPIKey bool
	BaseURL       string
}

// GenerateData holds data for the generate template
type GenerateData struct {
	Title         string
	Page          string
	Version       string
	RequireAPIKey bool
	BaseURL       string
}

// APIData holds data for the API documentation template
type APIData struct {
	Title         string
	Page          string
	Version       string
	RequireAPIKey bool
	BaseURL       string
}

// NewGUIHandler creates a new GUI handler
func NewGUIHandler(ca *CA, apiKey string) (*GUIHandler, error) {
	tmpl, err := template.ParseFS(templateFS, "gui/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &GUIHandler{
		ca:        ca,
		templates: tmpl,
		apiKey:    apiKey,
	}, nil
}

// RequireAPIKey middleware to check API key if configured
func (g *GUIHandler) RequireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if g.apiKey == "" {
			// No API key required
			next(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Check for API key in query parameter
			apiKey = r.URL.Query().Get("api_key")
		}

		if apiKey != g.apiKey {
			http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// HandleDashboard renders the dashboard page
func (g *GUIHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	certs := g.ca.GetIssuedCertificates()
	allCerts := g.prepareCertificates(certs)
	recentCerts := allCerts

	// Show only the 5 most recent certificates
	if len(recentCerts) > 5 {
		recentCerts = recentCerts[:5]
	}

	caInfo := g.ca.GetCAInfo()
	caCert := g.ca.Certificate()

	// Determine base URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	data := DashboardData{
		Title:                "Dashboard",
		Page:                 "dashboard",
		Version:              Version,
		CertCount:            len(certs),
		CAValidUntil:         caCert.NotAfter.Format("2006-01-02"),
		CAValidFrom:          caCert.NotBefore.Format("2006-01-02"),
		CASerialNumber:       caCert.SerialNumber.String(),
		CASubject:            caCert.Subject.String(),
		CAKeyAlgorithm:       caCert.PublicKeyAlgorithm.String(),
		CASignatureAlgorithm: caCert.SignatureAlgorithm.String(),
		RecentCerts:          recentCerts,
		AllCerts:             allCerts,
		RequireAPIKey:        g.apiKey != "",
		BaseURL:              baseURL,
	}

	// Add additional CA info from the CA's GetCAInfo method
	for key, value := range caInfo {
		switch key {
		case "common_name":
			if str, ok := value.(string); ok {
				data.CASubject = str
			}
		}
	}

	if err := g.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleCertificates renders the certificates list page
func (g *GUIHandler) HandleCertificates(w http.ResponseWriter, r *http.Request) {
	certs := g.ca.GetIssuedCertificates()
	certificates := g.prepareCertificates(certs)

	// Determine base URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	data := CertificatesData{
		Title:         "Certificates",
		Page:          "certs",
		Version:       Version,
		Certificates:  certificates,
		RequireAPIKey: g.apiKey != "",
		BaseURL:       baseURL,
	}

	if err := g.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGenerate renders the generate certificate page or processes the form
func (g *GUIHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Determine base URL from request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

		data := GenerateData{
			Title:         "Generate Certificate",
			Page:          "generate",
			Version:       Version,
			RequireAPIKey: g.apiKey != "",
			BaseURL:       baseURL,
		}

		if err := g.templates.ExecuteTemplate(w, "base.html", data); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		g.handleGenerateForm(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleAPI renders the API documentation page
func (g *GUIHandler) HandleAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Determine base URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	data := APIData{
		Title:         "API Documentation",
		Page:          "api",
		Version:       Version,
		RequireAPIKey: g.apiKey != "",
		BaseURL:       baseURL,
	}

	if err := g.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleGenerateForm processes the certificate generation form
func (g *GUIHandler) handleGenerateForm(w http.ResponseWriter, r *http.Request) {
	serviceName := strings.TrimSpace(r.FormValue("service_name"))
	serviceIP := strings.TrimSpace(r.FormValue("service_ip"))
	domainsText := strings.TrimSpace(r.FormValue("domains"))

	if serviceName == "" || serviceIP == "" || domainsText == "" {
		g.writeHTMLResponse(w, `
			<div class="alert alert-error">
				<strong>Error:</strong> All fields are required.
			</div>
		`)
		return
	}

	// Parse domains from textarea (one per line)
	domains := []string{}
	for _, domain := range strings.Split(domainsText, "\n") {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	if len(domains) == 0 {
		g.writeHTMLResponse(w, `
			<div class="alert alert-error">
				<strong>Error:</strong> At least one domain is required.
			</div>
		`)
		return
	}

	// Generate certificate
	certPEM, keyPEM, err := g.ca.GenerateCertificate(serviceName, serviceIP, domains)
	if err != nil {
		g.writeHTMLResponse(w, fmt.Sprintf(`
			<div class="alert alert-error">
				<strong>Error:</strong> Failed to generate certificate: %v
			</div>
		`, err))
		return
	}

	// Success response with certificate details
	html := fmt.Sprintf(`
		<div class="alert alert-success">
			<strong>Success!</strong> Certificate generated successfully for <strong>%s</strong>.
		</div>
		
		<div class="card">
			<h3>📋 Certificate Details</h3>
			<table class="table">
				<tbody>
					<tr><td><strong>Service Name</strong></td><td>%s</td></tr>
					<tr><td><strong>Service IP</strong></td><td>%s</td></tr>
					<tr><td><strong>Domains</strong></td><td>%s</td></tr>
					<tr><td><strong>Certificate Size</strong></td><td>%d bytes</td></tr>
					<tr><td><strong>Private Key Size</strong></td><td>%d bytes</td></tr>
				</tbody>
			</table>
		</div>
		
		<div class="card">
			<h4>📄 Certificate (PEM)</h4>
			<textarea class="form-input" rows="8" readonly onclick="this.select()">%s</textarea>
			<p style="margin-top: 8px; color: #86868b; font-size: 14px;">
				Save this as <code>%s.crt</code>
			</p>
		</div>
		
		<div class="card">
			<h4>🔑 Private Key (PEM)</h4>
			<textarea class="form-input" rows="8" readonly onclick="this.select()">%s</textarea>
			<p style="margin-top: 8px; color: #86868b; font-size: 14px;">
				Save this as <code>%s.key</code> and keep it secure!
			</p>
		</div>
		
		<div style="margin-top: 16px;">
			<a href="/ui/certs" class="btn">📋 View All Certificates</a>
			<a href="/ui/generate" class="btn btn-success">🔐 Generate Another</a>
		</div>
	`,
		serviceName,
		serviceName, serviceIP, strings.Join(domains, ", "),
		len(certPEM), len(keyPEM),
		certPEM, serviceName,
		keyPEM, serviceName,
	)

	g.writeHTMLResponse(w, html)
}

// HandleCertDetails renders certificate details for the modal
func (g *GUIHandler) HandleCertDetails(w http.ResponseWriter, r *http.Request) {
	// Extract serial number from URL path
	path := strings.TrimPrefix(r.URL.Path, "/ui/cert-details/")
	serialNumber := strings.TrimSpace(path)

	if serialNumber == "" {
		http.Error(w, "Serial number required", http.StatusBadRequest)
		return
	}

	certs := g.ca.GetIssuedCertificates()
	var foundCert *IssuedCert
	for _, cert := range certs {
		if cert.SerialNumber == serialNumber {
			foundCert = cert
			break
		}
	}

	if foundCert == nil {
		g.writeHTMLResponse(w, `
			<div class="alert alert-error">
				<strong>Error:</strong> Certificate not found.
			</div>
		`)
		return
	}

	certVM := g.prepareCertificate(foundCert)
	if err := g.templates.ExecuteTemplate(w, "cert-details", certVM); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// prepareCertificates converts IssuedCert slice to CertificateViewModel slice
func (g *GUIHandler) prepareCertificates(certs []*IssuedCert) []CertificateViewModel {
	result := make([]CertificateViewModel, 0, len(certs))

	for _, cert := range certs {
		vm := g.prepareCertificate(cert)
		result = append(result, vm)
	}

	// Sort by issued date (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].IssuedAt.After(result[j].IssuedAt)
	})

	return result
}

// prepareCertificate converts IssuedCert to CertificateViewModel
func (g *GUIHandler) prepareCertificate(cert *IssuedCert) CertificateViewModel {
	now := time.Now()
	expiringThreshold := 7 * 24 * time.Hour // 7 days

	return CertificateViewModel{
		IssuedCert:     cert,
		IsExpired:      now.After(cert.ExpiresAt),
		IsExpiringSoon: !now.After(cert.ExpiresAt) && cert.ExpiresAt.Sub(now) < expiringThreshold,
	}
}

// writeHTMLResponse writes an HTML response
func (g *GUIHandler) writeHTMLResponse(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleLegacyWebUI handles the legacy web UI (old inline HTML template)
func (g *GUIHandler) HandleLegacyWebUI(w http.ResponseWriter, r *http.Request) {
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

	certCount := g.ca.GetCertificateCount()

	data := struct {
		CertCount    int
		CAValidUntil string
	}{
		CertCount:    certCount,
		CAValidUntil: g.ca.Certificate().NotAfter.Format("Jan 2, 2006"),
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

// HandleLegacyCertsList handles the legacy certificates list (old inline HTML template)
func (g *GUIHandler) HandleLegacyCertsList(w http.ResponseWriter, r *http.Request) {
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
	certList := g.ca.GetIssuedCertificates()

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

// HandleLegacyGenerateForm handles the legacy generate form (old inline HTML template)
func (g *GUIHandler) HandleLegacyGenerateForm(w http.ResponseWriter, r *http.Request) {
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
		certPEM, keyPEM, err := g.ca.GenerateCertificate(serviceName, "internal", domains)
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
		caCertPEM := g.ca.CertificatePEM()

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

// HandleDownloadCA handles CA certificate download requests
func (g *GUIHandler) HandleDownloadCA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caCertPEM := g.ca.CertificatePEM()

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=sharedgolibs-ca.crt")
	w.Write(caCertPEM)
}

// HandleDownloadCAKey handles CA private key download requests
func (g *GUIHandler) HandleDownloadCAKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	caKeyPEM := g.ca.PrivateKeyPEM()

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=sharedgolibs-ca.key")
	w.Write(caKeyPEM)
}

// HandleDownloadCert handles individual certificate download requests
func (g *GUIHandler) HandleDownloadCert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract serial number from URL path
	path := strings.TrimPrefix(r.URL.Path, "/cert/")
	serialNumber := strings.Split(path, "/")[0]

	if serialNumber == "" {
		http.Error(w, "Serial number required", http.StatusBadRequest)
		return
	}

	certs := g.ca.GetIssuedCertificates()
	var foundCert *IssuedCert
	for _, cert := range certs {
		if cert.SerialNumber == serialNumber {
			foundCert = cert
			break
		}
	}

	if foundCert == nil {
		http.Error(w, "Certificate not found", http.StatusNotFound)
		return
	}

	filename := fmt.Sprintf("%s.crt", foundCert.ServiceName)
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte(foundCert.Certificate))
}

// HandleDownloadCertKey handles individual certificate private key download requests
func (g *GUIHandler) HandleDownloadCertKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract serial number from URL path
	path := strings.TrimPrefix(r.URL.Path, "/cert/")
	serialNumber := strings.Split(path, "/")[0]

	if serialNumber == "" {
		http.Error(w, "Serial number required", http.StatusBadRequest)
		return
	}

	certs := g.ca.GetIssuedCertificates()
	var foundCert *IssuedCert
	for _, cert := range certs {
		if cert.SerialNumber == serialNumber {
			foundCert = cert
			break
		}
	}

	if foundCert == nil {
		http.Error(w, "Certificate not found", http.StatusNotFound)
		return
	}

	filename := fmt.Sprintf("%s.key", foundCert.ServiceName)
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte(foundCert.PrivateKey))
}

// HandleCertsTable handles HTMX requests for the certificates table
func (g *GUIHandler) HandleCertsTable(w http.ResponseWriter, r *http.Request) {
	certs := g.ca.GetIssuedCertificates()
	certificates := g.prepareCertificates(certs)

	// Generate table HTML
	html := `<table class="table" hx-get="/ui/certs-table" hx-trigger="every 30s" hx-swap="outerHTML">
		<thead>
			<tr>
				<th>SERVICE</th>
				<th>COMMON NAME</th>
				<th>SUBJECT ALT NAMES</th>
				<th>SERIAL</th>
				<th>ISSUED</th>
				<th>EXPIRES</th>
				<th>STATUS</th>
				<th>ACTIONS</th>
			</tr>
		</thead>
		<tbody>`

	for _, cert := range certificates {
		statusClass := "badge-success"
		statusText := "VALID"
		if cert.IsExpired {
			statusClass = "badge-danger"
			statusText = "EXPIRED"
		} else if cert.IsExpiringSoon {
			statusClass = "badge-warning"
			statusText = "EXPIRING"
		}

		domainsHTML := ""
		if len(cert.Domains) > 1 {
			domainsHTML = fmt.Sprintf(`<details><summary>%d domains</summary>`, len(cert.Domains))
			for _, domain := range cert.Domains {
				domainsHTML += fmt.Sprintf(`<div><code>%s</code></div>`, domain)
			}
			domainsHTML += `</details>`
		} else if len(cert.Domains) > 0 {
			domainsHTML = fmt.Sprintf(`<code>%s</code>`, cert.Domains[0])
		}

		html += fmt.Sprintf(`
			<tr>
				<td><strong>%s</strong></td>
				<td><code>%s</code></td>
				<td>%s</td>
				<td><code>%s</code></td>
				<td>%s</td>
				<td>%s</td>
				<td><span class="badge %s">%s</span></td>
				<td>
					<div class="download-links">
						<a href="/cert/%s" class="btn" onclick="downloadFile('/cert/%s', '%s.crt')" title="Download certificate">CERT</a>
						<a href="/cert/%s/key" class="btn btn-danger" onclick="downloadFile('/cert/%s/key', '%s.key')" title="Download private key">KEY</a>
					</div>
				</td>
			</tr>`,
			cert.ServiceName,
			cert.Domains[0],
			domainsHTML,
			cert.SerialNumber,
			cert.IssuedAt.Format("01-02 15:04"),
			cert.ExpiresAt.Format("01-02 15:04"),
			statusClass, statusText,
			cert.SerialNumber, cert.SerialNumber, cert.ServiceName,
			cert.SerialNumber, cert.SerialNumber, cert.ServiceName,
		)
	}

	html += `</tbody></table>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleLogStream handles Server-Sent Events for live log streaming
func (g *GUIHandler) HandleLogStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial message
	fmt.Fprintf(w, "data: CA System initialized\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Create a channel to receive log messages
	logChan := make(chan string, 100)

	// TODO: Connect to actual log stream from CA
	// For now, send periodic status updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			certs := g.ca.GetIssuedCertificates()
			message := fmt.Sprintf("System status: %d certificates active", len(certs))
			fmt.Fprintf(w, "data: %s\n\n", message)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		case msg := <-logChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// HandleStatic serves static files (JS, CSS, fonts) from embedded filesystem
func (g *GUIHandler) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /ui/static/ prefix to get the file path within the static directory
	path := strings.TrimPrefix(r.URL.Path, "/ui/static/")
	if path == "" || path == "/" {
		http.NotFound(w, r)
		return
	}

	// Construct the full path within the embedded filesystem
	fullPath := "gui/static/" + path

	// Read the file from embedded filesystem
	data, err := staticFS.ReadFile(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set appropriate content type based on file extension
	var contentType string
	switch {
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".woff2"):
		contentType = "font/woff2"
	case strings.HasSuffix(path, ".woff"):
		contentType = "font/woff"
	case strings.HasSuffix(path, ".ttf"):
		contentType = "font/ttf"
	default:
		contentType = "application/octet-stream"
	}

	// Set headers for caching
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Write the file content
	w.Write(data)
}
