package chefapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-chef/chef"
)

// ChefAPI wraps the go-chef client and provides multi-organization support
type ChefAPI struct {
	BaseURL     string
	Name        string
	KeyMaterial string
	clients     map[string]*chef.Client // Cache clients per organization
}

// NewChefAPI initializes a ChefAPI client
// keyPathOrInline can be a filesystem path to the PEM private key or the inline PEM contents themselves.
// serverURL should be the base Chef server URL without organization path
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

	baseURL := ensureTrailingSlash(serverURL)

	return &ChefAPI{
		BaseURL:     baseURL,
		Name:        name,
		KeyMaterial: keyMaterial,
		clients:     make(map[string]*chef.Client),
	}, nil
}

// getClientForOrg returns a Chef client for the specified organization
func (api *ChefAPI) getClientForOrg(organization string) (*chef.Client, error) {
	if organization == "" {
		return nil, fmt.Errorf("organization cannot be empty")
	}

	// Check if we already have a client for this organization
	if client, exists := api.clients[organization]; exists {
		return client, nil
	}

	// Create new client for this organization
	orgURL := api.BaseURL + "organizations/" + organization + "/"
	client, err := chef.NewClient(&chef.Config{
		Name:    api.Name,
		Key:     api.KeyMaterial,
		BaseURL: orgURL,
	})
	if err != nil {
		return nil, fmt.Errorf("init chef client for org '%s': %w", organization, err)
	}

	// Cache the client
	api.clients[organization] = client
	return client, nil
}

// ListNodes returns a list of node names from the Chef server for the specified organization
func (api *ChefAPI) ListNodes(organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	nodesMap, err := client.Nodes.List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(nodesMap))
	for name := range nodesMap {
		names = append(names, name)
	}
	return names, nil
}

// GetNode returns a single node by name from the specified organization
func (api *ChefAPI) GetNode(name, organization string) (*chef.Node, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	n, err := client.Nodes.Get(name)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// ListRoles returns a slice of role names from the specified organization
func (api *ChefAPI) ListRoles(organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	rolesList, err := client.Roles.List()
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

// GetRole fetches a single role definition from the specified organization
func (api *ChefAPI) GetRole(name, organization string) (*chef.Role, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	return client.Roles.Get(name)
}

// ListUsers returns a slice of user names from the specified organization
func (api *ChefAPI) ListUsers(organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	usersMap, err := client.Users.List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(usersMap))
	for name := range usersMap {
		names = append(names, name)
	}
	return names, nil
}

// GetUser fetches a single user account from the specified organization
func (api *ChefAPI) GetUser(name, organization string) (*chef.User, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	u, err := client.Users.Get(name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Search executes a Chef search in the specified organization and returns the SearchResult
func (api *ChefAPI) Search(index, statement, organization string) (chef.SearchResult, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return chef.SearchResult{}, err
	}

	return client.Search.Exec(index, statement)
}

// SearchJSON executes a Chef search in the specified organization returning raw JSON rows
func (api *ChefAPI) SearchJSON(index, statement, organization string) (chef.JSearchResult, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return chef.JSearchResult{}, err
	}

	return client.Search.ExecJSON(index, statement)
}

// GetOrganization returns organization details for the specified organization
func (api *ChefAPI) GetOrganization(organization string) (*chef.Organization, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	org, err := client.Organizations.Get(organization)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// ListCookbooks returns a list of cookbook names and their versions from the specified organization
func (api *ChefAPI) ListCookbooks(organization string) (chef.CookbookListResult, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return chef.CookbookListResult{}, err
	}

	cookbooks, err := client.Cookbooks.List()
	if err != nil {
		return chef.CookbookListResult{}, err
	}
	return cookbooks, nil
}

// GetCookbook returns a cookbook with the specified version from the specified organization
// If version is empty or "_latest", it will get the latest version
func (api *ChefAPI) GetCookbook(name, version, organization string) (*chef.Cookbook, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	// Use "_latest" as default if version is empty
	if version == "" {
		version = "_latest"
	}

	cookbook, err := client.Cookbooks.GetVersion(name, version)
	if err != nil {
		return nil, err
	}
	return &cookbook, nil
}

// ListDataBags returns a list of data bag names from the specified organization
func (api *ChefAPI) ListDataBags(organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	dataBagsMap, err := client.DataBags.List()
	if err != nil {
		return nil, err
	}
	if dataBagsMap == nil {
		return []string{}, nil
	}
	names := make([]string, 0, len(*dataBagsMap))
	for name := range *dataBagsMap {
		names = append(names, name)
	}
	return names, nil
}

// ListDataBagItems returns a list of items in a data bag from the specified organization
func (api *ChefAPI) ListDataBagItems(name, organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	items, err := client.DataBags.ListItems(name)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []string{}, nil
	}
	names := make([]string, 0, len(*items))
	for itemName := range *items {
		names = append(names, itemName)
	}
	return names, nil
}

// GetDataBagItem returns a specific item from a data bag in the specified organization
func (api *ChefAPI) GetDataBagItem(bagName, itemName, organization string) (*chef.DataBagItem, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	item, err := client.DataBags.GetItem(bagName, itemName)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ListEnvironments returns a list of environment names from the specified organization
func (api *ChefAPI) ListEnvironments(organization string) ([]string, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	environmentsMap, err := client.Environments.List()
	if err != nil {
		return nil, err
	}
	if environmentsMap == nil {
		return []string{}, nil
	}
	names := make([]string, 0, len(*environmentsMap))
	for name := range *environmentsMap {
		names = append(names, name)
	}
	return names, nil
}

// GetEnvironment returns an environment definition from the specified organization
func (api *ChefAPI) GetEnvironment(name, organization string) (*chef.Environment, error) {
	client, err := api.getClientForOrg(organization)
	if err != nil {
		return nil, err
	}

	env, err := client.Environments.Get(name)
	if err != nil {
		return nil, err
	}
	return env, nil
}

// ensureTrailingSlash appends a slash if missing (so url.ResolveReference treats BaseURL as a directory path)
func ensureTrailingSlash(s string) string {
	if s == "" || strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}
