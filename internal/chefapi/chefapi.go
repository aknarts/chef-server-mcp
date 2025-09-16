package chefapi

import (
	"github.com/go-chef/chef"
)

// ChefAPI wraps the go-chef client
type ChefAPI struct {
	Client *chef.Client
}

// NewChefAPI initializes a ChefAPI client
func NewChefAPI(name, keyPath, serverURL string) (*ChefAPI, error) {
	client, err := chef.NewClient(&chef.Config{
		Name:    name,
		Key:     keyPath,
		BaseURL: serverURL,
	})
	if err != nil {
		return nil, err
	}
	return &ChefAPI{Client: client}, nil
}

// ListNodes returns a list of node names from the Chef server
func (api *ChefAPI) ListNodes() ([]string, error) {
	nodesMap, err := api.Client.Nodes.List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(nodesMap))
	for name := range nodesMap {
		names = append(names, name)
	}
	return names, nil
}
