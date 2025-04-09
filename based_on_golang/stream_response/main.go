package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	APIEndpoint     = "https://api.deepseek.com/v1/chat/completions"
	ToolNameGetTime = "get_current_time"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type DeepSeekClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *DeepSeekClient) ChatStream(messages []Message, tools []Tool, handleResponse func(StreamResponse)) error {
	requestBody := map[string]any{
		"model":    "deepseek-chat",
		"messages": messages,
		"stream":   true,
	}
	if len(tools) > 0 {
		requestBody["tools"] = tools
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", APIEndpoint, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var response StreamResponse
		if err := json.Unmarshal(line[6:], &response); err != nil { // 跳过"data: "前缀
			continue
		}

		handleResponse(response)
	}

	return nil
}

func handleToolCall(toolCall ToolCall) string {
	switch toolCall.Function.Name {
	case ToolNameGetTime:
		return time.Now().Format("2006-01-02 15:04:05")
	default:
		return "Tool not found"
	}
}

var exampleTools = []Tool{
	{
		Type: "function",
		Function: ToolFunction{
			Name:        ToolNameGetTime,
			Description: "Get the current time",
			Parameters: struct {
				Type string `json:"type"`
			}{Type: "object"},
		},
	},
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置 DEEPSEEK_API_KEY 环境变量")
		return
	}
	client := NewClient(apiKey)

	messages := []Message{
		{Role: "user", Content: "现在几点？帮我查下当前时间"},
	}

	handleResponse := func(response StreamResponse) {
		if len(response.Choices) == 0 {
			return
		}

		delta := response.Choices[0].Delta

		if len(delta.ToolCalls) > 0 {
			for _, toolCall := range delta.ToolCalls {
				if toolCall.Type != "function" {
					continue
				}
				result := handleToolCall(toolCall)
				fmt.Printf("\n[工具调用] 结果: %s\n", result)
			}
		}

		if delta.Content != "" {
			fmt.Print(delta.Content)
		}
	}

	err := client.ChatStream(messages, exampleTools, handleResponse)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
	}
}
