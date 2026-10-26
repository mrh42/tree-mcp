package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (d *Data) GetPersonInfo(ctx context.Context, id int) (r string, e error) {
	r = d.Info(id)
	return
}
func (d *Data) GetGedcom(ctx context.Context, id int) (r string, e error) {
	r = d.GEDCOMInfo(id)
	return
}
func (d *Data) SearchByName(ctx context.Context, name string) (r string, e error) {
	ids := d.Search("NAME", name)
	r = fmt.Sprintf("Matches for '%s': %d\n", name, ids)
	return
}
func (d *Data) SearchByBirthPlace(ctx context.Context, place string) (r string, e error) {
	ids := d.Search("BIRTH", place)
	r = fmt.Sprintf("Born in '%s': %d\n", place, ids)
	return
}
func (d *Data) SearchByDeathPlace(ctx context.Context, place string) (r string, e error) {
	ids := d.Search("DEATH", place)
	r = fmt.Sprintf("Death in '%s': %d\n", place, ids)
	return
}

func main() {
	db := NewData("mrh-tree.ged")

	s := server.NewMCPServer(
		"FamilyTreeServer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool: get_person_info
	s.AddTool(
		mcp.NewTool("get_person_info",
			mcp.WithDescription("Retrieves detailed profile information for an individual by their database ID."),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("Unique integer identifier")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idVal, err := request.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid argument 'id': %v", err)), nil
			}
			data, err := db.GetPersonInfo(ctx, idVal)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Lookup failed: %v", err)), nil
			}
			return mcp.NewToolResultText(data), nil
		},
	)

	// Tool: get_gedcom_record
	s.AddTool(
		mcp.NewTool("get_gedcom_record",
			mcp.WithDescription("Returns raw GEDCOM records for an individual."),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("Unique integer identifier")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idVal, err := request.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid argument 'id': %v", err)), nil
			}
			data, err := db.GetGedcom(ctx, idVal)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("GEDCOM query failed: %v", err)), nil
			}
			return mcp.NewToolResultText(data), nil
		},
	)

	// Tool: search_by_name
	s.AddTool(
		mcp.NewTool("search_by_name",
			mcp.WithDescription("Searches the family tree database by name."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name to search")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := request.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid argument 'name': %v", err)), nil
			}
			data, err := db.SearchByName(ctx, name)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Name search failed: %v", err)), nil
			}
			return mcp.NewToolResultText(data), nil
		},
	)

	// Tool: search_by_birth_place
	s.AddTool(
		mcp.NewTool("search_by_birth_place",
			mcp.WithDescription("Searches for individuals born in a specific location."),
			mcp.WithString("location", mcp.Required(), mcp.Description("Birth location")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			loc, err := request.RequireString("location")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid argument 'location': %v", err)), nil
			}
			data, err := db.SearchByBirthPlace(ctx, loc)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Birth search failed: %v", err)), nil
			}
			return mcp.NewToolResultText(data), nil
		},
	)

	// Tool: search_by_death_place
	s.AddTool(
		mcp.NewTool("search_by_death_place",
			mcp.WithDescription("Searches for individuals deceased in a specific location."),
			mcp.WithString("location", mcp.Required(), mcp.Description("Death location")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			loc, err := request.RequireString("location")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid argument 'location': %v", err)), nil
			}
			data, err := db.SearchByDeathPlace(ctx, loc)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Death search failed: %v", err)), nil
			}
			return mcp.NewToolResultText(data), nil
		},
	)

	// Create the Streamable HTTP Handler
	// Disabling localhost protection allows connections across network interfaces (e.g. Tailscale / 100.64.x.x)
	streamableServer := server.NewStreamableHTTPServer(
		s,
		server.WithDisableLocalhostProtection(true),
	)

	mux := http.NewServeMux()
	// Open WebUI can now target /mcp or /events with both GET and POST
	mux.Handle("/mcp", streamableServer)
	mux.Handle("/events", streamableServer)

	addr := ":8020"
	log.Printf("MCP Streamable HTTP Server listening on %s (routes: /mcp, /events)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
