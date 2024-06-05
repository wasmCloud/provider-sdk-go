package main

import (
	"github.com/wasmCloud/provider-sdk-go"
	"log"
)

func (p *Provider) establishSourceLink(link provider.InterfaceLinkDefinition) error {
	if _, exists := p.sourceLinks[link.Target]; exists {
		log.Println("Source link already exists, ignoring duplicate", link)
		return nil
	}

	if err := p.validateSourceLink(link); err != nil {
		return err
	}

	p.sourceLinks[link.Target] = link
	return nil
}

func (p *Provider) establishTargetLink(link provider.InterfaceLinkDefinition) error {
	if _, exists := p.targetLinks[link.SourceID]; exists {
		log.Println("Target link already exists, ignoring duplicate", link)
		return nil
	}

	if err := p.validateTargetLink(link); err != nil {
		return err
	}

	p.targetLinks[link.SourceID] = link
	return nil
}

func (p *Provider) validateSourceLink(link provider.InterfaceLinkDefinition) error {
	// TODO: Add validation checks
	return nil
}

func (p *Provider) validateTargetLink(link provider.InterfaceLinkDefinition) error {
	// TODO: Add validation checks
	return nil
}
