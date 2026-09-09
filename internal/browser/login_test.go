package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"sso-auto-login/internal/config"
)

func Test_fillOkta_types_totp_before_verify(t *testing.T) {
	exe := os.Getenv("CAMOUFOX")
	if exe == "" {
		exe = "/home/maxim/.cache/camoufox/camoufox-bin"
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("CAMOUFOX %s: %v", exe, err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/totp" {
			http.ServeFile(w, r, "testdata/okta-totp.html")
			return
		}
		http.ServeFile(w, r, "testdata/okta-login.html")
	}))
	t.Cleanup(srv.Close)

	pw, bctx, err := openCamoufox(exe, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = bctx.Close()
		_ = pw.Stop()
	})

	page, err := bctx.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = page.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if _, err := page.Goto(srv.URL + "/login"); err != nil {
		t.Fatalf("goto login: %v", err)
	}
	cfg := config.Config{
		Username:   "user@example.com",
		Password:   "secret",
		TOTPSecret: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
	}
	if err := fillOkta(ctx, page, cfg); err != nil {
		t.Fatalf("fillOkta: %v", err)
	}

	raw, err := page.Evaluate(`() => {
		const b = document.querySelector('button[type="submit"]');
		return b ? b.getAttribute('data-first') : null;
	}`)
	if err != nil {
		t.Fatalf("data-first: %v url=%s", err, page.URL())
	}
	first, _ := raw.(string)
	if len(first) != 6 {
		t.Fatalf("first Verify code %q, want 6 digits url=%s", first, page.URL())
	}
}
