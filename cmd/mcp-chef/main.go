package main

// Re-implemented using the official Model Context Protocol Go SDK.
// Knife fallback removed. All tools require valid Chef API credentials.

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

// ListNodesInput includes optional org alias.
type ListNodesInput struct {
	Org string `json:"org,omitempty"`
}

type ListNodesOutput struct {
	Nodes []string `json:"nodes" jsonschema:"List of Chef node names"`
}

// Additional tool input/output structs

type GetNodeInput struct {
	Name string `json:"name"`
	Org  string `json:"org,omitempty"`
}
type GetNodeOutput struct {
	Node *chef.Node `json:"node"`
}

type ListRolesInput struct {
	Org string `json:"org,omitempty"`
}
type ListRolesOutput struct {
	Roles []string `json:"roles"`
}

type GetRoleInput struct {
	Name string `json:"name"`
	Org  string `json:"org,omitempty"`
}
type GetRoleOutput struct {
	Role *chef.Role `json:"role"`
}

type ListUsersInput struct{}
type ListUsersOutput struct {
	Users []string `json:"users"` // server-scope (not org specific)
}

type GetUserInput struct {
	Name string `json:"name"`
}
type GetUserOutput struct {
	User *chef.User `json:"user"`
}

type SearchInput struct {
	Index string `json:"index"`
	Query string `json:"query"`
	Org   string `json:"org,omitempty"`
}

type SearchOutput struct {
	Total int           `json:"total"`
	Start int           `json:"start"`
	Rows  []interface{} `json:"rows"`
}

type SearchJSONInput struct {
	Index string `json:"index"`
	Query string `json:"query"`
	Org   string `json:"org,omitempty"`
}

type SearchJSONOutput struct {
	Total int               `json:"total"`
	Start int               `json:"start"`
	Rows  []json.RawMessage `json:"rows"`
}

// listOrgAliases tool
// Provides the configured alias -> organization mapping and default alias if any.
type ListOrgAliasesInput struct{}

type ListOrgAliasesOutput struct {
	Aliases map[string]string `json:"aliases"`
	Default string            `json:"default,omitempty"`
}

// listOrgClients tool (debug helper) - shows currently initialized org clients and their base URLs
type ListOrgClientsInput struct{}

type ListOrgClientsOutput struct {
	Clients map[string]string `json:"clients"`
}

func main() {
	log.SetOutput(os.Stderr)
	cfg := config.LoadFromEnv()
	if cfg.Debug {
		log.Printf("debug logging enabled (CHEF_DEBUG)")
	}
	chefapi.SetDebug(cfg.Debug)
	if err := cfg.Validate(); err != nil {
		log.Fatalf("configuration invalid: %v", err)
	}
	log.Printf("mcp-chef starting version=%s (multi-org support) root_url=%s aliases=%d default_alias=%s", version.Version, cfg.ChefServerURL, len(cfg.OrgAliases), cfg.DefaultOrgAlias)

	// Root/global client only used for user-level (server-scope) endpoints.
	rootClient, err := chefapi.NewChefAPI(cfg.ChefUser, cfg.ChefKeyPath, cfg.ChefServerURL)
	if err != nil {
		log.Fatalf("failed to init root Chef API client: %v", err)
	}
	log.Printf("root client initialized base_url=%s user=%s", cfg.ChefServerURL, cfg.ChefUser)

	orgMgr := chefapi.NewOrgClientManager(cfg.ChefServerURL, cfg.ChefUser, cfg.ChefKeyPath)

	resolveOrg := func(alias string) (string, error) {
		if alias == "" {
			if cfg.DefaultOrgAlias != "" {
				log.Printf("org alias empty; using default alias=%s", cfg.DefaultOrgAlias)
				alias = cfg.DefaultOrgAlias
			} else {
				return "", errors.New("org alias required (no default configured)")
			}
		}
		if len(cfg.OrgAliases) == 0 {
			// No alias mapping configured; treat alias as real org name directly.
			log.Printf("no alias map configured; treating provided alias='%s' as org name", alias)
			return alias, nil
		}
		org, ok := cfg.OrgAliases[alias]
		if !ok {
			return "", fmt.Errorf("unknown org alias '%s'", alias)
		}
		if alias != org {
			log.Printf("resolved org alias '%s' -> org '%s'", alias, org)
		} else {
			log.Printf("alias '%s' maps to same-named org", alias)
		}
		return org, nil
	}

	getOrgAPI := func(alias string) (*chefapi.ChefAPI, string, error) {
		org, err := resolveOrg(alias)
		if err != nil {
			log.Printf("getOrgAPI alias='%s' error=%v", alias, err)
			return nil, "", err
		}
		c, err := orgMgr.Get(org)
		if err != nil {
			log.Printf("getOrgAPI alias='%s' org='%s' creation_error=%v", alias, org, err)
			return nil, "", err
		}
		return c, org, nil
	}

	impl := &mcp.Implementation{Name: "chef-server-mcp", Version: version.Version}
	server := mcp.NewServer(impl, nil)

	// helper to wrap tool functions with logging
	wrap := func(name string, fn func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("tool call start name=%s input=%s", name, string(req.Arguments))
			res, err := fn(ctx, req)
			if err != nil {
				log.Printf("tool call error name=%s err=%v", name, err)
			} else {
				log.Printf("tool call success name=%s", name)
			}
			return res, err
		}
	}

	// listOrgAliases tool
	mcp.AddTool(server, &mcp.Tool{Name: "listOrgAliases", Description: "List configured Chef organization aliases and default"},
		wrap("listOrgAliases", func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{Content: []mcp.ToolOutputContent{{Type: "text", Text: fmt.Sprintf("aliases=%v default=%s", cfg.OrgAliases, cfg.DefaultOrgAlias)}}}, nil
		}))

	// listOrgClients tool (debug helper) - shows currently initialized org clients and their base URLs
	mcp.AddTypedTool(server, &mcp.Tool{Name: "listOrgClients", Description: "Show cached initialized organization clients (org -> baseURL)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListOrgClientsInput) (*mcp.CallToolResult, ListOrgClientsOutput, error) {
			return nil, ListOrgClientsOutput{Clients: orgMgr.ClientsSnapshot()}, nil
		})

	// listNodes tool (org-scoped)
	mcp.AddTypedTool(server, &mcp.Tool{
		Name:        "listNodes",
		Description: "List Chef node names within an organization (use 'org' alias or rely on default)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListNodesInput) (*mcp.CallToolResult, ListNodesOutput, error) {
		api, org, err := getOrgAPI(in.Org)
		if err != nil {
			return nil, ListNodesOutput{}, err
		}
		nodes, err := api.ListNodes()
		if err != nil {
			log.Printf("listNodes org=%s error=%v", org, err)
			return nil, ListNodesOutput{}, err
		}
		log.Printf("listNodes org=%s count=%d", org, len(nodes))
		return nil, ListNodesOutput{Nodes: nodes}, nil
	})

	// getNode
	mcp.AddTypedTool(server, &mcp.Tool{Name: "getNode", Description: "Get a single Chef node by name (org-scoped)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetNodeInput) (*mcp.CallToolResult, GetNodeOutput, error) {
			api, org, err := getOrgAPI(in.Org)
			if err != nil {
				return nil, GetNodeOutput{}, err
			}
			n, err := api.GetNode(in.Name)
			if err != nil {
				log.Printf("getNode org=%s name=%s error=%v", org, in.Name, err)
				return nil, GetNodeOutput{}, err
			}
			log.Printf("getNode org=%s name=%s", org, in.Name)
			return nil, GetNodeOutput{Node: n}, nil
		})

	// listRoles
	mcp.AddTypedTool(server, &mcp.Tool{Name: "listRoles", Description: "List Chef role names (org-scoped)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListRolesInput) (*mcp.CallToolResult, ListRolesOutput, error) {
			api, org, err := getOrgAPI(in.Org)
			if err != nil {
				return nil, ListRolesOutput{}, err
			}
			roles, err := api.ListRoles()
			if err != nil {
				log.Printf("listRoles org=%s error=%v", org, err)
				return nil, ListRolesOutput{}, err
			}
			log.Printf("listRoles org=%s count=%d", org, len(roles))
			return nil, ListRolesOutput{Roles: roles}, nil
		})

	// getRole
	mcp.AddTypedTool(server, &mcp.Tool{Name: "getRole", Description: "Get a single Chef role by name (org-scoped)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetRoleInput) (*mcp.CallToolResult, GetRoleOutput, error) {
			api, org, err := getOrgAPI(in.Org)
			if err != nil {
				return nil, GetRoleOutput{}, err
			}
			r, err := api.GetRole(in.Name)
			if err != nil {
				log.Printf("getRole org=%s name=%s error=%v", org, in.Name, err)
				return nil, GetRoleOutput{}, err
			}
			log.Printf("getRole org=%s name=%s", org, in.Name)
			return nil, GetRoleOutput{Role: r}, nil
		})

	// listUsers (server-scope)
	mcp.AddTypedTool(server, &mcp.Tool{Name: "listUsers", Description: "List Chef user names (server-scope)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListUsersInput) (*mcp.CallToolResult, ListUsersOutput, error) {
			users, err := rootClient.ListUsers()
			if err != nil {
				log.Printf("listUsers error=%v", err)
				return nil, ListUsersOutput{}, err
			}
			log.Printf("listUsers count=%d", len(users))
			return nil, ListUsersOutput{Users: users}, nil
		})

	// getUser (server-scope)
	mcp.AddTypedTool(server, &mcp.Tool{Name: "getUser", Description: "Get a single Chef user by name (server-scope)"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetUserInput) (*mcp.CallToolResult, GetUserOutput, error) {
			u, err := rootClient.GetUser(in.Name)
			if err != nil {
				log.Printf("getUser name=%s error=%v", in.Name, err)
				return nil, GetUserOutput{}, err
			}
			log.Printf("getUser name=%s", in.Name)
			return nil, GetUserOutput{User: u}, nil
		})

	// search (org scope)
	mcp.AddTypedTool(server, &mcp.Tool{Name: "search", Description: "Execute a Chef search (org-scoped) and return decoded rows"},
		func(ctx context.Context, req *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
			api, org, err := getOrgAPI(in.Org)
			if err != nil {
				return nil, SearchOutput{}, err
			}
			res, err := api.Search(in.Index, in.Query)
			if err != nil {
				log.Printf("search org=%s index=%s query=%q error=%v", org, in.Index, in.Query, err)
				return nil, SearchOutput{}, err
			}
			log.Printf("search org=%s index=%s query=%q total=%d rows=%d", org, in.Index, in.Query, res.Total, len(res.Rows))
			return nil, SearchOutput{Total: res.Total, Start: res.Start, Rows: res.Rows}, nil
		})

	// searchJSON (org scope)
	mcp.AddTypedTool(server, &mcp.Tool{Name: "searchJSON", Description: "Execute a Chef search (org-scoped) and return raw JSON rows"},
		func(ctx context.Context, req *mcp.CallToolRequest, in SearchJSONInput) (*mcp.CallToolResult, SearchJSONOutput, error) {
			api, org, err := getOrgAPI(in.Org)
			if err != nil {
				return nil, SearchJSONOutput{}, err
			}
			res, err := api.SearchJSON(in.Index, in.Query)
			if err != nil {
				log.Printf("searchJSON org=%s index=%s query=%q error=%v", org, in.Index, in.Query, err)
				return nil, SearchJSONOutput{}, err
			}
			rows := make([]json.RawMessage, 0, len(res.Rows))
			for _, r := range res.Rows { // marshal each row back to raw JSON blob
				b, err := json.Marshal(r)
				if err != nil {
					log.Printf("searchJSON marshal row error org=%s index=%s err=%v", org, in.Index, err)
					return nil, SearchJSONOutput{}, err
				}
				rows = append(rows, b)
			}
			log.Printf("searchJSON org=%s index=%s query=%q total=%d rows=%d", org, in.Index, in.Query, res.Total, len(rows))
			return nil, SearchJSONOutput{Total: res.Total, Start: res.Start, Rows: rows}, nil
		})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Allow graceful cancel on SIGINT/SIGTERM while letting host manage stdio session.
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
