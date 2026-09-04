package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"sso-auto-login/internal/allowurl"
	"sso-auto-login/internal/browser"
	"sso-auto-login/internal/config"
	"sso-auto-login/internal/tunnel"
)

func Main(args []string) int {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(log)

	fs := flag.NewFlagSet("sso-auto-login", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	testURL := fs.String("test-url", "", "open one allowed URL and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *testURL != "" {
		if err := runTest(ctx, cfg, *testURL); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}

	if err := runDaemon(ctx, cfg, log); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runTest(ctx context.Context, cfg config.Config, raw string) error {
	u, err := allowurl.Parse(raw)
	if err != nil {
		return err
	}
	sess, err := browser.New(cfg)
	if err != nil {
		return err
	}
	return sess.Open(ctx, u)
}

func runDaemon(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	ln, err := net.Listen("tcp4", listenAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", listenAddr, err)
	}
	defer ln.Close()

	sess, err := browser.New(cfg)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tunnel.Maintain(ctx, tunnel.Spec{
			Host: cfg.SSHHost,
			User: cfg.SSHUser,
			Key:  cfg.SSHKey,
		}, log)
	}()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	err = serve(ctx, ln, sess, log)
	wg.Wait()
	if ctx.Err() != nil {
		return nil
	}
	return err
}
