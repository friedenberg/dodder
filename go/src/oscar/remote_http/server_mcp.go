package remote_http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/mcp"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func (server *Server) handleMCP(request Request) (response Response) {
	response.Headers().Set("Content-Type", "application/json")

	var mcpRequest mcp.Request

	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&mcpRequest); err != nil {
		response.MCPError(
			http.StatusBadRequest,
			nil, -32700, "Parse error", nil,
		)

		return
	}

	if mcpRequest.JSONRPC != "2.0" {
		response.MCPError(
			http.StatusBadRequest,
			mcpRequest.ID,
			-32600,
			"Invalid Request",
			nil,
		)

		return
	}

	mcpResponse := mcp.Response{
		JSONRPC: "2.0",
		ID:      mcpRequest.ID,
	}

	switch mcpRequest.Method {
	case "initialize":
		mcpResponse.Result = mcp.InitializeResponse{
			ProtocolVersion: "2024-11-05",
			Capabilities: mcp.ServerCapabilities{
				Resources: &mcp.ResourcesCapability{
					Subscribe:   false,
					ListChanged: true,
				},
			},
			ServerInfo: mcp.ServerInfo{
				Name:    "dodder",
				Version: "1.0.0",
			},
		}

	case "resources/list":
		resources := server.getMCPResources()
		mcpResponse.Result = mcp.ResourcesListResult{Resources: resources}

	case "resources/read":
		var params mcp.ResourcesReadParams
		if mcpRequest.Params != nil {
			paramsBytes, err := json.Marshal(mcpRequest.Params)
			if err != nil {
				mcpResponse.Error = &mcp.Error{
					Code:    -32602,
					Message: "Invalid params",
				}
				break
			}
			if err := json.Unmarshal(paramsBytes, &params); err != nil {
				mcpResponse.Error = &mcp.Error{
					Code:    -32602,
					Message: "Invalid params",
				}
				break
			}
		}

		contents, err := server.readMCPResource(params.URI)
		if err != nil {
			mcpResponse.Error = &mcp.Error{
				Code:    -32602,
				Message: fmt.Sprintf("Failed to read resource: %v", err),
			}
		} else {
			mcpResponse.Result = mcp.ResourcesReadResult{Contents: contents}
		}

	default:
		mcpResponse.Error = &mcp.Error{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	responseBytes, err := json.Marshal(mcpResponse)
	if err != nil {
		response.MCPError(
			http.StatusInternalServerError,
			mcpRequest.ID,
			-32603,
			"Internal error",
			nil,
		)

		return
	}

	response.StatusCode = http.StatusOK
	response.Body = io.NopCloser(bytes.NewReader(responseBytes))

	return
}

func (server *Server) getMCPResources() []mcp.Resource {
	// For now, return a simple list of example resources
	// TODO: Implement proper querying when we understand the query system
	// better
	resources := []mcp.Resource{
		{
			URI:         "dodder://example/test-zettel",
			Name:        "Test Zettel",
			Description: "A sample Zettel for testing MCP implementation",
			MimeType:    "text/markdown",
		},
	}

	// Log that we're returning example resources
	ui.Log().Print("getMCPResources: returning example resources")

	return resources
}

func (server *Server) readMCPResource(
	uri string,
) ([]mcp.ResourceContent, error) {
	// For now, return example content for testing
	// TODO: Implement proper object reading when we understand the store system
	// better

	if uri == "dodder://example/test-zettel" {
		content := mcp.ResourceContent{
			URI:      uri,
			MimeType: "text/markdown",
			Text:     "# Test Zettel\n\nThis is example content for the test Zettel.\n\n- Item 1\n- Item 2\n- Item 3\n",
		}
		return []mcp.ResourceContent{content}, nil
	}

	return nil, errors.Errorf("resource not found: %s", uri)
}
