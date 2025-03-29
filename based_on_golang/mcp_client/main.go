package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPConfig struct {
	McpServers map[string]MCPServer `json:"mcpServers"`
}

type MCPServer struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	Disabled    bool              `json:"disabled"`
	AutoApprove []string          `json:"autoApprove"`
}

func main() {

	//pgsql其它环境变量已经省略
	configData := `{
		"mcpServers": {
		  "pgsql-mcp-server": {
			"command": "C:\\Users\\Administrator\\go\\bin\\sql-mcp-server.exe",
			"args": [],
			"env": {
			  "GOPATH": "C:\\Users\\Administrator\\go",
			  "GOMODCACHE": "C:\\Users\\Administrator\\go\\pkg\\mod",
			  "DB_HOST": "localhost",
			  //pgsql其它环境变量已经省略
			},
			"disabled": false,
			"autoApprove": []
		  }
		}
	  }`

	// 解析配置文件
	var config MCPConfig
	if err := json.Unmarshal([]byte(configData), &config); err != nil {
		fmt.Printf("Failed to parse config file: %v\n", err)
		return
	}

	// 获取MCP Server配置
	mcpServerConfig, exists := config.McpServers["pgsql-mcp-server"]
	if !exists {
		fmt.Println("MCP Server configuration not found")
		return
	}

	// 检查是否禁用
	if mcpServerConfig.Disabled {
		fmt.Println("MCP Server is disabled")
		return
	}

	// 将环境变量转换为key=value格式的切片
	envSlice := make([]string, 0, len(mcpServerConfig.Env))
	for key, value := range mcpServerConfig.Env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}

	// 使用NewStdioMCPClient启动MCP Server
	c, err := client.NewStdioMCPClient(
		mcpServerConfig.Command,
		envSlice,
		mcpServerConfig.Args...,
	)
	if err != nil {
		fmt.Printf("Failed to start MCP Server: %v\n", err)
		return
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1130*time.Second)
	defer cancel()

	// Initialize the client
	fmt.Println("Initializing client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-client",
		Version: "0.0.0",
	}

	initResult, err := c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf(
		"Initialized with server: %s %s\n\n",
		initResult.ServerInfo.Name,
		initResult.ServerInfo.Version,
	)

	// List Tools
	fmt.Println("Listing available tools...")
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// List tables
	fmt.Println("Listing tables...")
	listTableRequest := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}
	listTableRequest.Params.Name = "list_tables"

	result, err := c.CallTool(ctx, listTableRequest)
	if err != nil {
		log.Fatalf("Failed to list allowed directories: %v", err)
	}
	printToolResult(result)
	fmt.Println()

	// Create table
	fmt.Println("Creating table...")
	createTableRequest := mcp.CallToolRequest{}
	createTableRequest.Params.Name = "create_table"
	createTableRequest.Params.Arguments = map[string]interface{}{
		"schema": "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100), email VARCHAR(100) UNIQUE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
	}

	result, err = c.CallTool(ctx, createTableRequest)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	printToolResult(result)
	fmt.Println()

	// Write table
	fmt.Println("Writing table...")
	wirteTableRequest := mcp.CallToolRequest{}
	wirteTableRequest.Params.Name = "write_query"
	wirteTableRequest.Params.Arguments = map[string]interface{}{
		"query": "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')",
	}

	result, err = c.CallTool(ctx, wirteTableRequest)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	printToolResult(result)
	fmt.Println()

	// Read table
	fmt.Println("Reading table...")
	readTableRequest := mcp.CallToolRequest{}
	readTableRequest.Params.Name = "read_query"
	readTableRequest.Params.Arguments = map[string]interface{}{
		"query": "SELECT * FROM users LIMIT 10",
	}

	result, err = c.CallTool(ctx, readTableRequest)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	printToolResult(result)
	fmt.Println()

}

func printToolResult(result *mcp.CallToolResult) {
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		} else {
			jsonBytes, _ := json.MarshalIndent(content, "", "  ")
			fmt.Println(string(jsonBytes))
		}
	}
}
