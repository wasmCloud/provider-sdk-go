package main

import (
	"github.com/wasmCloud/provider-sdk-go"
	"log"
)

func (p *Provider) establishSourceLink(link provider.InterfaceLinkDefinition) error {
	if err := p.validateSourceLink(link); err != nil {
		return err
	}

	p.sourceLinks[link.Target] = link
	return nil
}

func (p *Provider) establishTargetLink(link provider.InterfaceLinkDefinition) error {
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

func (p *Provider) handleNewSourceLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling new source link", link)
	err := p.establishSourceLink(link)
	if err != nil {
		log.Println("Failed to establish source link", link, err)
		p.failedSourceLinks[link.Target] = link
		return err
	}
	p.sourceLinks[link.Target] = link
	return nil
}

func (p *Provider) handleNewTargetLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling new target link", link)
	err := p.establishTargetLink(link)
	if err != nil {
		log.Println("Failed to establish target link", link, err)
		p.failedTargetLinks[link.SourceID] = link
		return err
	}
	p.targetLinks[link.SourceID] = link
	return nil
}

func (p *Provider) handleDelSourceLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling del source link", link)
	delete(p.sourceLinks, link.Target)
	return nil
}

func (p *Provider) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling del target link", link)
	delete(p.targetLinks, link.SourceID)
	return nil
}

func (p *Provider) handleHealthCheck() string {
	log.Println("Handling health check")
	return "provider healthy"
}

func (p *Provider) handleShutdown() error {
	log.Println("Handling shutdown")
	return nil
}
