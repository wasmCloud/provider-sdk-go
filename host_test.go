package provider

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactedStringLogging(t *testing.T) {
	var buf bytes.Buffer
	secret := "its-a-secret"
	redactedString := RedactedString(secret)

	jsonSlog := slog.New(slog.NewJSONHandler(&buf, nil))
	jsonSlog.Info("jsonSlog", "redactedString", redactedString)

	if !strings.Contains(buf.String(), "\"redactedString\":\"redacted(string)\"") || strings.Contains(buf.String(), secret) {
		t.Error("json slog handler output should not have contained the secret string")
	}

	buf.Reset()

	textSlog := slog.New(slog.NewTextHandler(&buf, nil))
	textSlog.Info("textSlog", "redactedString", redactedString)

	if !strings.Contains(buf.String(), "redactedString=redacted(string)") || strings.Contains(buf.String(), secret) {
		t.Error("text slog handler should not have contained the secret string")
	}
}

func TestOtelConfigProtocol(t *testing.T) {
	type test struct {
		name     string
		config   OtelConfig
		protocol string
	}

	tests := []test{
		{
			name:     "Defaults to http",
			config:   OtelConfig{},
			protocol: "http",
		},
		{
			name:     "Explicit Grpc Rust enum variant",
			config:   OtelConfig{Protocol: "Grpc"},
			protocol: "grpc",
		},
		{
			name:     "Explicit Http Rust enum variant",
			config:   OtelConfig{Protocol: "Http"},
			protocol: "http",
		},
	}

	for _, tc := range tests {
		if tc.config.OtelProtocol() != tc.protocol {
			t.Fatalf("%s / OtelProtocol: expected %q, got: %q", tc.name, tc.protocol, tc.config.OtelProtocol())
		}
	}
}

func TestOtelConfigURLs(t *testing.T) {
	type test struct {
		name       string
		config     OtelConfig
		tracesURL  string
		metricsURL string
		logsURL    string
	}

	tests := []test{
		{
			name:       "Defaults with HTTP",
			config:     OtelConfig{},
			tracesURL:  "http://localhost:4318/v1/traces",
			metricsURL: "http://localhost:4318/v1/metrics",
			logsURL:    "http://localhost:4318/v1/logs",
		},
		{
			name:       "Defaults with gRPC",
			config:     OtelConfig{Protocol: "Grpc"},
			tracesURL:  "http://localhost:4317",
			metricsURL: "http://localhost:4317",
			logsURL:    "http://localhost:4317",
		},
		{
			name:       "Custom ObservabilityEndpoint",
			config:     OtelConfig{ObservabilityEndpoint: "https://api.opentelemetry.com"},
			tracesURL:  "https://api.opentelemetry.com/v1/traces",
			metricsURL: "https://api.opentelemetry.com/v1/metrics",
			logsURL:    "https://api.opentelemetry.com/v1/logs",
		},
		{
			name:       "Custom ObservabilityEndpoint with gRPC",
			config:     OtelConfig{Protocol: "grpc", ObservabilityEndpoint: "https://api.opentelemetry.com"},
			tracesURL:  "https://api.opentelemetry.com",
			metricsURL: "https://api.opentelemetry.com",
			logsURL:    "https://api.opentelemetry.com",
		},
		{
			name:       "Custom TracesEndpoint",
			config:     OtelConfig{TracesEndpoint: "https://api.opentelemetry.com/v1/traces"},
			tracesURL:  "https://api.opentelemetry.com/v1/traces",
			metricsURL: "http://localhost:4318/v1/metrics",
			logsURL:    "http://localhost:4318/v1/logs",
		},
		{
			name:       "Custom MetricsEndpoint",
			config:     OtelConfig{MetricsEndpoint: "https://api.opentelemetry.com/v1/metrics"},
			tracesURL:  "http://localhost:4318/v1/traces",
			metricsURL: "https://api.opentelemetry.com/v1/metrics",
			logsURL:    "http://localhost:4318/v1/logs",
		},
		{
			name:       "Custom LogsEndpoint",
			config:     OtelConfig{LogsEndpoint: "https://api.opentelemetry.com/v1/logs"},
			tracesURL:  "http://localhost:4318/v1/traces",
			metricsURL: "http://localhost:4318/v1/metrics",
			logsURL:    "https://api.opentelemetry.com/v1/logs",
		},
	}

	for _, tc := range tests {
		if tc.config.TracesURL() != tc.tracesURL {
			t.Fatalf("%s / TracesURL: expected %s, got: %s", tc.name, tc.tracesURL, tc.config.TracesURL())
		}
		if tc.config.MetricsURL() != tc.metricsURL {
			t.Fatalf("%s / MetricsURL: expected %s, got: %s", tc.name, tc.metricsURL, tc.config.MetricsURL())
		}
		if tc.config.LogsURL() != tc.logsURL {
			t.Fatalf("%s / LogsURL: expected %s, got: %s", tc.name, tc.logsURL, tc.config.LogsURL())
		}
	}
}

func TestOtelConfigBooleans(t *testing.T) {
	type test struct {
		name           string
		config         OtelConfig
		tracesEnabled  bool
		metricsEnabled bool
		logsEnabled    bool
	}

	tests := []test{
		{
			name:           "Defaults",
			config:         OtelConfig{},
			tracesEnabled:  false,
			metricsEnabled: false,
			logsEnabled:    false,
		},
		{
			name:           "Enable all with EnableObservability",
			config:         OtelConfig{EnableObservability: true},
			tracesEnabled:  true,
			metricsEnabled: true,
			logsEnabled:    true,
		},
		{
			name:           "Enable just traces",
			config:         OtelConfig{EnableTraces: true},
			tracesEnabled:  true,
			metricsEnabled: false,
			logsEnabled:    false,
		},
		{
			name:           "Enable just metrics",
			config:         OtelConfig{EnableMetrics: true},
			tracesEnabled:  false,
			metricsEnabled: true,
			logsEnabled:    false,
		},
		{
			name:           "Enable just logs",
			config:         OtelConfig{EnableLogs: true},
			tracesEnabled:  false,
			metricsEnabled: false,
			logsEnabled:    true,
		},
	}
	for _, tc := range tests {
		if tc.config.TracesEnabled() != tc.tracesEnabled {
			t.Fatalf("%s / TracesEnabled: expected %t, got: %t", tc.name, tc.tracesEnabled, tc.config.TracesEnabled())
		}
		if tc.config.MetricsEnabled() != tc.metricsEnabled {
			t.Fatalf("%s / MetricsEnabled: expected %t, got: %t", tc.name, tc.metricsEnabled, tc.config.MetricsEnabled())
		}
		if tc.config.LogsEnabled() != tc.logsEnabled {
			t.Fatalf("%s / LogsEnabled: expected %t, got: %t", tc.name, tc.logsEnabled, tc.config.LogsEnabled())
		}
	}
}
