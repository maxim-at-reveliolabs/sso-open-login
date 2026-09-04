package tunnel

import (
	"slices"
	"testing"
)

func Test_args_uses_user_key_and_remote_forward(t *testing.T) {
	t.Parallel()

	got := args(Spec{
		Host: "dev.box",
		User: "maxim",
		Key:  "/home/maxim/code/revelio-labs/.ssh/maxim.pem",
	})
	want := []string{
		"-N", "-T",
		"-l", "maxim",
		"-i", "/home/maxim/code/revelio-labs/.ssh/maxim.pem",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-R", "18201:127.0.0.1:18201",
		"dev.box",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("args() = %v, want %v", got, want)
	}
}
