package browser

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/pquerna/otp/totp"

	"sso-auto-login/internal/allowurl"
	"sso-auto-login/internal/config"
)

func login(ctx context.Context, page playwright.Page, cfg config.Config, u allowurl.URL) error {
	if _, err := page.Goto(u.String()); err != nil {
		return fmt.Errorf("navigate: %w", err)
	}
	_ = fillDeviceCode(ctx, page, u.DeviceCode())
	if err := fillOkta(ctx, page, cfg); err != nil {
		return err
	}
	if err := clickAWS(ctx, page, 90*time.Second); err != nil {
		saveShot(page, "fail.png")
		slog.Error("aws allow failed")
		return fmt.Errorf("aws allow: %w", err)
	}
	return granted(ctx, page)
}

func fillOkta(ctx context.Context, page playwright.Page, cfg config.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if allowVisible(page) {
		return nil
	}
	user := page.Locator(`#okta-signin-username, [name="identifier"]`)
	aws := namedButton(page, "Confirm and continue").Or(namedButton(page, "Allow access"))
	if err := user.Or(aws).WaitFor(playwright.LocatorWaitForOptions{Timeout: ms(12 * time.Second)}); err != nil {
		return nil
	}
	if allowVisible(page) {
		return nil
	}
	if err := user.First().Fill(cfg.Username, playwright.LocatorFillOptions{Timeout: ms(5 * time.Second)}); err != nil {
		return nil
	}
	pass := page.Locator(`#okta-signin-password, input[type="password"]`)
	_ = pass.First().Fill(cfg.Password, playwright.LocatorFillOptions{Timeout: ms(3 * time.Second)})
	_ = clickOktaSubmit(page)
	_ = pass.First().Fill(cfg.Password, playwright.LocatorFillOptions{Timeout: ms(8 * time.Second)})
	_ = clickOktaSubmit(page)
	code, err := totp.GenerateCode(cfg.TOTPSecret, time.Now())
	if err != nil {
		slog.Error("okta totp failed")
		return fmt.Errorf("totp: %w", err)
	}
	totpBox := page.Locator(`[name="answer"], [name="passcode"], [name="credentials.passcode"]`)
	if err := totpBox.First().Fill(code, playwright.LocatorFillOptions{Timeout: ms(15 * time.Second)}); err == nil {
		_ = clickOktaSubmit(page)
	}
	if n, _ := page.Locator(`input[name="SAMLResponse"]`).Count(); n > 0 {
		_, _ = page.Evaluate(`() => { const el = document.querySelector('input[name="SAMLResponse"]'); if (el && el.form) el.form.submit(); }`)
	}
	return nil
}

func clickOktaSubmit(page playwright.Page) error {
	loc := page.Locator(`#okta-signin-submit, [data-type="save"], button[type="submit"]`)
	return loc.First().Click(playwright.LocatorClickOptions{Timeout: ms(4 * time.Second)})
}
