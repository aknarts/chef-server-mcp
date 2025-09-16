package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aknarts/chef-server-mcp/internal/chefapi"
	"github.com/aknarts/chef-server-mcp/internal/config"
	"github.com/aknarts/chef-server-mcp/internal/knife"
	"github.com/aknarts/chef-server-mcp/internal/nodes"
	"github.com/aknarts/chef-server-mcp/internal/version"
)

// --- JSON-RPC data structures ---

type jsonrpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params,omitempty"`
}

type toolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type initializeResult struct {
	Capabilities struct {
		Tools []toolDescriptor `json:"tools"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolCallSuccess struct {
	Content []contentItem `json:"content"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonrpcResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *jsonrpcError    `json:"error,omitempty"`
}

// --- Globals ---

var (
	cfg        *config.Config
	chefClient interface{ ListNodes() ([]string, error) }
)

// --- Main ---

func main() {
	log.SetOutput(os.Stderr)
	cfg = config.LoadFromEnv()
	if cfg.ChefUser != "" && cfg.ChefKeyPath != "" && cfg.ChefServerURL != "" {
		if c, err := chefapi.NewChefAPI(cfg.ChefUser, cfg.ChefKeyPath, cfg.ChefServerURL); err == nil {
			chefClient = c
		} else {
			log.Printf("warning: Chef API init failed: %v", err)
		}
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		msgBytes, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			writeError(nil, -32700, "read error: "+err.Error())
			return
		}
		processMessage(msgBytes)
	}
}

// --- Message Processing ---

func processMessage(raw []byte) {
	var req jsonrpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		writeError(nil, -32700, "parse error: "+err.Error())
		return
	}
	switch req.Method {
	case "initialize":
		res := buildInitializeResult()
		writeResponse(jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: res})
	case "tools/call":
		var p toolCallParams
		if req.Params != nil {
			_ = json.Unmarshal(*req.Params, &p)
		}
		if p.Name == "listNodes" {
			list, err := nodes.ListNodes(chefClient, cfg.KnifeFallback, runKnife)
			if err != nil {
				writeError(req.ID, 1, err.Error())
				return
			}
			var buf bytes.Buffer
			for i, n := range list {
				if i > 0 {
					buf.WriteByte('\n')
				}
				buf.WriteString(n)
			}
			writeResponse(jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: toolCallSuccess{Content: []contentItem{{Type: "text", Text: buf.String()}}}})
		} else {
			writeError(req.ID, -32601, "unknown tool: "+p.Name)
		}
	case "shutdown":
		writeResponse(jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]string{"status": "ok"}})
	case "exit":
		os.Exit(0)
	default:
		writeError(req.ID, -32601, "unknown method: "+req.Method)
	}
}

// --- Helpers ---

func buildInitializeResult() initializeResult {
	var r initializeResult
	r.ServerInfo.Name = "chef-server-mcp"
	r.ServerInfo.Version = version.Version
	r.Capabilities.Tools = []toolDescriptor{{
		Name:        "listNodes",
		Description: "List Chef node names (Chef API first, fallback to knife if enabled)",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`),
	}}
	return r
}

func writeResponse(resp jsonrpcResponse) {
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}

func writeError(id *json.RawMessage, code int, msg string) {
	writeResponse(jsonrpcResponse{JSONRPC: "2.0", ID: id, Error: &jsonrpcError{Code: code, Message: msg}})
}

// readMessage supports either newline-delimited JSON or Content-Length framing.
func readMessage(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(strings.ToLower(trim), "content-length:") {
		parts := strings.SplitN(trim, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid content-length header")
		}
		n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid content-length value: %w", err)
		}
		blank, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(blank) != "" {
			return nil, fmt.Errorf("expected blank line after Content-Length header")
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return buf, nil
	}
	return []byte(trim), nil
}

func runKnife(args ...string) (string, error) { return knife.RunKnifeCommand(args...) }
