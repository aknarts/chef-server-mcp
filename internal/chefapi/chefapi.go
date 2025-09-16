package chefapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-chef/chef"
)

// ChefAPI wraps the go-chef client
type ChefAPI struct {
	Client *chef.Client
}

// NewChefAPI initializes a ChefAPI client
// keyPathOrInline can be a filesystem path to the PEM private key or the inline PEM contents themselves.
func NewChefAPI(name, keyPathOrInline, serverURL string) (*ChefAPI, error) {
	var keyMaterial string
	if strings.Contains(keyPathOrInline, "-----BEGIN") {
		// Looks like inline PEM content already.
		keyMaterial = keyPathOrInline
	} else {
		b, err := os.ReadFile(keyPathOrInline)
		if err != nil {
			return nil, fmt.Errorf("read key file '%s': %w", keyPathOrInline, err)
		}
		keyMaterial = string(b)
	}

	serverURL = ensureTrailingSlash(serverURL)

	client, err := chef.NewClient(&chef.Config{
		Name:    name,
		Key:     keyMaterial,
		BaseURL: serverURL,
	})
	if err != nil {
		return nil, fmt.Errorf("init chef client: %w", err)
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

// GetNode returns a single node by name.
func (api *ChefAPI) GetNode(name string) (*chef.Node, error) {
	n, err := api.Client.Nodes.Get(name)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// ListRoles returns a slice of role names.
func (api *ChefAPI) ListRoles() ([]string, error) {
	rolesList, err := api.Client.Roles.List()
	if err != nil {
		return nil, err
	}
	if rolesList == nil {
		return []string{}, nil
	}
	names := make([]string, 0, len(*rolesList))
	for name := range *rolesList {
		names = append(names, name)
	}
	return names, nil
}

// GetRole fetches a single role definition.
func (api *ChefAPI) GetRole(name string) (*chef.Role, error) {
	return api.Client.Roles.Get(name)
}

// ListUsers returns a slice of user names.
func (api *ChefAPI) ListUsers() ([]string, error) {
	usersMap, err := api.Client.Users.List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(usersMap))
	for name := range usersMap {
		names = append(names, name)
	}
	return names, nil
}

// GetUser fetches a single user account.
func (api *ChefAPI) GetUser(name string) (*chef.User, error) {
	u, err := api.Client.Users.Get(name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Search executes a Chef search (standard Exec path accumulating all pages) and returns the SearchResult.
func (api *ChefAPI) Search(index, statement string) (chef.SearchResult, error) {
	return api.Client.Search.Exec(index, statement)
}

// SearchJSON executes a Chef search returning raw JSON rows (JSearchResult).
func (api *ChefAPI) SearchJSON(index, statement string) (chef.JSearchResult, error) {
	return api.Client.Search.ExecJSON(index, statement)
}

// ensureTrailingSlash appends a slash if missing (so url.ResolveReference treats BaseURL as a directory path)
func ensureTrailingSlash(s string) string {
	if s == "" || strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}
