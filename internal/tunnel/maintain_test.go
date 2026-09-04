package tunnel

import (
	"testing"
	"time"
)

func Test_nextBackoff_grows_until_max_then_resets_after_stable_run(t *testing.T) {
	t.Parallel()

	got := nextBackoff(time.Second, 0)
	if got != 2*time.Second {
		t.Fatalf("after fail = %s, want 2s", got)
	}
	got = nextBackoff(maxBackoff, 0)
	if got != maxBackoff {
		t.Fatalf("at max = %s, want %s", got, maxBackoff)
	}
	got = nextBackoff(maxBackoff, stableFor)
	if got != minBackoff {
		t.Fatalf("after stable = %s, want %s", got, minBackoff)
	}
}
