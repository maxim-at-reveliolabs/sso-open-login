package browser

import (
	"context"
	"log/slog"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func ms(d time.Duration) *float64 {
	return playwright.Float(float64(d.Milliseconds()))
}

func namedButton(page playwright.Page, name string) playwright.Locator {
	exact := true
	vis := true
	return page.GetByRole(*playwright.AriaRoleButton, playwright.PageGetByRoleOptions{Name: name, Exact: &exact}).
		Filter(playwright.LocatorFilterOptions{Visible: &vis})
}

func clickAWS(ctx context.Context, page playwright.Page, wait time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	confirm := namedButton(page, "Confirm and continue")
	allow := namedButton(page, "Allow access")
	if err := confirm.Or(allow).WaitFor(playwright.LocatorWaitForOptions{Timeout: ms(wait)}); err != nil {
		return err
	}
	if n, err := confirm.Count(); err == nil && n > 0 {
		if err := confirm.Click(playwright.LocatorClickOptions{Timeout: ms(5 * time.Second)}); err == nil {
			slog.Info("aws button clicked", slog.String("name", "Confirm and continue"))
		}
	}
	if err := allow.Click(playwright.LocatorClickOptions{Timeout: ms(wait)}); err != nil {
		return err
	}
	slog.Info("aws button clicked", slog.String("name", "Allow access"))
	return nil
}

func allowVisible(page playwright.Page) bool {
	n, err := namedButton(page, "Confirm and continue").Or(namedButton(page, "Allow access")).Count()
	return err == nil && n > 0
}

func fillDeviceCode(ctx context.Context, page playwright.Page, code string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if code == "" {
		return nil
	}
	box := page.GetByLabel("Verification code").Or(page.GetByPlaceholder("code"))
	if err := box.Fill(code, playwright.LocatorFillOptions{Timeout: ms(5 * time.Second)}); err != nil {
		return nil
	}
	slog.Info("device code filled")
	return namedButton(page, "Continue").Click(playwright.LocatorClickOptions{Timeout: ms(5 * time.Second)})
}
