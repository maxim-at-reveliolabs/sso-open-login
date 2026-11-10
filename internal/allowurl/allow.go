package allowurl

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrRejected = errors.New("url not allowed")

type URL struct {
	raw  string
	code string
}

func (u URL) String() string { return u.raw }

func (u URL) DeviceCode() string { return u.code }

func Parse(raw string) (URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return URL{}, fmt.Errorf("empty: %w", ErrRejected)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return URL{}, fmt.Errorf("parse: %w", ErrRejected)
	}
	if u.Scheme != "https" {
		return URL{}, fmt.Errorf("scheme %q: %w", u.Scheme, ErrRejected)
	}
	if u.User != nil {
		return URL{}, fmt.Errorf("userinfo: %w", ErrRejected)
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "" {
		return URL{}, fmt.Errorf("empty host: %w", ErrRejected)
	}
	if !deviceLogin(u, host) && !signinAuthorize(u, host) {
		return URL{}, fmt.Errorf("host %q: %w", host, ErrRejected)
	}
	return URL{raw: u.String(), code: userCode(u)}, nil
}

func userCode(u *url.URL) string {
	frag, err := url.Parse(u.Fragment)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(frag.Query().Get("user_code"))
}

func deviceLogin(u *url.URL, host string) bool {
	if !strings.HasSuffix(host, ".awsapps.com") {
		return false
	}
	if strings.TrimSuffix(u.Path, "/") != "/start" {
		return false
	}
	frag, err := url.Parse(u.Fragment)
	if err != nil {
		return false
	}
	return strings.TrimSuffix(frag.Path, "/") == "/device" &&
		strings.TrimSpace(frag.Query().Get("user_code")) != ""
}

func signinAuthorize(u *url.URL, host string) bool {
	name, ok := strings.CutSuffix(host, ".signin.aws")
	if !ok || name == "" || strings.Contains(name, ".") {
		return false
	}
	if strings.TrimSuffix(u.Path, "/") != "/platform/authorize" {
		return false
	}
	q := u.Query()
	return q.Get("request_uri") != "" && q.Get("identitystore_id") != ""
}
