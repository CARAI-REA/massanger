package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTokenBearerPreferred(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/ws?token=query-token", nil)
	r.Header.Set("Authorization", "Bearer header-token")
	require.Equal(t, "header-token", extractToken(r, true))
	require.Equal(t, "header-token", extractToken(r, false))
}

func TestExtractTokenQueryAllowed(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/ws?token=query-token", nil)
	require.Equal(t, "query-token", extractToken(r, true))
}

func TestExtractTokenQueryDenied(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/ws?token=query-token", nil)
	require.Equal(t, "", extractToken(r, false))
}
