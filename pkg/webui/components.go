package webui

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

const version = "1.5.0"

// PageData represents data for rendering a complete page
type PageData struct {
	Title           string
	Subtitle        string
	Version         string
	Content         string
	CustomCSS       string
	CustomJS        string
	AutoRefresh     bool
	RefreshInterval int
	Theme           string
}

// ServiceData represents a service for the service card component
type ServiceData struct {
	Name        string
	Status      string // "running", "stopped", "warning"
	Port        string
	URL         string
	Description string
	Uptime      string
}

// ComponentRenderer handles rendering of web components
type ComponentRenderer struct {
	baseTemplate    *template.Template
	staticPath      string
	enableHotReload bool
}

// NewComponentRenderer creates a new component renderer
func NewComponentRenderer(staticPath string, enableHotReload bool) (*ComponentRenderer, error) {
	cr := &ComponentRenderer{
		staticPath:      staticPath,
		enableHotReload: enableHotReload,
	}

	if err := cr.loadTemplates(); err != nil {
		return nil, err
	}

	return cr, nil
}

// loadTemplates loads all component templates
func (cr *ComponentRenderer) loadTemplates() error {
	basePath := filepath.Join(cr.staticPath, "components", "base.html")
	tmpl, err := template.ParseFiles(basePath)
	if err != nil {
		return err
	}
	cr.baseTemplate = tmpl
	return nil
}

// RenderPage renders a complete page with the given data
func (cr *ComponentRenderer) RenderPage(w http.ResponseWriter, data PageData) error {
	if cr.enableHotReload {
		cr.loadTemplates() // Reload templates on each request for development
	}

	// Set defaults
	if data.Theme == "" {
		data.Theme = "glassmorphism"
	}
	if data.RefreshInterval == 0 {
		data.RefreshInterval = 30
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return cr.baseTemplate.Execute(w, data)
}

// CreateServiceCard creates HTML for a service card
func CreateServiceCard(service ServiceData) string {
	statusClass := getStatusClass(service.Status)

	var urlLink string
	if service.URL != "" {
		urlLink = `<a href="` + service.URL + `" target="_blank" class="ge-btn ge-btn-primary ge-btn-sm">Open</a>`
	}

	var description string
	if service.Description != "" {
		description = `<p class="description">` + service.Description + `</p>`
	}

	var uptime string
	if service.Uptime != "" {
		uptime = `<div class="uptime">Uptime: ` + service.Uptime + `</div>`
	}

	return `
		<div class="ge-service-card ge-fade-in">
			<h3>
				<span class="ge-status-dot ge-status-dot-` + statusClass + `"></span>
				` + service.Name + `
			</h3>
			<div class="service-info">
				<span class="ge-status ge-status-` + statusClass + `">` + service.Status + `</span>
				<span class="port">Port: ` + service.Port + `</span>
				` + urlLink + `
			</div>
			` + description + `
			` + uptime + `
		</div>
	`
}

// CreateHeader creates HTML for a page header
func CreateHeader(title, subtitle, version string) string {
	var subtitleHTML, versionHTML string

	if subtitle != "" {
		subtitleHTML = `<div class="subtitle">` + subtitle + `</div>`
	}

	if version != "" {
		versionHTML = `<div class="version">v` + version + `</div>`
	}

	return `
		<div class="ge-header ge-fade-in">
			<h1>` + title + `</h1>
			` + subtitleHTML + `
			` + versionHTML + `
		</div>
	`
}

// CreateButton creates HTML for a button
func CreateButton(text, buttonType, size, onclick, href string) string {
	sizeClass := ""
	if size != "" {
		sizeClass = "ge-btn-" + size
	}

	clickHandler := ""
	if onclick != "" {
		clickHandler = `onclick="` + onclick + `"`
	}

	if href != "" {
		return `<a href="` + href + `" class="ge-btn ge-btn-` + buttonType + ` ` + sizeClass + `">` + text + `</a>`
	}

	return `<button class="ge-btn ge-btn-` + buttonType + ` ` + sizeClass + `" ` + clickHandler + `>` + text + `</button>`
}

// CreateGrid creates HTML for a responsive grid
func CreateGrid(items []string, columns string) string {
	gridClass := "ge-grid-auto"
	if columns != "auto" {
		gridClass = "ge-grid-" + columns
	}

	itemsHTML := strings.Join(items, "")
	return `<div class="ge-grid ` + gridClass + `">` + itemsHTML + `</div>`
}

// CreateServiceGrid creates a grid of service cards
func CreateServiceGrid(services []ServiceData, columns string) string {
	var cards []string
	for _, service := range services {
		cards = append(cards, CreateServiceCard(service))
	}
	return CreateGrid(cards, columns)
}

// StaticFileHandler serves static assets with proper caching headers
func (cr *ComponentRenderer) StaticFileHandler() http.Handler {
	fs := http.FileServer(http.Dir(cr.staticPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set caching headers for static assets
		if !cr.enableHotReload {
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
				w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "application/javascript")
				w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
			}
		} else {
			// Disable caching in development
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		fs.ServeHTTP(w, r)
	})
}

// Helper functions
func getStatusClass(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return "running"
	case "stopped":
		return "stopped"
	case "warning":
		return "warning"
	default:
		return "warning"
	}
}

// Theme helpers
func GetAvailableThemes() map[string]string {
	return map[string]string{
		"glassmorphism": "🌌 Glassmorphism",
		"professional":  "💼 Professional",
		"hacker":        "🖥️ Hacker",
		"puppies":       "🐕 Puppies",
		"weyland":       "🛸 Weyland-Yutani",
		"line-minimum":  "📝 Line Minimum",
		"waifu":         "🌸 Waifu/UwU",
	}
}

// BuildDashboardPage builds a complete dashboard page with services
func BuildDashboardPage(title, subtitle, pageVersion string, services []ServiceData, autoRefresh bool) PageData {
	content := CreateServiceGrid(services, "auto")

	return PageData{
		Title:           title,
		Subtitle:        subtitle,
		Version:         pageVersion,
		Content:         content,
		AutoRefresh:     autoRefresh,
		RefreshInterval: 30,
		Theme:           "glassmorphism",
	}
}
