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

//go:embed gui/templates/*.html
var templateFS embed.FS

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
}

// CertificatesData holds data for the certificates template
type CertificatesData struct {
	Title        string
	Page         string
	Version      string
	Certificates []CertificateViewModel
}

// GenerateData holds data for the generate template
type GenerateData struct {
	Title   string
	Page    string
	Version string
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
	recentCerts := g.prepareCertificates(certs)

	// Show only the 5 most recent certificates
	if len(recentCerts) > 5 {
		recentCerts = recentCerts[:5]
	}

	caInfo := g.ca.GetCAInfo()
	caCert := g.ca.Certificate()

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

	data := CertificatesData{
		Title:        "Certificates",
		Page:         "certs",
		Version:      Version,
		Certificates: certificates,
	}

	if err := g.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGenerate renders the generate certificate page or processes the form
func (g *GUIHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := GenerateData{
			Title:   "Generate Certificate",
			Page:    "generate",
			Version: Version,
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
