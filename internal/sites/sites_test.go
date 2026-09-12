package sites

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSitesManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "benzcloud_plugin_web_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mgr, err := NewManager(tempDir, "benzjeremy.de")
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// 1. Initial provision of "home" site
	subs := mgr.Subdomains()
	if len(subs) == 0 {
		t.Fatal("Expected default home site to be provisioned")
	}

	// 2. Create custom sites
	blogSite, err := mgr.CreateSite("tech-blog", "html", "Jeremy Tech Blog")
	if err != nil {
		t.Fatalf("CreateSite failed: %v", err)
	}
	if blogSite.Subdomain != "tech-blog" {
		t.Fatalf("Expected subdomain tech-blog, got %s", blogSite.Subdomain)
	}

	// 3. Test reserved system subdomain protection
	_, err = mgr.CreateSite("vpn", "html", "Fake VPN")
	if err == nil {
		t.Fatal("Expected error creating reserved subdomain 'vpn'")
	}
	_, err = mgr.CreateSite("mail", "html", "Fake Mail")
	if err == nil {
		t.Fatal("Expected error creating reserved subdomain 'mail'")
	}

	// 4. Test HTTP serving
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mgr.ServeSite(rec, req, "tech-blog")

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 from ServeSite, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Jeremy Tech Blog") {
		t.Fatalf("Expected body to contain title, got: %s", rec.Body.String())
	}

	// 5. Test PHP engine creation
	_, err = mgr.CreateSite("api-test", "php", "PHP Backend")
	if err != nil {
		t.Fatalf("CreateSite php failed: %v", err)
	}
	recPHP := httptest.NewRecorder()
	reqPHP := httptest.NewRequest(http.MethodGet, "/", nil)
	mgr.ServeSite(recPHP, reqPHP, "api-test")
	if recPHP.Code != http.StatusOK {
		t.Fatalf("Expected 200 from ServeSite php, got %d", recPHP.Code)
	}

	// 6. Delete site
	if err := mgr.DeleteSite("tech-blog"); err != nil {
		t.Fatalf("DeleteSite failed: %v", err)
	}
	recDel := httptest.NewRecorder()
	mgr.ServeSite(recDel, req, "tech-blog")
	if recDel.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for deleted site, got %d", recDel.Code)
	}
}
