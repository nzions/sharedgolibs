# GoogleEmu Web UI Component Library

A modern, themeable component library for building consistent frontend interfaces across GoogleEmu services.

## Features

🎨 **Multiple Themes**
- **Glassmorphism**: Futuristic glassmorphism design (inspired by current portdash)
- **Professional**: Clean corporate interface  
- **Hacker**: Terminal-inspired dark theme
- **Puppies**: Fun and colorful design

🧩 **Reusable Components**
- Service cards with status indicators
- Page headers with branding
- Buttons with consistent styling
- Responsive grid layouts
- Auto-refresh functionality

⚡ **Modern Web Standards**
- CSS custom properties for theming
- Web Components (Custom Elements)
- Responsive design
- Smooth animations and transitions

## Quick Start

### 1. Include CSS and JavaScript

```html
<!DOCTYPE html>
<html lang="en" data-theme="glassmorphism">
<head>
    <link rel="stylesheet" href="/static/css/googleemu.css">
</head>
<body>
    <!-- Your content -->
    <script src="/static/js/googleemu.js"></script>
</body>
</html>
```

### 2. Use Components in HTML

```html
<!-- Page Header -->
<ge-header 
    title="Gmail Emulator" 
    subtitle="View and manage sent emails"
    version="1.2.0">
</ge-header>

<!-- Service Card -->
<ge-service-card
    name="Gmail Emulator"
    status="running"
    port="8086"
    url="https://localhost:8086"
    description="Gmail API emulator for testing"
    uptime="2h 15m">
</ge-service-card>

<!-- Button -->
<ge-button 
    text="Refresh" 
    type="primary" 
    onclick="location.reload()">
</ge-button>
```

### 3. Use Go Helper Functions

```go
package main

import "github.com/nzions/sharedgolibs/pkg/webui"

func main() {
    // Initialize component renderer
    renderer, err := webui.NewComponentRenderer("static", true)
    if err != nil {
        log.Fatal(err)
    }

    // Create service data
    services := []webui.ServiceData{
        {
            Name:        "Gmail Emulator",
            Status:      "running",
            Port:        "8086", 
            URL:         "https://localhost:8086",
            Description: "Gmail API emulator",
            Uptime:      "2h 15m",
        },
    }

    // Build complete page
    pageData := webui.BuildDashboardPage(
        "Gmail Emulator",
        "View and manage sent emails", 
        "1.2.0",
        services,
        true, // Auto-refresh enabled
    )

    // Render to HTTP response
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        renderer.RenderPage(w, pageData)
    })
}
```

## Themes

Switch themes by setting the `data-theme` attribute:

```html
<html data-theme="glassmorphism"> <!-- 🌌 Glassmorphism -->
<html data-theme="professional"> <!-- 💼 Corporate -->
<html data-theme="hacker">       <!-- 🖥️ Terminal -->
<html data-theme="puppies">      <!-- 🐕 Colorful -->
```

Or use JavaScript:

```javascript
// Auto-loads saved theme preference
const ge = new GoogleEmuComponents();

// Change theme programmatically
ge.applyTheme('hacker');
```

## Component Reference

### Service Card

Display service status with visual indicators:

```javascript
// JavaScript helper
GoogleEmuComponents.createServiceCard({
    name: "Firebase Emulator",
    status: "running", // "running", "stopped", "warning"
    port: "8083",
    url: "http://localhost:8083", 
    description: "Firebase Auth & Firestore emulator",
    uptime: "1d 4h 23m"
});
```

```go
// Go helper
webui.CreateServiceCard(webui.ServiceData{
    Name:        "Firebase Emulator",
    Status:      "running",
    Port:        "8083",
    URL:         "http://localhost:8083",
    Description: "Firebase Auth & Firestore emulator", 
    Uptime:      "1d 4h 23m",
})
```

### Header

Page headers with optional subtitle and version:

```javascript
GoogleEmuComponents.createHeader(
    "GoogleEmu Dashboard",
    "Monitor all emulator services", 
    "1.0.0"
);
```

### Button

Consistent button styling:

```javascript
GoogleEmuComponents.createButton(
    "Refresh",        // text
    "primary",        // type: primary, secondary, success, warning, danger
    "sm",            // size: sm, (default), lg
    "location.reload()", // onclick
    ""               // href (for links)
);
```

### Grid Layout

Responsive grid for multiple items:

```javascript
const serviceCards = [
    GoogleEmuComponents.createServiceCard(service1),
    GoogleEmuComponents.createServiceCard(service2),
];

GoogleEmuComponents.createGrid(serviceCards, "auto"); // auto, 2, 3, 4
```

## Auto-Refresh

Enable automatic page refresh:

```javascript
const ge = new GoogleEmuComponents();

// Refresh every 30 seconds with visual countdown
ge.startAutoRefresh(30);

// Custom refresh callback
ge.startAutoRefresh(15, () => {
    // Custom refresh logic
    fetchUpdatedData();
});

// Stop auto-refresh
ge.stopAutoRefresh();
```

## Notifications

Show temporary notifications:

```javascript
GoogleEmuComponents.showNotification(
    "Service started successfully!", 
    "success",  // info, success, warning, danger
    5000       // duration in ms
);
```

## CSS Classes

### Layout

- `.ge-container` - Main content container (max-width: 1200px)
- `.ge-container-sm` - Small container (max-width: 640px) 
- `.ge-container-full` - Full width container
- `.ge-card` - Glass morphism card component
- `.ge-grid`, `.ge-grid-auto`, `.ge-grid-2` - Grid layouts

### Components

- `.ge-btn`, `.ge-btn-primary`, `.ge-btn-secondary` - Buttons
- `.ge-status`, `.ge-status-running`, `.ge-status-stopped` - Status badges
- `.ge-status-dot` - Minimal status indicators
- `.ge-service-card` - Service information cards
- `.ge-header` - Page headers

### Utilities

- `.ge-text-center`, `.ge-text-left`, `.ge-text-right` - Text alignment
- `.ge-mb-{size}`, `.ge-mt-{size}` - Margins (xs, sm, md, lg, xl)
- `.ge-p-{size}` - Padding
- `.ge-flex`, `.ge-items-center`, `.ge-justify-between` - Flexbox
- `.ge-fade-in`, `.ge-slide-up`, `.ge-pulse` - Animations

## File Structure

```
pkg/webui/
├── static/
│   ├── css/
│   │   └── googleemu.css      # Main stylesheet with themes
│   ├── js/
│   │   └── googleemu.js       # Component library JavaScript
│   └── components/
│       ├── base.html          # Complete page template
│       ├── header.html        # Header component template
│       └── service-card.html  # Service card template
├── components.go              # Go helper functions
└── README.md                  # This file
```

## Migration Guide

### From Embedded HTML Strings

**Before:**
```go
html := `
<div style="background: rgba(255,255,255,0.1); border-radius: 12px;">
    <h3>` + serviceName + `</h3>
    <span>` + status + `</span>
</div>
`
```

**After:**
```go
import "github.com/nzions/sharedgolibs/pkg/webui"

html := webui.CreateServiceCard(webui.ServiceData{
    Name:   serviceName,
    Status: status,
    Port:   port,
})
```

### From Custom CSS

Replace custom styling with theme-aware classes:

**Before:**
```html
<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
```

**After:**
```html
<div class="ge-card">
```

## Examples

Complete examples are available in:
- [`examples/portdash-modernized.go`](../examples/portdash-modernized.go) - Modern port dashboard
- Integration examples for Gmail, GCS, and other emulators

## Version

Current version: **1.0.0**

## License

Part of the GoogleEmu project.
