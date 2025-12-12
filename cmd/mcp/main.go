// M3M MCP Server - Model Context Protocol server for M3M JavaScript API documentation
// This server provides Claude with access to M3M runtime module documentation.
//
// Usage:
//
//	# Stdio mode (for Claude Code CLI)
//	claude mcp add m3m-api ./m3m-mcp
//
//	# HTTP/SSE mode (for web clients)
//	./m3m-mcp --http :3100
//
// The server provides tools to search and retrieve M3M JavaScript API documentation.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"
	"time"

	"github.com/levskiy0/m3m/internal/runtime/modules"
	pkgplugin "github.com/levskiy0/m3m/pkg/plugin"
	"github.com/levskiy0/m3m/pkg/schema"
)

// JSON-RPC types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP types
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

type ToolsCapability struct{}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Module cache
var moduleSchemas []schema.ModuleSchema

// HTTP mode variables
var (
	httpAddr    string
	pluginsPath string
	sseClients  = make(map[string]chan string)
	sseMutex    sync.RWMutex
)

func init() {
	// Built-in modules are loaded first
	moduleSchemas = modules.GetAllSchemas()
}

// loadPluginSchemas loads .so plugin files and extracts their schemas
func loadPluginSchemas(pluginsDir string) []schema.ModuleSchema {
	var schemas []schema.ModuleSchema

	if pluginsDir == "" {
		return schemas
	}

	// Check if directory exists
	info, err := os.Stat(pluginsDir)
	if err != nil || !info.IsDir() {
		return schemas
	}

	// Scan for .so files
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to read plugins directory: %v\n", err)
		return schemas
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(pluginsDir, entry.Name())
		s, err := loadPluginSchema(pluginPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", entry.Name(), err)
			continue
		}

		schemas = append(schemas, s)
		fmt.Fprintf(os.Stderr, "Loaded plugin: %s\n", s.Name)
	}

	return schemas
}

// loadPluginSchema loads a single .so plugin and extracts its schema
func loadPluginSchema(path string) (schema.ModuleSchema, error) {
	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return schema.ModuleSchema{}, fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for NewPlugin function
	newPluginSym, err := p.Lookup("NewPlugin")
	if err != nil {
		return schema.ModuleSchema{}, fmt.Errorf("plugin does not export NewPlugin: %w", err)
	}

	// Cast to function
	newPluginFunc, ok := newPluginSym.(func() interface{})
	if !ok {
		return schema.ModuleSchema{}, fmt.Errorf("NewPlugin has wrong signature")
	}

	// Create plugin instance and cast to Plugin interface
	rawPlugin := newPluginFunc()
	pluginInstance, ok := rawPlugin.(pkgplugin.Plugin)
	if !ok {
		return schema.ModuleSchema{}, fmt.Errorf("plugin does not implement Plugin interface")
	}

	// Initialize plugin with empty config (only need schema)
	_ = pluginInstance.Init(map[string]interface{}{})

	return pluginInstance.GetSchema(), nil
}

func main() {
	flag.StringVar(&httpAddr, "http", "", "HTTP address to listen on (e.g., :3100)")
	flag.StringV
	flag.Parse()

 "plugins", "./plugins", "Path to plugins directory containing .so files"
	if httpAddr != "" {
		runHTTPServer()
	} else {
		runStdioServer()
	}
}

// runStdioServer runs the MCP server in stdio mode (for Claude Code CLI)
func runStdioServer() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(nil, -32700, "Parse error")
			continue
		}

		handleRequest(&req)
	}
}

// runHTTPServer runs the MCP server in HTTP/SSE mode (for web clients)
func runHTTPServer() {
	mux := http.NewServeMux()

	// SSE endpoint for MCP protocol
	mux.HandleFunc("/sse", handleSSE)

	// POST endpoint for sending messages (alternative to SSE client->server)
	mux.HandleFunc("/message", handleMessage)

	// Simple REST API endpoints (easier for web integration)
	mux.HandleFunc("/api/modules", handleAPIModules)
	mux.HandleFunc("/api/module/", handleAPIModule)
	mux.HandleFunc("/api/search", handleAPISearch)
	mux.HandleFunc("/api/docs", handleAPIDocs)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// CORS middleware
	handler := corsMiddleware(mux)

	fmt.Printf("M3M MCP Server starting on %s\n", httpAddr)
	fmt.Printf("  SSE endpoint: http://localhost%s/sse\n", httpAddr)
	fmt.Printf("  REST API:     http://localhost%s/api/\n", httpAddr)

	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleSSE implements MCP over SSE transport
func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Generate client ID
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	msgChan := make(chan string, 100)

	sseMutex.Lock()
	sseClients[clientID] = msgChan
	sseMutex.Unlock()

	defer func() {
		sseMutex.Lock()
		delete(sseClients, clientID)
		sseMutex.Unlock()
		close(msgChan)
	}()

	// Send endpoint info
	fmt.Fprintf(w, "event: endpoint\ndata: /message?clientId=%s\n\n", clientID)
	flusher.Flush()

	// Keep connection alive and send messages
	for {
		select {
		case msg := <-msgChan:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// Keep-alive ping
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// handleMessage receives MCP requests from web clients
func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("clientId")

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process request and get response
	response := processRequest(&req)

	// If client has SSE connection, send via SSE
	if clientID != "" {
		sseMutex.RLock()
		if ch, ok := sseClients[clientID]; ok {
			data, _ := json.Marshal(response)
			select {
			case ch <- string(data):
			default:
			}
		}
		sseMutex.RUnlock()
	}

	// Also return as HTTP response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// REST API handlers (simpler for web integration)

func handleAPIModules(w http.ResponseWriter, r *http.Request) {
	type ModuleInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Methods     int    `json:"methods"`
		Types       int    `json:"types"`
	}

	var result []ModuleInfo
	for _, m := range moduleSchemas {
		result = append(result, ModuleInfo{
			Name:        m.Name,
			Description: m.Description,
			Methods:     len(m.Methods),
			Types:       len(m.Types),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleAPIModule(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/module/")
	name = strings.TrimPrefix(name, "$")

	for _, m := range moduleSchemas {
		moduleName := strings.TrimPrefix(m.Name, "$")
		if strings.EqualFold(moduleName, name) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m)
			return
		}
	}

	http.Error(w, "Module not found", http.StatusNotFound)
}

func handleAPISearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	type SearchResult struct {
		Module      string `json:"module"`
		Type        string `json:"type"` // "module" or "method"
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var results []SearchResult

	for _, m := range moduleSchemas {
		moduleName := strings.ToLower(m.Name)
		moduleDesc := strings.ToLower(m.Description)

		if strings.Contains(moduleName, query) || strings.Contains(moduleDesc, query) {
			results = append(results, SearchResult{
				Module:      m.Name,
				Type:        "module",
				Name:        m.Name,
				Description: m.Description,
			})
		}

		for _, method := range m.Methods {
			methodName := strings.ToLower(method.Name)
			methodDesc := strings.ToLower(method.Description)
			if strings.Contains(methodName, query) || strings.Contains(methodDesc, query) {
				results = append(results, SearchResult{
					Module:      m.Name,
					Type:        "method",
					Name:        method.Name,
					Description: method.Description,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(moduleSchemas)
	case "typescript", "ts":
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(schema.GenerateAllTypeScript(moduleSchemas)))
	default:
		w.Header().Set("Content-Type", "text/markdown")
		w.Write([]byte(schema.GenerateAllMarkdown(moduleSchemas)))
	}
}

// processRequest handles MCP request and returns response (for HTTP mode)
func processRequest(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: ServerCapabilities{
					Tools: &ToolsCapability{},
				},
				ServerInfo: ServerInfo{
					Name:    "m3m-api",
					Version: "1.0.0",
				},
			},
		}

	case "tools/list":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolsListResult{
				Tools: getToolsList(),
			},
		}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: "Invalid params"},
			}
		}
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  executeToolCall(&params),
		}

	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: "Method not found"},
		}
	}
}

func getToolsList() []Tool {
	return []Tool{
		{
			Name:        "list_modules",
			Description: "List all available M3M JavaScript runtime modules with descriptions",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_module",
			Description: "Get detailed documentation for a specific M3M module including all methods and types",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: "Module name (e.g., '$router', '$database', 'router', 'database')"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "search_api",
			Description: "Search M3M API documentation by keyword. Searches module names, descriptions, and method names.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {Type: "string", Description: "Search keyword (e.g., 'http', 'schedule', 'database')"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "get_full_docs",
			Description: "Get the complete M3M JavaScript API documentation in Markdown format",
			InputSchema: InputSchema{Type: "object"},
		},
	}
}

func executeToolCall(params *ToolCallParams) ToolResult {
	var result string

	switch params.Name {
	case "list_modules":
		result = listModules()
	case "get_module":
		name, _ := params.Arguments["name"].(string)
		result = getModule(name)
	case "search_api":
		query, _ := params.Arguments["query"].(string)
		result = searchAPI(query)
	case "get_full_docs":
		result = schema.GenerateAllMarkdown(moduleSchemas)
	default:
		return ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "Unknown tool: " + params.Name}},
			IsError: true,
		}
	}

	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}
}

func handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		sendResult(req.ID, InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{},
			},
			ServerInfo: ServerInfo{
				Name:    "m3m-api",
				Version: "1.0.0",
			},
		})

	case "notifications/initialized":
		// No response needed for notifications

	case "tools/list":
		sendResult(req.ID, ToolsListResult{
			Tools: []Tool{
				{
					Name:        "list_modules",
					Description: "List all available M3M JavaScript runtime modules with descriptions",
					InputSchema: InputSchema{Type: "object"},
				},
				{
					Name:        "get_module",
					Description: "Get detailed documentation for a specific M3M module including all methods and types",
					InputSchema: InputSchema{
						Type: "object",
						Properties: map[string]Property{
							"name": {Type: "string", Description: "Module name (e.g., '$router', '$database', 'router', 'database')"},
						},
						Required: []string{"name"},
					},
				},
				{
					Name:        "search_api",
					Description: "Search M3M API documentation by keyword. Searches module names, descriptions, and method names.",
					InputSchema: InputSchema{
						Type: "object",
						Properties: map[string]Property{
							"query": {Type: "string", Description: "Search keyword (e.g., 'http', 'schedule', 'database')"},
						},
						Required: []string{"query"},
					},
				},
				{
					Name:        "get_full_docs",
					Description: "Get the complete M3M JavaScript API documentation in Markdown format",
					InputSchema: InputSchema{Type: "object"},
				},
			},
		})

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params")
			return
		}
		handleToolCall(req.ID, &params)

	default:
		sendError(req.ID, -32601, "Method not found")
	}
}

func handleToolCall(id interface{}, params *ToolCallParams) {
	var result string

	switch params.Name {
	case "list_modules":
		result = listModules()

	case "get_module":
		name, _ := params.Arguments["name"].(string)
		result = getModule(name)

	case "search_api":
		query, _ := params.Arguments["query"].(string)
		result = searchAPI(query)

	case "get_full_docs":
		result = schema.GenerateAllMarkdown(moduleSchemas)

	default:
		sendToolError(id, "Unknown tool: "+params.Name)
		return
	}

	sendResult(id, ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	})
}

func listModules() string {
	var sb strings.Builder
	sb.WriteString("# M3M JavaScript Runtime Modules\n\n")
	sb.WriteString("All modules are available globally with `$` prefix.\n\n")

	for _, m := range moduleSchemas {
		sb.WriteString(fmt.Sprintf("## %s\n", m.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))
		sb.WriteString(fmt.Sprintf("**Methods:** %d\n", len(m.Methods)))
		if len(m.Types) > 0 {
			sb.WriteString(fmt.Sprintf("**Types:** %d\n", len(m.Types)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func getModule(name string) string {
	// Normalize name
	name = strings.TrimPrefix(name, "$")
	name = strings.ToLower(name)

	for _, m := range moduleSchemas {
		moduleName := strings.TrimPrefix(m.Name, "$")
		if strings.ToLower(moduleName) == name {
			return m.GenerateMarkdown()
		}
	}

	return fmt.Sprintf("Module '%s' not found. Use list_modules to see available modules.", name)
}

func searchAPI(query string) string {
	query = strings.ToLower(query)
	var sb strings.Builder
	found := false

	sb.WriteString(fmt.Sprintf("# Search Results for '%s'\n\n", query))

	for _, m := range moduleSchemas {
		moduleName := strings.ToLower(m.Name)
		moduleDesc := strings.ToLower(m.Description)

		// Check if module matches
		moduleMatches := strings.Contains(moduleName, query) || strings.Contains(moduleDesc, query)

		// Check methods
		var matchingMethods []schema.MethodSchema
		for _, method := range m.Methods {
			methodName := strings.ToLower(method.Name)
			methodDesc := strings.ToLower(method.Description)
			if strings.Contains(methodName, query) || strings.Contains(methodDesc, query) {
				matchingMethods = append(matchingMethods, method)
			}
		}

		if moduleMatches || len(matchingMethods) > 0 {
			found = true
			sb.WriteString(fmt.Sprintf("## %s\n", m.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))

			if len(matchingMethods) > 0 {
				sb.WriteString("**Matching methods:**\n")
				for _, method := range matchingMethods {
					sb.WriteString(fmt.Sprintf("- `%s` - %s\n", method.Name, method.Description))
				}
			}
			sb.WriteString("\n")
		}
	}

	if !found {
		sb.WriteString("No results found. Try different keywords or use list_modules to see all available modules.\n")
	}

	return sb.String()
}

func sendResult(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func sendError(id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func sendToolError(id interface{}, message string) {
	sendResult(id, ToolResult{
		Content: []ContentBlock{{Type: "text", Text: message}},
		IsError: true,
	})
}
