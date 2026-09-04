package browser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"sso-auto-login/internal/allowurl"
	"sso-auto-login/internal/config"
)

const loginTimeout = 3 * time.Minute

type Session struct {
	cfg config.Config
	mu  sync.Mutex
}

func New(cfg config.Config) (*Session, error) {
	if _, err := totp.GenerateCode(cfg.TOTPSecret, time.Now()); err != nil {
		if errors.Is(err, otp.ErrValidateSecretInvalidBase32) {
			return nil, fmt.Errorf("totp_secret is not valid base32")
		}
		return nil, fmt.Errorf("totp_secret: %w", err)
	}
	return &Session{cfg: cfg}, nil
}

func openCamoufox(exe, profile string) (*playwright.Playwright, playwright.BrowserContext, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("start playwright: %w", err)
	}
	headless := true
	noSig := false
	noView := true
	env := map[string]string{}
	for _, e := range os.Environ() {
		k, v, ok := strings.Cut(e, "=")
		if ok {
			env[k] = v
		}
	}
	env["CAMOU_CONFIG"] = `{"addons":[]}`
	bctx, err := pw.Firefox.LaunchPersistentContext(profile, playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath: &exe,
		Headless:       &headless,
		HandleSIGHUP:   &noSig,
		HandleSIGINT:   &noSig,
		HandleSIGTERM:  &noSig,
		NoViewport:     &noView,
		Env:            env,
		FirefoxUserPrefs: map[string]any{
			"extensions.enabledScopes":     0,
			"extensions.autoDisableScopes": 15,
		},
	})
	if err != nil {
		_ = pw.Stop()
		return nil, nil, fmt.Errorf("start camoufox: %w", err)
	}
	return pw, bctx, nil
}

func (s *Session) Open(ctx context.Context, u allowurl.URL) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, err := profileDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return fmt.Errorf("camoufox profile: %w", err)
	}
	pw, bctx, err := openCamoufox(s.cfg.Camoufox, profile)
	if err != nil {
		return err
	}
	defer func() {
		_ = bctx.Close()
		_ = pw.Stop()
	}()
	page, err := bctx.NewPage()
	if err != nil {
		return fmt.Errorf("new page: %w", err)
	}
	defer page.Close()
	tctx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()
	return login(tctx, page, s.cfg, u)
}

func profileDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("camoufox profile: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("camoufox profile: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), "camoufox-profile"), nil
}
