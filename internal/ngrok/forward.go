package forward

import (
	"context"
	"errors"
	"fmt"
	"hafh-server/internal/logger"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"golang.ngrok.com/ngrok"
	ngrokConfig "golang.ngrok.com/ngrok/config"
)

// ForwarderConfig is the configuration for the Forwarder.
type ForwarderConfig struct {
	BackendUrl string
	DomainUrl  string
	AuthToken  string
	Region     string
}

// Forwarder is the struct that handles the ngrok forwarding.
type Forwarder struct {
	backendUrl *url.URL
	domain     string
	authToken  string
	region     string
	log        *zap.SugaredLogger
}

// New creates a new Forwarder instance.
func New(config *ForwarderConfig) (*Forwarder, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	} else if config.BackendUrl == "" {
		return nil, errors.New("backend url cannot be empty")
	} else if config.DomainUrl == "" {
		return nil, errors.New("domain url cannot be empty")
	} else if config.AuthToken == "" {
		return nil, errors.New("auth token cannot be empty")
	} else if config.Region == "" {
		return nil, errors.New("region cannot be empty")
	}

	backend := config.BackendUrl
	if !strings.Contains(backend, "://") {
		backend = fmt.Sprintf("tcp://%s", backend)
	}

	backendUrl, err := url.Parse(backend)
	if err != nil {
		return nil, errors.New("invalid backend url: " + err.Error())
	}

	return &Forwarder{
		backendUrl: backendUrl,
		authToken:  config.AuthToken,
		domain:     config.DomainUrl,
		log:        logger.Named("forwarder"),
	}, nil
}

// Start starts the ngrok forwarding.
func (f *Forwarder) Start(ctx context.Context) error {
	session, err := ngrok.Connect(ctx,
		ngrok.WithAuthtoken(f.authToken),
		ngrok.WithRegion(f.region),
	)
	if err != nil {
		return err
	}

	for {
		fwd, err := session.ListenAndForward(ctx,
			f.backendUrl,
			ngrokConfig.HTTPEndpoint(ngrokConfig.WithDomain(f.domain)),
		)
		if err != nil {
			return err
		}

		f.log.Info("Ingress established", map[string]any{
			"url": fwd.URL(),
		})

		err = fwd.Wait()
		if err == nil {
			return nil
		}

		f.log.Warn("Accept error. now setting up a new forwarder.",
			map[string]any{"err": err})
	}
}
