package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"

	"sso-auto-login/internal/allowurl"
)

type fakeOpener struct {
	mu   sync.Mutex
	urls []string
}

func (f *fakeOpener) Open(_ context.Context, u allowurl.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.urls = append(f.urls, u.String())
	return nil
}

func (f *fakeOpener) got() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.urls))
	copy(out, f.urls)
	return out
}

func send(t *testing.T, opener *fakeOpener, line string) {
	t.Helper()
	c, s := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		handle(context.Background(), s, opener, slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	if _, err := c.Write([]byte(line)); err != nil {
		t.Fatal(err)
	}
	_ = c.Close()
	<-done
}

func Test_handle_opens_allowed_url_and_ignores_other_hosts(t *testing.T) {
	t.Parallel()

	opener := &fakeOpener{}
	send(t, opener, "https://revelio.awsapps.com/start/#/device?user_code=TEST-TEST\nextra\n")
	got := opener.got()
	if len(got) != 1 || !strings.Contains(got[0], "revelio.awsapps.com") {
		t.Fatalf("opened %v, want one awsapps device url", got)
	}
	send(t, opener, "https://evil.example.com/\n")
	if n := len(opener.got()); n != 1 {
		t.Fatalf("opened %d urls, want 1", n)
	}
}
