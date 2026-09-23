package affinity_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"sfu/internal/repository/affinity"
)

func TestNoopClaimRedirect(t *testing.T) {
	n := affinity.NewNoop()
	ok, url, err := n.Claim(context.Background(), "r1", "a", "ws://a", time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "ws://a", url)

	ok, url, err = n.Claim(context.Background(), "r1", "b", "ws://b", time.Minute)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, "ws://a", url)

	require.NoError(t, n.Release(context.Background(), "r1", "a"))
	ok, _, err = n.Claim(context.Background(), "r1", "b", "ws://b", time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
}
