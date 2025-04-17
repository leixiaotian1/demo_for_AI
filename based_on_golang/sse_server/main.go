package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	server *server.MCPServer
}

func NewMCPServer() *MCPServer {
	mcpServer := server.NewMCPServer(
		"example-server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)
	// Add echo tool
	mcpServer.AddTool(mcp.NewTool("echo",
		mcp.WithDescription("Echo back the input"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("Message to echo back"),
		),
	), echoHandler)

	return &MCPServer{
		server: mcpServer,
	}
}

func main() {
	s := NewMCPServer()
	sseServer := s.ServeSSE("localhost:8080")
	log.Printf("SSE server listening on :8080")
	if err := sseServer.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func echoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	msg, ok := req.Params.Arguments["message"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid message parameter")
	}
	return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", msg)), nil
}

func (s *MCPServer) ServeSSE(addr string) *server.SSEServer {
	return server.NewSSEServer(s.server,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
	)
}
