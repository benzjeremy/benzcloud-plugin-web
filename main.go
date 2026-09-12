package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/benzjeremy/benzcloud-plugin-web/internal/sites"
)

//go:embed web/*
var webFS embed.FS

const Version = "v1.0"

func main() {
	var (
		portFlag    int
		domainFlag  string
		tokenFlag   string
		dataDirFlag string
		versionFlag bool
	)

	flag.IntVar(&portFlag, "port", 8091, "Web plugin HTTP port")
	flag.StringVar(&domainFlag, "domain", "intern", "Base domain")
	flag.StringVar(&tokenFlag, "token", "", "BenzCloud server auth token")
	flag.StringVar(&dataDirFlag, "datadir", "", "Sites data directory")
	flag.BoolVar(&versionFlag, "version", false, "Print version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Printf("BenzCloud Web Plugin %s (Lead Engineer: Jeremy Benz • GNU GPLv3)\n", Version)
		return
	}

	if dataDirFlag == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			dataDirFlag = "./benzcloud-web-data"
		} else {
			dataDirFlag = filepath.Join(home, ".benzcloud", "plugins", "web")
		}
	}

	mgr, err := sites.NewManager(dataDirFlag, domainFlag)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize sites manager: %v\n", err)
	}

	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Fatal: failed to extract web assets: %v\n", err)
	}
	uiFileServer := http.FileServer(http.FS(subFS))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health & IPC endpoint for benzcloud-server
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "healthy",
				"subdomains": mgr.Subdomains(),
				"version":    Version,
			})
			return
		}

		// API endpoints for managing sites
		if strings.HasPrefix(r.URL.Path, "/api/sites") {
			switch r.Method {
			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"sites":       mgr.ListSites(),
					"base_domain": domainFlag,
				})
				return
			case http.MethodPost:
				var req struct {
					Subdomain string `json:"subdomain"`
					Title     string `json:"title"`
					Engine    string `json:"engine"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "Invalid payload", http.StatusBadRequest)
					return
				}
				s, err := mgr.CreateSite(req.Subdomain, req.Engine, req.Title)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(s)
				return
			case http.MethodDelete:
				sub := r.URL.Query().Get("subdomain")
				if err := mgr.DeleteSite(sub); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
				return
			}
		}

		// Host detection: Is this request for a specific hosted site subdomain?
		host := strings.ToLower(r.Host)
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		cleanBase := strings.ToLower(strings.Trim(domainFlag, "."))
		suffix := "." + cleanBase

		if strings.HasSuffix(host, suffix) {
			sub := strings.TrimSuffix(host, suffix)
			if sub != "" {
				mgr.ServeSite(w, r, sub)
				return
			}
		}

		// Default: Serve the Plugin UI
		uiFileServer.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", portFlag),
		Handler: handler,
	}

	go func() {
		log.Printf("🚀 [BenzCloud Web Plugin %s] Listening on port %d (Domain: %s)\n", Version, portFlag, domainFlag)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal: web plugin server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Stopping BenzCloud Web Plugin...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("Web Plugin stopped safely.")
}
