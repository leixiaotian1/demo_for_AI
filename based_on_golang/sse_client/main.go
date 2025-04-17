package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()
	client, err := client.NewSSEMCPClient("http://localhost:8080/sse")
	if err != nil {
		log.Fatalf("Failed to create SSE MCP client: %v", err)
	}
	err = client.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start SSE MCP client: %v", err)
	}
	// Initialize
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}

	_, err = client.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to Initialize SSE MCP client: %v", err)
	}
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}

	arguments := map[string]interface{}{
		"message": "Hello SSE!",
	}

	request.Params.Name = "echo"
	request.Params.Arguments = arguments

	// Test echo tool
	result, err := client.CallTool(context.Background(), request)
	if err != nil {
		return
	}
	textContent := result.Content[0].(mcp.TextContent)
	fmt.Println(textContent.Text)
	time.Sleep(100 * time.Second)
}
