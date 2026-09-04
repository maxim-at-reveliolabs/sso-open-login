package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_clickAWS_clicks_confirm_and_continue_then_allow_access(t *testing.T) {
	exe := os.Getenv("CAMOUFOX")
	if exe == "" {
		exe = "/home/maxim/.cache/camoufox/camoufox-bin"
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("CAMOUFOX %s: %v", exe, err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/allow":
			http.ServeFile(w, r, "testdata/allow.html")
		case "/done":
			http.ServeFile(w, r, "testdata/done.html")
		default:
			http.ServeFile(w, r, "testdata/confirm.html")
		}
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

	if _, err := page.Goto(srv.URL + "/confirm"); err != nil {
		t.Fatalf("goto confirm: %v", err)
	}
	if err := clickAWS(ctx, page, 30*time.Second); err != nil {
		t.Fatalf("clickAWS: %v", err)
	}

	body, err := page.Content()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "You can close this window") {
		t.Fatalf("after both buttons, url=%s body does not say done", page.URL())
	}
}
