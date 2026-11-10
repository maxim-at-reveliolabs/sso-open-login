package browser

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func granted(ctx context.Context, page playwright.Page) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	loc := approved(page)
	if err := loc.First().WaitFor(playwright.LocatorWaitForOptions{Timeout: ms(45 * time.Second)}); err != nil {
		saveShot(page, "fail.png")
		slog.Error("aws grant did not finish")
		return err
	}
	saveShot(page, "granted.png")
	slog.Info("aws access granted")
	return nil
}

func saveShot(page playwright.Page, name string) {
	buf, err := page.Screenshot()
	if err != nil {
		slog.Error("screenshot failed", slog.Any("err", err))
		return
	}
	dir, err := profileDir()
	if err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, name), buf, 0o600); err != nil {
		slog.Error("screenshot failed", slog.Any("err", err))
	}
}
