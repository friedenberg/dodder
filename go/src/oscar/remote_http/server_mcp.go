package remote_http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/mcp"
)

func (server *Server) handleMCP(request Request) (response Response) {
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
	case "resources/list":
		resources := server.getMCPResources()
		mcpResponse.Result = mcp.ResourcesListResult{Resources: resources}

	case "resources/read":
		resources := server.getMCPResources()
		mcpResponse.Result = mcp.ResourcesListResult{Resources: resources}

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
	// TODO: Replace this boilerplate with actual application logic
	// This method should integrate with the Dodder storage system to:
	// - List available Zettels as resources
	// - Query the repository for specific content types
	// - Generate URIs using the existing object ID system
	// - Leverage the existing query system via server.Repo

	return []mcp.Resource{
		{
			URI:         "dodder://example/resource",
			Name:        "Example Resource",
			Description: "A sample resource for testing MCP implementation",
			MimeType:    "text/plain",
		},
	}
}
