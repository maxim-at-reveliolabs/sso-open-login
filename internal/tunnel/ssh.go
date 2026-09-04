package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Spec struct {
	Host string
	User string
	Key  string
}

func args(s Spec) []string {
	return []string{
		"-N", "-T",
		"-l", s.User,
		"-i", s.Key,
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-R", "18201:127.0.0.1:18201",
		s.Host,
	}
}

type tunnel struct {
	cmd *exec.Cmd
}

func start(s Spec) (*tunnel, error) {
	cmd := exec.Command("ssh", args(s)...)
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ssh: %w", err)
	}
	return &tunnel{cmd: cmd}, nil
}

func (t *tunnel) wait() error {
	if t == nil || t.cmd == nil {
		return nil
	}
	if err := t.cmd.Wait(); err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	return nil
}

func (t *tunnel) kill() {
	if t == nil || t.cmd == nil || t.cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-t.cmd.Process.Pid, syscall.SIGTERM)
}
