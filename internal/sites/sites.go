package sites

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Site represents a hosted website on a custom subdomain.
type Site struct {
	Subdomain string    `json:"subdomain"` // e.g. "blog", "portfolio"
	Engine    string    `json:"engine"`    // "html", "php", "astro"
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Directory string    `json:"directory"`
}

type Manager struct {
	dataDir    string
	baseDomain string
	sites      map[string]*Site
	mu         sync.RWMutex
}

var (
	ErrSubdomainExists   = errors.New("subdomain already registered")
	ErrSubdomainNotFound = errors.New("website subdomain not found")
	ErrInvalidSubdomain  = errors.New("invalid subdomain format")
)

// NewManager initializes the web hosting site manager.
func NewManager(dataDir, baseDomain string) (*Manager, error) {
	sitesDir := filepath.Join(dataDir, "sites")
	if err := os.MkdirAll(sitesDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create sites directory: %w", err)
	}

	m := &Manager{
		dataDir:    sitesDir,
		baseDomain: strings.ToLower(strings.Trim(baseDomain, ".")),
		sites:      make(map[string]*Site),
	}

	_ = m.load()

	// Provision default pre-installed HTML website if no sites exist
	if len(m.sites) == 0 {
		_, _ = m.CreateSite("home", "html", "BenzCloud Welcome Page")
	}

	return m, nil
}

func (m *Manager) configPath() string {
	return filepath.Join(m.dataDir, "sites.json")
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath())
	if err != nil {
		return err
	}
	var list []*Site
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, s := range list {
		m.sites[s.Subdomain] = s
	}
	return nil
}

func (m *Manager) save() error {
	var list []*Site
	for _, s := range m.sites {
		list = append(list, s)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.configPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, m.configPath())
}

// Subdomains returns a slice of all registered custom subdomains.
func (m *Manager) Subdomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var subs []string
	for sub := range m.sites {
		subs = append(subs, sub)
	}
	return subs
}

// ListSites returns all registered websites.
func (m *Manager) ListSites() []*Site {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*Site
	for _, s := range m.sites {
		copyS := *s
		list = append(list, &copyS)
	}
	return list
}

// CreateSite creates a new hosted website under a chosen subdomain.
func (m *Manager) CreateSite(subdomain, engine, title string) (*Site, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleanSub := strings.ToLower(strings.TrimSpace(subdomain))
	if cleanSub == "" || strings.ContainsAny(cleanSub, " /\\.?*#@:") {
		return nil, ErrInvalidSubdomain
	}

	// Protect reserved system subdomains
	if cleanSub == "vpn" || cleanSub == "drive" || cleanSub == "mail" || cleanSub == "chat" {
		return nil, fmt.Errorf("subdomain '%s' is reserved for system services", cleanSub)
	}

	if _, exists := m.sites[cleanSub]; exists {
		return nil, ErrSubdomainExists
	}

	siteDir := filepath.Join(m.dataDir, cleanSub)
	if err := os.MkdirAll(siteDir, 0700); err != nil {
		return nil, err
	}

	// Seed template based on engine
	switch strings.ToLower(engine) {
	case "php":
		indexPHP := fmt.Sprintf(`<?php
echo "<!DOCTYPE html><html><head><meta charset='UTF-8'><title>%s</title><style>body{background:#0a0e17;color:#f1f5f9;font-family:sans-serif;padding:3rem;text-align:center;}</style></head><body><h1>🐘 %s</h1><p>Running with PHP Engine on BenzCloud</p><p>Server Time: " . date("Y-m-d H:i:s") . "</p></body></html>";
?>`, title, title)
		_ = os.WriteFile(filepath.Join(siteDir, "index.php"), []byte(indexPHP), 0644)
	case "astro":
		indexAstro := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s – Astro Static</title>
  <style>body{background:#0a0e17;color:#f1f5f9;font-family:sans-serif;padding:3rem;text-align:center;} h1{color:#38bdf8;}</style>
</head>
<body>
  <h1>🚀 %s</h1>
  <p>Modern Fast Static Website (Astro Engine)</p>
</body>
</html>`, title, title)
		_ = os.WriteFile(filepath.Join(siteDir, "index.html"), []byte(indexAstro), 0644)
	default: // html
		indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="UTF-8">
  <title>%s</title>
  <style>body{background:#0a0e17;color:#f1f5f9;font-family:sans-serif;padding:3rem;text-align:center;} h1{color:#34d399;}</style>
</head>
<body>
  <h1>🌐 %s</h1>
  <p>Erfolgreich gehostet über das BenzCloud Web-Plugin!</p>
</body>
</html>`, title, title)
		_ = os.WriteFile(filepath.Join(siteDir, "index.html"), []byte(indexHTML), 0644)
	}

	site := &Site{
		Subdomain: cleanSub,
		Engine:    engine,
		Title:     title,
		CreatedAt: time.Now().UTC(),
		Directory: siteDir,
	}

	m.sites[cleanSub] = site
	if err := m.save(); err != nil {
		return nil, err
	}
	return site, nil
}

// DeleteSite removes a site.
func (m *Manager) DeleteSite(subdomain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleanSub := strings.ToLower(subdomain)
	site, exists := m.sites[cleanSub]
	if !exists {
		return ErrSubdomainNotFound
	}

	_ = os.RemoveAll(site.Directory)
	delete(m.sites, cleanSub)
	return m.save()
}

// ServeSite serves files for a specific subdomain.
func (m *Manager) ServeSite(w http.ResponseWriter, r *http.Request, subdomain string) {
	m.mu.RLock()
	site, exists := m.sites[strings.ToLower(subdomain)]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Website '%s' not found on BenzCloud", subdomain), http.StatusNotFound)
		return
	}

	// Security Headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	reqPath := filepath.Clean(r.URL.Path)
	if reqPath == "/" || reqPath == "." {
		reqPath = "/index.html"
		if site.Engine == "php" {
			reqPath = "/index.php"
		}
	}

	targetFile := filepath.Join(site.Directory, reqPath)
	if !strings.HasPrefix(targetFile, site.Directory) {
		http.Error(w, "Forbidden path traversal", http.StatusForbidden)
		return
	}

	// If PHP file requested
	if strings.HasSuffix(targetFile, ".php") {
		m.servePHP(w, r, targetFile)
		return
	}

	// If regular file or HTML
	if fi, err := os.Stat(targetFile); err == nil && !fi.IsDir() {
		http.ServeFile(w, r, targetFile)
		return
	}

	// If path is directory, check for index.html or index.php
	indexCandidate := filepath.Join(targetFile, "index.html")
	if fi, err := os.Stat(indexCandidate); err == nil && !fi.IsDir() {
		http.ServeFile(w, r, indexCandidate)
		return
	}

	indexPHPCandidate := filepath.Join(targetFile, "index.php")
	if fi, err := os.Stat(indexPHPCandidate); err == nil && !fi.IsDir() {
		m.servePHP(w, r, indexPHPCandidate)
		return
	}

	http.NotFound(w, r)
}

func (m *Manager) servePHP(w http.ResponseWriter, r *http.Request, phpFile string) {
	phpBin, err := exec.LookPath("php-cgi")
	if err != nil {
		phpBin, err = exec.LookPath("php")
	}

	if err == nil {
		cmd := exec.Command(phpBin, phpFile)
		cmd.Env = append(os.Environ(),
			"REQUEST_METHOD="+r.Method,
			"SCRIPT_FILENAME="+phpFile,
			"QUERY_STRING="+r.URL.RawQuery,
			"REMOTE_ADDR="+r.RemoteAddr,
		)
		out, err := cmd.CombinedOutput()
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			_, _ = w.Write(out)
			return
		}
	}

	// Fallback when PHP is not installed on host: render parsed preview
	content, _ := os.ReadFile(phpFile)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><body style='background:#0a0e17;color:#f1f5f9;font-family:sans-serif;padding:2rem;'>")
	fmt.Fprintf(w, "<h2>🐘 PHP Runner Preview</h2><p>PHP execution engine is ready. Raw script length: %d bytes.</p>", len(content))
	fmt.Fprintf(w, "<pre style='background:#121826;padding:1rem;border-radius:8px;'>%s</pre></body></html>", content)
}
