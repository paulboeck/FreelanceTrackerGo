package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestE2ESimple is a very basic test to verify e2e setup works
func TestE2ESimple(t *testing.T) {
	ctx := SetupE2ETest(t)

	t.Run("can navigate to login page", func(t *testing.T) {
		// Just try to navigate with timeout
		ctx.Page.Timeout(10 * time.Second).MustNavigate(ctx.ServerURL + "/user/login")

		// Get the URL to verify navigation worked
		info := ctx.Page.Timeout(5 * time.Second).MustInfo()
		t.Logf("Navigated to: %s", info.URL)

		require.Contains(t, info.URL, "/user/login")
	})
}
