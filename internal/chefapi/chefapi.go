package chefapi

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-chef/chef"
)

// ChefAPI wraps the go-chef client
type ChefAPI struct {
	Client  *chef.Client
	BaseURL string // cached sanitized base URL for logging/debug
}

// debug flag & helper
var debug bool

// SetDebug enables or disables verbose logging for this package.
func SetDebug(v bool) { debug = v }

func dlog(format string, args ...interface{}) {
	if debug {
		log.Printf("chefapi: "+format, args...)
	}
}

// NewChefAPI initializes a ChefAPI client
// keyPathOrInline can be a filesystem path to the PEM private key or the inline PEM contents themselves.
func NewChefAPI(name, keyPathOrInline, serverURL string) (*ChefAPI, error) {
	var keyMaterial string
	if strings.Contains(keyPathOrInline, "-----BEGIN") {
		// Looks like inline PEM content already.
		keyMaterial = keyPathOrInline
		if debug {
			dlog("using inline key material for user=%s", name)
		}
	} else {
		b, err := os.ReadFile(keyPathOrInline)
		if err != nil {
			return nil, fmt.Errorf("read key file '%s': %w", keyPathOrInline, err)
		}
		keyMaterial = string(b)
		if debug {
			dlog("loaded key file path=%s for user=%s", keyPathOrInline, name)
		}
	}

	serverURL = ensureTrailingSlash(serverURL)
	if debug {
		dlog("initializing go-chef client base_url=%s user=%s", serverURL, name)
	}
	client, err := chef.NewClient(&chef.Config{
		Name:    name,
		Key:     keyMaterial,
		BaseURL: serverURL,
	})
	if err != nil {
		return nil, fmt.Errorf("init chef client: %w", err)
	}
	if debug {
		dlog("client initialized base_url=%s", serverURL)
	}
	return &ChefAPI{Client: client, BaseURL: serverURL}, nil
}

// ListNodes returns a list of node names from the Chef server
func (api *ChefAPI) ListNodes() ([]string, error) {
	if debug {
		dlog("ListNodes start base_url=%s", api.BaseURL)
	}
	nodesMap, err := api.Client.Nodes.List()
	if err != nil {
		if debug {
			dlog("ListNodes error: %v", err)
		}
		return nil, err
	}
	names := make([]string, 0, len(nodesMap))
	for name := range nodesMap {
		names = append(names, name)
	}
	if debug {
		dlog("ListNodes success count=%d", len(names))
	}
	return names, nil
}

// GetNode returns a single node by name.
func (api *ChefAPI) GetNode(name string) (*chef.Node, error) {
	if debug {
		dlog("GetNode start name=%s", name)
	}
	n, err := api.Client.Nodes.Get(name)
	if err != nil {
		if debug {
			dlog("GetNode error name=%s err=%v", name, err)
		}
		return nil, err
	}
	if debug {
		dlog("GetNode success name=%s", name)
	}
	return &n, nil
}

// ListRoles returns a slice of role names.
func (api *ChefAPI) ListRoles() ([]string, error) {
	if debug {
		dlog("ListRoles start")
	}
	rolesList, err := api.Client.Roles.List()
	if err != nil {
		if debug {
			dlog("ListRoles error: %v", err)
		}
		return nil, err
	}
	if rolesList == nil {
		if debug {
			dlog("ListRoles success count=0 (nil list)")
		}
		return []string{}, nil
	}
	names := make([]string, 0, len(*rolesList))
	for name := range *rolesList {
		names = append(names, name)
	}
	if debug {
		dlog("ListRoles success count=%d", len(names))
	}
	return names, nil
}

// GetRole fetches a single role definition.
func (api *ChefAPI) GetRole(name string) (*chef.Role, error) {
	if debug {
		dlog("GetRole start name=%s", name)
	}
	r, err := api.Client.Roles.Get(name)
	if err != nil {
		if debug {
			dlog("GetRole error name=%s err=%v", name, err)
		}
		return nil, err
	}
	if debug {
		dlog("GetRole success name=%s", name)
	}
	return r, nil
}

// ListUsers returns a slice of user names.
func (api *ChefAPI) ListUsers() ([]string, error) {
	if debug {
		dlog("ListUsers start")
	}
	usersMap, err := api.Client.Users.List()
	if err != nil {
		if debug {
			dlog("ListUsers error: %v", err)
		}
		return nil, err
	}
	names := make([]string, 0, len(usersMap))
	for name := range usersMap {
		names = append(names, name)
	}
	if debug {
		dlog("ListUsers success count=%d", len(names))
	}
	return names, nil
}

// GetUser fetches a single user account.
func (api *ChefAPI) GetUser(name string) (*chef.User, error) {
	if debug {
		dlog("GetUser start name=%s", name)
	}
	u, err := api.Client.Users.Get(name)
	if err != nil {
		if debug {
			dlog("GetUser error name=%s err=%v", name, err)
		}
		return nil, err
	}
	if debug {
		dlog("GetUser success name=%s", name)
	}
	return &u, nil
}

// Search executes a Chef search (standard Exec path accumulating all pages) and returns the SearchResult.
func (api *ChefAPI) Search(index, statement string) (chef.SearchResult, error) {
	if debug {
		dlog("Search start index=%s query=%s", index, statement)
	}
	res, err := api.Client.Search.Exec(index, statement)
	if err != nil {
		if debug {
			dlog("Search error index=%s err=%v", index, err)
		}
		return res, err
	}
	if debug {
		dlog("Search success index=%s total=%d start=%d rows=%d", index, res.Total, res.Start, len(res.Rows))
	}
	return res, nil
}

// SearchJSON executes a Chef search returning raw JSON rows (JSearchResult).
func (api *ChefAPI) SearchJSON(index, statement string) (chef.JSearchResult, error) {
	if debug {
		dlog("SearchJSON start index=%s query=%s", index, statement)
	}
	res, err := api.Client.Search.ExecJSON(index, statement)
	if err != nil {
		if debug {
			dlog("SearchJSON error index=%s err=%v", index, err)
		}
		return res, err
	}
	if debug {
		dlog("SearchJSON success index=%s total=%d start=%d rows=%d", index, res.Total, res.Start, len(res.Rows))
	}
	return res, nil
}

// ensureTrailingSlash appends a slash if missing (so url.ResolveReference treats BaseURL as a directory path)
func ensureTrailingSlash(s string) string {
	if s == "" || strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}
