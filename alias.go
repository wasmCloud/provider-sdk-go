package provider

import "go.wasmcloud.dev/provider"

type (
	HeahthCheckResponse     = provider.HealthCheckResponse
	HostData                = provider.HostData
	InterfaceLinkDefinition = provider.InterfaceLinkDefinition
	Level                   = provider.Level
	OtelConfig              = provider.OtelConfig
	ProviderHandler         = provider.ProviderHandler
	SecretBytesValue        = provider.SecretBytesValue
	SecretStringValue       = provider.SecretStringValue
	SecretValue             = provider.SecretValue
	Topics                  = provider.Topics
	WasmcloudProvider       = provider.WasmcloudProvider
)

const (
	Critical         = provider.Critical
	Debug            = provider.Debug
	Error            = provider.Error
	Info             = provider.Info
	OtelProtocolGRPC = provider.OtelProtocolGRPC
	OtelProtocolHTTP = provider.OtelProtocolHTTP
	Trace            = provider.Trace
	Warn             = provider.Warn
)

var (
	HealthCheck   = provider.HealthCheck
	LatticeTopics = provider.LatticeTopics
	New           = provider.New
	Shutdown      = provider.Shutdown
	SourceLinkDel = provider.SourceLinkDel
	SourceLinkPut = provider.SourceLinkPut
	TargetLinkDel = provider.TargetLinkDel
	TargetLinkPut = provider.TargetLinkPut
)
