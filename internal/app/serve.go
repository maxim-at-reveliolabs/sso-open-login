package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"sso-auto-login/internal/allowurl"
)

const (
	listenAddr = "127.0.0.1:18201"
	maxLine    = 2048
)

type opener interface {
	Open(ctx context.Context, u allowurl.URL) error
}

func serve(ctx context.Context, ln net.Listener, opener opener, log *slog.Logger) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept: %w", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle(ctx, conn, opener, log)
		}()
	}
}

func handle(ctx context.Context, conn net.Conn, opener opener, log *slog.Logger) {
	defer conn.Close()
	line, err := readLine(conn)
	if err != nil {
		return
	}
	u, err := allowurl.Parse(line)
	if err != nil {
		log.Info("ignored url")
		return
	}
	if err := opener.Open(ctx, u); err != nil {
		log.Error("login failed", slog.Any("err", err))
	}
}

func readLine(conn net.Conn) (string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return "", err
	}
	r := bufio.NewReaderSize(io.LimitReader(conn, maxLine), maxLine)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
