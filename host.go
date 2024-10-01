package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RedactedString string

func (rs RedactedString) String() string {
	if rs != "" {
		return "redacted(string)"
	}
	return ""
}

func (rs RedactedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(rs.String())
}

func (rs RedactedString) Reveal() string {
	return string(rs)
}

type OtelConfig struct {
	EnableObservability   bool   `json:"enable_observability"`
	EnableTraces          bool   `json:"enable_traces,omitempty"`
	EnableMetrics         bool   `json:"enable_metrics,omitempty"`
	EnableLogs            bool   `json:"enable_logs,omitempty"`
	ObservabilityEndpoint string `json:"observability_endpoint,omitempty"`
	TracesEndpoint        string `json:"traces_endpoint,omitempty"`
	MetricsEndpoint       string `json:"metrics_endpoint,omitempty"`
	LogsEndpoint          string `json:"logs_endpoint,omitempty"`
	Protocol              string `json:"protocol,omitempty"`
}

type otelSignal int

const (
	traces otelSignal = iota
	metrics
	logs

	// https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint
	OtelExporterGrpcEndpoint = "http://localhost:4317"
	OtelExporterHttpEndpoint = "http://localhost:4318"

	// https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_traces_endpoint
	OtelExporterHttpTracesPath = "/v1/traces"
	// https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_metrics_endpoint
	OtelExporterHttpMetricsPath = "/v1/metrics"
	// https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_logs_endpoint
	OtelExporterHttpLogsPath = "/v1/logs"
)

// OtelProtocol returns the configured OpenTelemetry protocol if one is provided,
// otherwise defaulting to http.
func (config *OtelConfig) OtelProtocol() string {
	if config.Protocol != "" {
		return strings.ToLower(config.Protocol)
	}

	return OtelProtocolHTTP
}

// TracesURL returns the configured TracesEndpoint as-is if one is provided,
// otherwise it resolves the URL based on ObservabilityEndpoint value and the
// Protocol appropriate path.
func (config *OtelConfig) TracesURL() string {
	if config.TracesEndpoint != "" {
		return config.TracesEndpoint
	}

	return config.resolveSignalUrl(traces)
}

// MetricsURL returns the configured MetricsEndpoint as-is if one is provided,
// otherwise it resolves the URL based on ObservabilityEndpoint value and the
// Protocol appropriate path.
func (config *OtelConfig) MetricsURL() string {
	if config.MetricsEndpoint != "" {
		return config.MetricsEndpoint
	}

	return config.resolveSignalUrl(metrics)
}

// LogsURL returns the configured LogsEndpoint as-is if one is provided,
// otherwise it resolves the URL based on ObservabilityEndpoint value and the
// Protocol appropriate path.
func (config *OtelConfig) LogsURL() string {
	if config.LogsEndpoint != "" {
		return config.LogsEndpoint
	}

	return config.resolveSignalUrl(logs)
}

// TracesEnabled returns whether emitting traces has been enabled.
func (config *OtelConfig) TracesEnabled() bool {
	return config.EnableObservability || config.EnableTraces
}

// MetricsEnabled returns whether emitting metrics has been enabled.
func (config *OtelConfig) MetricsEnabled() bool {
	return config.EnableObservability || config.EnableMetrics
}

// LogsEnabled returns whether emitting logs has been enabled.
func (config *OtelConfig) LogsEnabled() bool {
	return config.EnableObservability || config.EnableLogs
}

func (config *OtelConfig) resolveSignalUrl(signal otelSignal) string {
	endpoint := config.defaultEndpoint()
	if config.ObservabilityEndpoint != "" {
		endpoint = config.ObservabilityEndpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")

	return fmt.Sprintf("%s%s", endpoint, config.defaultSignalPath(signal))
}

func (config *OtelConfig) defaultEndpoint() string {
	if config.OtelProtocol() == OtelProtocolGRPC {
		return OtelExporterGrpcEndpoint
	}

	return OtelExporterHttpEndpoint
}

func (config *OtelConfig) defaultSignalPath(signal otelSignal) string {
	// In case of gRPC, we return empty string gRPC doesn't need a path to be set for it.
	if config.OtelProtocol() == OtelProtocolGRPC {
		return ""
	}

	switch signal {
	case traces:
		return OtelExporterHttpTracesPath
	case metrics:
		return OtelExporterHttpMetricsPath
	case logs:
		return OtelExporterHttpLogsPath
	}
	return ""
}

type HostData struct {
	HostID                 string                     `json:"host_id,omitempty"`
	LatticeRPCPrefix       string                     `json:"lattice_rpc_prefix,omitempty"`
	LatticeRPCUserJWT      string                     `json:"lattice_rpc_user_jwt,omitempty"`
	LatticeRPCUserSeed     string                     `json:"lattice_rpc_user_seed,omitempty"`
	LatticeRPCURL          string                     `json:"lattice_rpc_url,omitempty"`
	ProviderKey            string                     `json:"provider_key,omitempty"`
	EnvValues              map[string]string          `json:"env_values,omitempty"`
	InstanceID             string                     `json:"instance_id,omitempty"`
	LinkDefinitions        []linkWithEncryptedSecrets `json:"link_definitions,omitempty"`
	ClusterIssuers         []string                   `json:"cluster_issuers,omitempty"`
	Config                 map[string]string          `json:"config,omitempty"`
	Secrets                map[string]SecretValue     `json:"secrets,omitempty"`
	HostXKeyPublicKey      string                     `json:"host_xkey_public_key,omitempty"`
	ProviderXKeyPrivateKey RedactedString             `json:"provider_xkey_private_key,omitempty"`
	DefaultRPCTimeoutMS    *uint64                    `json:"default_rpc_timeout_ms,omitempty"`
	StructuredLogging      bool                       `json:"structured_logging,omitempty"`
	LogLevel               *Level                     `json:"log_level,omitempty"`
	OtelConfig             OtelConfig                 `json:"otel_config,omitempty"`
}

type HealthCheckResponse struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}
