package provider

import (
	core "github.com/wasmcloud/interfaces/core/tinygo"
)

type ProviderOption func(*WasmcloudProvider) error

func WithProviderActionFunc(inFunc func(ProviderAction) (*ProviderResponse, error)) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.providerActionFunc = inFunc
		return nil
	}
}

func WithNewLinkFunc(inFunc func(core.LinkDefinition) error) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.newLinkFunc = func(linkdef core.LinkDefinition) error {
			if wp.isLinked(linkdef.ActorId) {
				wp.Logger.Info("duplicate link", "actorId", linkdef.ActorId)
				return nil
			}
			err := inFunc(linkdef)
			if err != nil {
				return err
			}

			go wp.listenForActor(linkdef.ActorId)
			wp.putLink(linkdef)

			return nil
		}
		return nil
	}
}

func WithDelLinkFunc(inFunc func(core.LinkDefinition) error) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.delLinkFunc = func(linkdef core.LinkDefinition) error {
			err := inFunc(linkdef)
			if err != nil {
				return err
			}
			// shutdown specific NATs subscription
			wp.natsSubscriptions[linkdef.ActorId].Drain()
			wp.natsSubscriptions[linkdef.ActorId].Unsubscribe()
			
			wp.deleteLink(linkdef)

			return nil
		}
		return nil
	}
}

func WithShutdownFunc(inFunc func() error) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.shutdownFunc = func() error {
			err := inFunc()
			if err != nil {
				return err
			}

			for _, s := range wp.natsSubscriptions {
				err := s.Drain()
				if err != nil {
					return err
				}
			}
			err = wp.natsConnection.Drain()
			if err != nil {
				return err
			}

			wp.cancel()
			return nil
		}
		return nil
	}
}

func WithHealthCheckMsg(inFunc func() string) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.healthMsgFunc = inFunc
		return nil
	}
}
