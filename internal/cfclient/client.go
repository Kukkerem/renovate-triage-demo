// Package cfclient builds a Cloud Foundry API client for the CF-brokered
// services layered on top of a BTP subaccount (this provider's real-world
// analogue talks to CF the same way it talks to the BTP REST APIs).
//
// The v3 client line has not had a stable release yet, so this package
// deliberately keeps its surface small: NewSpaceLister and ListSpaceNames
// are the only things the provider depends on.
package cfclient

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
)

// SpaceLister lists the Cloud Foundry space names visible to a client.
type SpaceLister func(ctx context.Context) ([]string, error)

// NewSpaceLister builds a Cloud Foundry API client scoped to apiRoot and
// returns a function that lists the space names visible to it.
//
// It is returned as a func value rather than invoked here so that this
// package compiles and links without ever dialing a real Cloud Foundry API —
// go-cfclient performs live UAA/API discovery inside config.New, which would
// make `go build`/`go test` network-dependent if we called it eagerly.
func NewSpaceLister(apiRoot, clientID, clientSecret string) (SpaceLister, error) {
	cfg, err := config.New(apiRoot, config.ClientCredentials(clientID, clientSecret))
	if err != nil {
		return nil, fmt.Errorf("cfclient: build config: %w", err)
	}
	cf, err := client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("cfclient: build client: %w", err)
	}
	return func(ctx context.Context) ([]string, error) {
		spaces, err := cf.Spaces.ListAll(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("cfclient: list spaces: %w", err)
		}
		names := make([]string, 0, len(spaces))
		for _, s := range spaces {
			names = append(names, s.Name)
		}
		return names, nil
	}, nil
}
