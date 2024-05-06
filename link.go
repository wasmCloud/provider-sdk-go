package provider

type InterfaceLinkDefinition struct {
	SourceID     string            `json:"source_id,omitempty"`
	Target       string            `json:"target,omitempty"`
	Name         string            `json:"name,omitempty"`
	WitNamespace string            `json:"wit_namespace,omitempty"`
	WitPackage   string            `json:"wit_package,omitempty"`
	Interfaces   []string          `json:"interfaces,omitempty"`
	SourceConfig map[string]string `json:"source_config,omitempty"`
	TargetConfig map[string]string `json:"target_config,omitempty"`
}
