package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/altinity/altinity-mcp/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestRegisterMetricsRouteDisabledByDefault(t *testing.T) {
	mux := http.NewServeMux()
	registerMetricsRoute(mux, config.Config{})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRegisterMetricsRouteWhenEnabled(t *testing.T) {
	mux := http.NewServeMux()
	cfg := config.Config{}
	cfg.Server.Metrics.Enabled = true
	registerMetricsRoute(mux, cfg)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, strings.Contains(rr.Body.String(), "altinity_mcp_clickhouse_up"))
}
