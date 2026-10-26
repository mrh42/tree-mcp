# Family Tree MCP Server

A small, read-only example of a **Model Context Protocol (MCP)** server written in Go.

The server exposes a local GEDCOM family-tree data file as a set of tools that an MCP-capable AI client can discover and call. It is intended as a simple demonstration of how an AI assistant can interact with approved, locally hosted data or services in an air-gapped or otherwise restricted network.

> This is a demonstration project. Review the security considerations before connecting it to operational, sensitive, or production data.

---

## What Is MCP?

The **Model Context Protocol (MCP)** is an open client-server protocol that gives AI applications a standardized way to discover and use external capabilities.

In practical terms, MCP separates responsibilities:

- **MCP client** An AI-enabled application that connects to one or more MCP servers. The client presents available tools to the model and sends approved tool calls to the server.
- **MCP server** A service that advertises capabilities and performs narrowly defined actions, such as querying a database, retrieving a document, calling an internal API, or running an approved workflow.
- **Tools** Callable functions exposed by an MCP server. Each tool has a name, description, input schema, and result.
- **Resources**  Data made available to an MCP client, typically addressed by URI.
- **Prompts** Reusable prompt templates that a server can offer to clients.

This project implements the **tools** capability. It does not currently expose MCP resources or prompts.

MCP does not make an AI model inherently able to access a network, application, or dataset. The MCP server developer and system owner decide exactly what data and actions are available. The server should enforce appropriate authentication, authorization, validation, auditing, and data-handling controls.

---

## What This Example Does

This server loads a GEDCOM file named `mrh-tree.ged` and exposes read-only family-tree lookup functions through MCP.

When an MCP-capable client connects, it can discover the server's available tools and invoke them with structured arguments. The server performs the requested lookup against the locally loaded GEDCOM data and returns text containing JSON or search results.

### High-Level Flow

```text
+--------------------+
| MCP-capable client |
| (AI application)   |
+---------+----------+
          |
          | MCP / Streamable HTTP
          | HTTP requests to /mcp or /events
          v
+-------------------------------+
| FamilyTreeServer              |
| Go + mcp-go                   |
| Port 8020                     |
+---------------+---------------+
                |
                | Local file access
                v
+-------------------------------+
| mrh-tree.ged                  |
| GEDCOM family-tree data       |
+-------------------------------+
