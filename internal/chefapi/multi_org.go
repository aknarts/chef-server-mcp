package chefapi

import (
	"fmt"
	"strings"
	"sync"
)

// OrgClientManager lazily creates and caches ChefAPI clients per organization.
type OrgClientManager struct {
	rootURL         string
	user            string
	keyPathOrInline string
	mu              sync.Mutex
	clients         map[string]*ChefAPI // orgName -> client
}

// NewOrgClientManager constructs a manager for a root Chef server URL (no /organizations/<org> segment).
func NewOrgClientManager(rootURL, user, keyPathOrInline string) *OrgClientManager {
	rootURL = strings.TrimSuffix(rootURL, "/")
	return &OrgClientManager{rootURL: rootURL, user: user, keyPathOrInline: keyPathOrInline, clients: make(map[string]*ChefAPI)}
}

// Get returns a ChefAPI client scoped to the given organization name (not alias).
func (m *OrgClientManager) Get(org string) (*ChefAPI, error) {
	if org == "" {
		if debug {
			dlog("OrgClientManager Get error: empty org")
		}
		return nil, fmt.Errorf("organization name required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.clients[org]; ok {
		if debug {
			dlog("OrgClientManager cache hit org=%s", org)
		}
		return c, nil
	}
	orgURL := fmt.Sprintf("%s/organizations/%s", m.rootURL, org)
	if debug {
		dlog("OrgClientManager creating client org=%s org_url=%s root_url=%s", org, orgURL, m.rootURL)
	}
	client, err := NewChefAPI(m.user, m.keyPathOrInline, orgURL)
	if err != nil {
		if debug {
			dlog("OrgClientManager client creation failed org=%s err=%v", org, err)
		}
		return nil, err
	}
	m.clients[org] = client
	return client, nil
}

// ClientsSnapshot returns a copy of the currently cached organization -> baseURL mapping.
func (m *OrgClientManager) ClientsSnapshot() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make(map[string]string, len(m.clients))
	for org, c := range m.clients {
		if c != nil {
			res[org] = c.BaseURL
		} else {
			res[org] = "<nil>"
		}
	}
	return res
}
