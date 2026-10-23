package mcpbridge

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type connectedProfile struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

var profileSchema = json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","minLength":1,"pattern":"\\S"},"name":{"type":"string"},"email":{"type":"string"}},"required":["id"],"additionalProperties":false}`)

// The designated OpenAI profile tool has a standard top-level response instead
// of the operation envelope. Identity still comes from the actual scoped API.
func (s *Server) profileTool(d definition) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := map[string]json.RawMessage{}
		if req == nil || req.Params == nil {
			return profileError("Profile request unavailable"), nil
		}
		if raw := req.Params.Arguments; len(raw) > 0 {
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &args) != nil || args == nil || len(args) != 0 {
				return profileError("This tool accepts an empty object and reads only the connected account."), nil
			}
		}
		result, output, err := s.call(ctx, req, d, args)
		if err != nil {
			return nil, err
		}
		if output.Status != "success" {
			if result == nil {
				result = &mcp.CallToolResult{}
			}
			result.IsError = true
			message := output.Message
			if message == "" {
				message = "Profile unavailable. Check your Hackways connection and retry."
			}
			result.Content = []mcp.Content{&mcp.TextContent{Text: message}}
			return result, nil
		}
		var envelope struct {
			User struct {
				ID    string `json:"id"`
				Name  string `json:"display_name"`
				Email string `json:"email"`
			} `json:"user"`
		}
		if json.Unmarshal(output.Data, &envelope) != nil || strings.TrimSpace(envelope.User.ID) == "" {
			return profileError("Profile identity unavailable. Reconnect your Hackways account."), nil
		}
		profile := connectedProfile{ID: envelope.User.ID, Name: envelope.User.Name, Email: envelope.User.Email}
		raw, err := json.Marshal(profile)
		if err != nil {
			return profileError("Profile response unavailable"), nil
		}
		return &mcp.CallToolResult{StructuredContent: profile, Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil
	}
}
func profileError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: message}}}
}
