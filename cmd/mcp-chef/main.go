package main

// MCP server providing read-only access to Chef Server resources.
// All tools are enforced read-only at multiple layers:
//   1. MCP ToolAnnotations (ReadOnlyHint: true) — signals to clients
//   2. ChefReader interface — compile-time enforcement in chefapi
//   3. ReadOnlyTransport — runtime HTTP method blocking in safeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chef/chef"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aknarts/chef-server-mcp/internal/chefapi"
	"github.com/aknarts/chef-server-mcp/internal/config"
	"github.com/aknarts/chef-server-mcp/internal/version"
)

// readOnlyAnnotations returns MCP tool annotations indicating a read-only,
// non-destructive, open-world (network-accessing) tool.
func readOnlyAnnotations() *mcp.ToolAnnotations {
	destructive := false
	openWorld := true
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		DestructiveHint: &destructive,
		IdempotentHint:  true,
		OpenWorldHint:   &openWorld,
	}
}

// Tool input/output types — all tools accept optional organization parameter

type ListNodesInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListNodesOutput struct {
	Nodes        []string `json:"nodes"`
	Organization string   `json:"organization"`
}

type GetNodeInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type GetNodeOutput struct {
	Node         *chef.Node `json:"node"`
	Organization string     `json:"organization"`
}

type ListRolesInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListRolesOutput struct {
	Roles        []string `json:"roles"`
	Organization string   `json:"organization"`
}

type GetRoleInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type GetRoleOutput struct {
	Role         *chef.Role `json:"role"`
	Organization string     `json:"organization"`
}

type ListUsersInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListUsersOutput struct {
	Users        []string `json:"users"`
	Organization string   `json:"organization"`
}

type GetUserInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type GetUserOutput struct {
	User         *chef.User `json:"user"`
	Organization string     `json:"organization"`
}

type SearchInput struct {
	Index        string  `json:"index"`
	Query        string  `json:"query"`
	Organization *string `json:"organization,omitempty"`
}
type SearchOutput struct {
	Total        int           `json:"total"`
	Start        int           `json:"start"`
	Rows         []interface{} `json:"rows"`
	Organization string        `json:"organization"`
}

type SearchJSONInput struct {
	Index        string  `json:"index"`
	Query        string  `json:"query"`
	Organization *string `json:"organization,omitempty"`
}
type SearchJSONOutput struct {
	Total        int               `json:"total"`
	Start        int               `json:"start"`
	Rows         []json.RawMessage `json:"rows"`
	Organization string            `json:"organization"`
}

type GetOrganizationInput struct {
	Organization *string `json:"organization,omitempty"`
}
type GetOrganizationOutput struct {
	Organization *chef.Organization `json:"organization"`
	OrgName      string             `json:"orgName"`
}

type ListCookbooksInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListCookbooksOutput struct {
	Cookbooks    chef.CookbookListResult `json:"cookbooks"`
	Organization string                  `json:"organization"`
}

type GetCookbookInput struct {
	Name         string  `json:"name"`
	Version      *string `json:"version,omitempty"`
	Organization *string `json:"organization,omitempty"`
}
type GetCookbookOutput struct {
	Cookbook      *chef.Cookbook `json:"cookbook"`
	Organization string        `json:"organization"`
}

type ListDataBagsInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListDataBagsOutput struct {
	DataBags     []string `json:"dataBags"`
	Organization string   `json:"organization"`
}

type ListDataBagItemsInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type ListDataBagItemsOutput struct {
	Items        []string `json:"items"`
	DataBag      string   `json:"dataBag"`
	Organization string   `json:"organization"`
}

type GetDataBagItemInput struct {
	BagName      string  `json:"bagName"`
	ItemName     string  `json:"itemName"`
	Organization *string `json:"organization,omitempty"`
}
type GetDataBagItemOutput struct {
	Item         *chef.DataBagItem `json:"item"`
	DataBag      string            `json:"dataBag"`
	Organization string            `json:"organization"`
}

type ListEnvironmentsInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListEnvironmentsOutput struct {
	Environments []string `json:"environments"`
	Organization string   `json:"organization"`
}

type GetEnvironmentInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type GetEnvironmentOutput struct {
	Environment  *chef.Environment `json:"environment"`
	Organization string            `json:"organization"`
}

type ListClientsInput struct {
	Organization *string `json:"organization,omitempty"`
}
type ListClientsOutput struct {
	Clients      []string `json:"clients"`
	Organization string   `json:"organization"`
}

type GetClientInput struct {
	Name         string  `json:"name"`
	Organization *string `json:"organization,omitempty"`
}
type GetClientOutput struct {
	Client       *chef.ApiClient `json:"client"`
	Organization string          `json:"organization"`
}

func main() {
	log.SetOutput(os.Stderr)
	cfg := config.LoadFromEnv()
	log.Printf("mcp-chef starting version=%s (read-only enforced)", version.Version)

	// Require Chef API credentials (no fallback mode)
	if cfg.ChefUser == "" || cfg.ChefKeyPath == "" || cfg.ChefServerURL == "" {
		log.Fatalf("Chef API credentials incomplete: need CHEF_USER, CHEF_KEY_PATH, CHEF_SERVER_URL")
	}

	if cfg.DefaultOrg == "" {
		log.Printf("Warning: CHEF_DEFAULT_ORG not set. Organization must be specified in each request.")
	}

	chefClient, err := chefapi.NewChefAPI(cfg.ChefUser, cfg.ChefKeyPath, cfg.ChefServerURL)
	if err != nil {
		log.Fatalf("failed to init Chef API client: %v", err)
	}

	impl := &mcp.Implementation{Name: "chef-server-mcp", Version: version.Version}
	server := mcp.NewServer(impl, nil)

	// Use ChefReader interface to enforce read-only at compile time
	var api chefapi.ChefReader = chefClient

	needAPI := func() (chefapi.ChefReader, error) {
		if api == nil {
			return nil, errors.New("Chef API client not initialized")
		}
		return api, nil
	}

	getOrgString := func(orgPtr *string) string {
		if orgPtr == nil {
			return ""
		}
		return *orgPtr
	}

	annotations := readOnlyAnnotations()

	// listNodes
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listNodes",
		Description: "List Chef node names - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListNodesInput) (*mcp.CallToolResult, ListNodesOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListNodesOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListNodesOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		nodes, err := a.ListNodes(org)
		if err != nil {
			return nil, ListNodesOutput{}, err
		}
		return nil, ListNodesOutput{Nodes: nodes, Organization: org}, nil
	})

	// getNode
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getNode",
		Description: "Get a single Chef node by name - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetNodeInput) (*mcp.CallToolResult, GetNodeOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetNodeOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetNodeOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		n, err := a.GetNode(in.Name, org)
		if err != nil {
			return nil, GetNodeOutput{}, err
		}
		return nil, GetNodeOutput{Node: n, Organization: org}, nil
	})

	// listRoles
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listRoles",
		Description: "List Chef role names - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListRolesInput) (*mcp.CallToolResult, ListRolesOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListRolesOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListRolesOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		roles, err := a.ListRoles(org)
		if err != nil {
			return nil, ListRolesOutput{}, err
		}
		return nil, ListRolesOutput{Roles: roles, Organization: org}, nil
	})

	// getRole
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getRole",
		Description: "Get a single Chef role by name - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetRoleInput) (*mcp.CallToolResult, GetRoleOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetRoleOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetRoleOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		r, err := a.GetRole(in.Name, org)
		if err != nil {
			return nil, GetRoleOutput{}, err
		}
		return nil, GetRoleOutput{Role: r, Organization: org}, nil
	})

	// listUsers
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listUsers",
		Description: "List Chef user names - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListUsersInput) (*mcp.CallToolResult, ListUsersOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListUsersOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListUsersOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		users, err := a.ListUsers(org)
		if err != nil {
			return nil, ListUsersOutput{}, err
		}
		return nil, ListUsersOutput{Users: users, Organization: org}, nil
	})

	// getUser
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getUser",
		Description: "Get a single Chef user by name - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetUserInput) (*mcp.CallToolResult, GetUserOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetUserOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetUserOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		u, err := a.GetUser(in.Name, org)
		if err != nil {
			return nil, GetUserOutput{}, err
		}
		return nil, GetUserOutput{User: u, Organization: org}, nil
	})

	// search
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Execute a Chef search and return decoded rows - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, SearchOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, SearchOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		res, err := a.Search(in.Index, in.Query, org)
		if err != nil {
			return nil, SearchOutput{}, err
		}
		return nil, SearchOutput{Total: res.Total, Start: res.Start, Rows: res.Rows, Organization: org}, nil
	})

	// searchJSON
	mcp.AddTool(server, &mcp.Tool{
		Name:        "searchJSON",
		Description: "Execute a Chef search and return raw JSON rows - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in SearchJSONInput) (*mcp.CallToolResult, SearchJSONOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, SearchJSONOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, SearchJSONOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		res, err := a.SearchJSON(in.Index, in.Query, org)
		if err != nil {
			return nil, SearchJSONOutput{}, err
		}
		rows := make([]json.RawMessage, 0, len(res.Rows))
		for _, r := range res.Rows {
			b, err := json.Marshal(r)
			if err != nil {
				return nil, SearchJSONOutput{}, err
			}
			rows = append(rows, b)
		}
		return nil, SearchJSONOutput{Total: res.Total, Start: res.Start, Rows: rows, Organization: org}, nil
	})

	// getOrganization
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getOrganization",
		Description: "Get organization details - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetOrganizationInput) (*mcp.CallToolResult, GetOrganizationOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetOrganizationOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetOrganizationOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		orgDetails, err := a.GetOrganization(org)
		if err != nil {
			return nil, GetOrganizationOutput{}, err
		}
		return nil, GetOrganizationOutput{Organization: orgDetails, OrgName: org}, nil
	})

	// listCookbooks
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listCookbooks",
		Description: "List Chef cookbooks and their versions - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListCookbooksInput) (*mcp.CallToolResult, ListCookbooksOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListCookbooksOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListCookbooksOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		cookbooks, err := a.ListCookbooks(org)
		if err != nil {
			return nil, ListCookbooksOutput{}, err
		}
		return nil, ListCookbooksOutput{Cookbooks: cookbooks, Organization: org}, nil
	})

	// getCookbook
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getCookbook",
		Description: "Get a Chef cookbook by name and version (defaults to _latest) - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetCookbookInput) (*mcp.CallToolResult, GetCookbookOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetCookbookOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetCookbookOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		ver := "_latest"
		if in.Version != nil && *in.Version != "" {
			ver = *in.Version
		}
		cookbook, err := a.GetCookbook(in.Name, ver, org)
		if err != nil {
			return nil, GetCookbookOutput{}, err
		}
		return nil, GetCookbookOutput{Cookbook: cookbook, Organization: org}, nil
	})

	// listDataBags
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listDataBags",
		Description: "List Chef data bags - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListDataBagsInput) (*mcp.CallToolResult, ListDataBagsOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListDataBagsOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListDataBagsOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		dataBags, err := a.ListDataBags(org)
		if err != nil {
			return nil, ListDataBagsOutput{}, err
		}
		return nil, ListDataBagsOutput{DataBags: dataBags, Organization: org}, nil
	})

	// listDataBagItems
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listDataBagItems",
		Description: "List items in a Chef data bag - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListDataBagItemsInput) (*mcp.CallToolResult, ListDataBagItemsOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListDataBagItemsOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListDataBagItemsOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		items, err := a.ListDataBagItems(in.Name, org)
		if err != nil {
			return nil, ListDataBagItemsOutput{}, err
		}
		return nil, ListDataBagItemsOutput{Items: items, DataBag: in.Name, Organization: org}, nil
	})

	// getDataBagItem
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDataBagItem",
		Description: "Get a specific item from a Chef data bag - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetDataBagItemInput) (*mcp.CallToolResult, GetDataBagItemOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetDataBagItemOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetDataBagItemOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		item, err := a.GetDataBagItem(in.BagName, in.ItemName, org)
		if err != nil {
			return nil, GetDataBagItemOutput{}, err
		}
		return nil, GetDataBagItemOutput{Item: item, DataBag: in.BagName, Organization: org}, nil
	})

	// listEnvironments
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listEnvironments",
		Description: "List Chef environments - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListEnvironmentsInput) (*mcp.CallToolResult, ListEnvironmentsOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListEnvironmentsOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListEnvironmentsOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		environments, err := a.ListEnvironments(org)
		if err != nil {
			return nil, ListEnvironmentsOutput{}, err
		}
		return nil, ListEnvironmentsOutput{Environments: environments, Organization: org}, nil
	})

	// getEnvironment
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getEnvironment",
		Description: "Get a Chef environment by name - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetEnvironmentInput) (*mcp.CallToolResult, GetEnvironmentOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetEnvironmentOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetEnvironmentOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		environment, err := a.GetEnvironment(in.Name, org)
		if err != nil {
			return nil, GetEnvironmentOutput{}, err
		}
		return nil, GetEnvironmentOutput{Environment: environment, Organization: org}, nil
	})

	// listClients
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listClients",
		Description: "List Chef API client names - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListClientsInput) (*mcp.CallToolResult, ListClientsOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, ListClientsOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, ListClientsOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		clients, err := a.ListClients(org)
		if err != nil {
			return nil, ListClientsOutput{}, err
		}
		return nil, ListClientsOutput{Clients: clients, Organization: org}, nil
	})

	// getClient
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getClient",
		Description: "Get a Chef API client by name - optionally specify organization",
		Annotations: annotations,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GetClientInput) (*mcp.CallToolResult, GetClientOutput, error) {
		a, err := needAPI()
		if err != nil {
			return nil, GetClientOutput{}, err
		}
		org := cfg.ResolveOrganization(getOrgString(in.Organization))
		if org == "" {
			return nil, GetClientOutput{}, fmt.Errorf("organization must be specified or CHEF_DEFAULT_ORG must be set")
		}
		c, err := a.GetClient(in.Name, org)
		if err != nil {
			return nil, GetClientOutput{}, err
		}
		return nil, GetClientOutput{Client: c, Organization: org}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		s := <-ch
		log.Printf("signal received: %s - cancelling MCP server context", s)
		cancel()
	}()

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Printf("mcp server stopped with error: %v", err)
	} else {
		log.Printf("mcp server stopped (EOF)")
	}
}
