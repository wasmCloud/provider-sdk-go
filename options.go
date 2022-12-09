package provider

type ProviderOption func(*WasmcloudProvider) error

func WithProviderActionFunc(inFunc func(ProviderAction) (*ProviderResponse, error)) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.providerActionFunc = inFunc
		return nil
	}
}

func WithNewLinkFunc(inFunc func(ActorConfig) error) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.newLinkFunc = func(a ActorConfig) error {
			err := inFunc(a)
			if err != nil {
				return err
			}

			go wp.listenForActor(a.ActorID)
			return nil
		}
		return nil
	}
}

func WithDelLinkFunc(inFunc func(ActorConfig) error) ProviderOption {
	return func(wp *WasmcloudProvider) error {
		wp.delLinkFunc = func(a ActorConfig) error {
			err := inFunc(a)
			if err != nil {
				return err
			}
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
