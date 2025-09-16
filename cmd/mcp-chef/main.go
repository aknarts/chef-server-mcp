package main

// Re-implemented using the official Model Context Protocol Go SDK.
// Knife fallback removed. All tools require valid Chef API credentials.

import (
	"context"
	"encoding/json"
	"errors"
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

// ListNodesInput intentionally empty (no arguments needed)
type ListNodesInput struct{}

// ListNodesOutput provides a structured response for the tool.
type ListNodesOutput struct {
	Nodes []string `json:"nodes" jsonschema:"List of Chef node names"`
}

// Additional tool input/output structs

type GetNodeInput struct {
	Name string `json:"name"`
}
type GetNodeOutput struct {
	Node *chef.Node `json:"node"`
}

type ListRolesInput struct{}
type ListRolesOutput struct {
	Roles []string `json:"roles"`
}

type GetRoleInput struct {
	Name string `json:"name"`
}
type GetRoleOutput struct {
	Role *chef.Role `json:"role"`
}

type ListUsersInput struct{}
type ListUsersOutput struct {
	Users []string `json:"users"`
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
}

type SearchOutput struct {
	Total int           `json:"total"`
	Start int           `json:"start"`
	Rows  []interface{} `json:"rows"`
}

type SearchJSONInput struct {
	Index string `json:"index"`
	Query string `json:"query"`
}

type SearchJSONOutput struct {
	Total int               `json:"total"`
	Start int               `json:"start"`
	Rows  []json.RawMessage `json:"rows"`
}

func main() {
	log.SetOutput(os.Stderr)
	cfg := config.LoadFromEnv()
	log.Printf("mcp-chef starting version=%s (knife fallback removed)", version.Version)

	// Require Chef API credentials now (no fallback mode)
	if cfg.ChefUser == "" || cfg.ChefKeyPath == "" || cfg.ChefServerURL == "" {
		log.Fatalf("Chef API credentials incomplete: need CHEF_USER, CHEF_KEY_PATH, CHEF_SERVER_URL")
	}

	chefClient, err := chefapi.NewChefAPI(cfg.ChefUser, cfg.ChefKeyPath, cfg.ChefServerURL)
	if err != nil {
		log.Fatalf("failed to init Chef API client: %v", err)
	}

	impl := &mcp.Implementation{Name: "chef-server-mcp", Version: version.Version}
	server := mcp.NewServer(impl, nil)

	needAPI := func() (*chefapi.ChefAPI, error) {
		if chefClient == nil {
			return nil, errors.New("Chef API client not initialized")
		}
		return chefClient, nil
	}

	// listNodes tool (API only)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listNodes",
		Description: "List Chef node names (Chef API)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in ListNodesInput) (*mcp.CallToolResult, ListNodesOutput, error) {
		api, err := needAPI()
		if err != nil {
			return nil, ListNodesOutput{}, err
		}
		nodes, err := api.ListNodes()
		if err != nil {
			return nil, ListNodesOutput{}, err
		}
		return nil, ListNodesOutput{Nodes: nodes}, nil
	})

	// getNode
	mcp.AddTool(server, &mcp.Tool{Name: "getNode", Description: "Get a single Chef node by name"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetNodeInput) (*mcp.CallToolResult, GetNodeOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, GetNodeOutput{}, err
			}
			n, err := api.GetNode(in.Name)
			if err != nil {
				return nil, GetNodeOutput{}, err
			}
			return nil, GetNodeOutput{Node: n}, nil
		})

	// listRoles
	mcp.AddTool(server, &mcp.Tool{Name: "listRoles", Description: "List Chef role names"},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListRolesInput) (*mcp.CallToolResult, ListRolesOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, ListRolesOutput{}, err
			}
			roles, err := api.ListRoles()
			if err != nil {
				return nil, ListRolesOutput{}, err
			}
			return nil, ListRolesOutput{Roles: roles}, nil
		})

	// getRole
	mcp.AddTool(server, &mcp.Tool{Name: "getRole", Description: "Get a single Chef role by name"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetRoleInput) (*mcp.CallToolResult, GetRoleOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, GetRoleOutput{}, err
			}
			r, err := api.GetRole(in.Name)
			if err != nil {
				return nil, GetRoleOutput{}, err
			}
			return nil, GetRoleOutput{Role: r}, nil
		})

	// listUsers
	mcp.AddTool(server, &mcp.Tool{Name: "listUsers", Description: "List Chef user names"},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListUsersInput) (*mcp.CallToolResult, ListUsersOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, ListUsersOutput{}, err
			}
			users, err := api.ListUsers()
			if err != nil {
				return nil, ListUsersOutput{}, err
			}
			return nil, ListUsersOutput{Users: users}, nil
		})

	// getUser
	mcp.AddTool(server, &mcp.Tool{Name: "getUser", Description: "Get a single Chef user by name"},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetUserInput) (*mcp.CallToolResult, GetUserOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, GetUserOutput{}, err
			}
			u, err := api.GetUser(in.Name)
			if err != nil {
				return nil, GetUserOutput{}, err
			}
			return nil, GetUserOutput{User: u}, nil
		})

	// search
	mcp.AddTool(server, &mcp.Tool{Name: "search", Description: "Execute a Chef search and return decoded rows"},
		func(ctx context.Context, req *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, SearchOutput{}, err
			}
			res, err := api.Search(in.Index, in.Query)
			if err != nil {
				return nil, SearchOutput{}, err
			}
			return nil, SearchOutput{Total: res.Total, Start: res.Start, Rows: res.Rows}, nil
		})

	// searchJSON
	mcp.AddTool(server, &mcp.Tool{Name: "searchJSON", Description: "Execute a Chef search and return raw JSON rows"},
		func(ctx context.Context, req *mcp.CallToolRequest, in SearchJSONInput) (*mcp.CallToolResult, SearchJSONOutput, error) {
			api, err := needAPI()
			if err != nil {
				return nil, SearchJSONOutput{}, err
			}
			res, err := api.SearchJSON(in.Index, in.Query)
			if err != nil {
				return nil, SearchJSONOutput{}, err
			}
			rows := make([]json.RawMessage, 0, len(res.Rows))
			for _, r := range res.Rows { // marshal each row back to raw JSON blob
				b, err := json.Marshal(r)
				if err != nil {
					return nil, SearchJSONOutput{}, err
				}
				rows = append(rows, b)
			}
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
