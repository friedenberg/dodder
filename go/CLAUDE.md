# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zit is a Zettelkasten-style knowledge management system written in Go. It provides Git-like version control for managing interconnected notes (Zettels) with content-addressable storage, sophisticated querying, and remote synchronization capabilities.

## Build and Development Commands

### Core Development Tasks
- **Build**: `just build` (builds binary to `build/zit`)
- **Test**: `just test` (runs Go unit tests + BATS integration tests)
- **Unit Tests Only**: `just test-go-unit` or `go test -v ./...`
- **Single Package Test**: `go test -v ./src/path/to/package`
- **Clean**: `just clean` (clears Go caches)
- **Check**: `just check` (vulnerability scan + vet)
- **Generate**: `just build-go-generate` (runs `go generate ./...`)

### Code Quality
- **Format**: `just codemod-go-imports` (runs goimports on all Go files)
- **Vulnerability Check**: `just check-go-vuln`
- **Go Vet**: `just check-go-vet`

### Alternative Build Systems
- **Nix**: `just build-nix` (requires Nix with flakes)
- **Docker**: `just build-docker`

## Architecture Overview

### NATO Phonetic Module Organization
The codebase uses NATO phonetic alphabet naming for layered architecture with strict dependency ordering (each layer can only depend on previous layers):

- **alfa**: Foundation (errors, interfaces, primitives)
- **bravo**: Basic utilities (UI, pools, values, flags)
- **charlie**: Collections and data structures (sets, I/O, files)
- **delta**: Data processing (SHA, strings, Lua VM, encryption)
- **echo**: Object identification and metadata systems
- **foxtrot**: Configuration and workspace management
- **golf**: Command framework and object metadata
- **hotel**: Repository and environment management
- **india**: Indexing and search systems
- **juliett**: SKU (core object management and storage)
- **kilo**: Query system and object formatting
- **lima**: Storage engines and text organization
- **mike**: Main store implementation
- **november**: Local working copy management
- **oscar**: Remote operations and HTTP API
- **papa**: User operations and command components
- **quebec**: CLI command definitions

### Core Domain Model

#### Key Concepts
- **Zettels**: Fundamental content units with unique IDs and metadata
- **SKUs (Stock Keeping Units)**: Versioned objects representing all content types
- **Object IDs**: Three-part identifiers with genre/type information
- **Content-Addressable Storage**: SHA-based blob storage with inventory tracking
- **Tags and Types**: Hierarchical organization system
- **Working Copy**: Git-like checked-out file system

#### Critical Types and Interfaces
- `sku.Transacted`: Core versioned object type (src/juliett/sku/)
- `interfaces.ObjectId`: Universal identifier interface (src/alfa/interfaces/)
- `store.Store`: Main storage engine (src/mike/store/)
- `command.Command`: CLI command interface (src/golf/command/)

### Storage Architecture
- **Blob Store**: Content-addressable binary storage
- **Inventory Lists**: Object metadata and relationship tracking  
- **Stream Index**: Binary indexing for fast object access
- **Zettel ID Index**: Specialized indexing for note relationships
- **Dormant Index**: Inactive/archived object tracking

### Command System
Commands are registered in `src/quebec/commands/` and follow a consistent pattern:
- Flag parsing via `flag.FlagSet`
- Request objects with context and configuration
- Standardized error handling through `alfa/errors`

## Key Development Patterns

### Error Handling
Uses custom error system in `src/alfa/errors/` with:
- Context-aware error wrapping
- Stack trace support
- Signal handling for graceful shutdown
- Helpful error formatting

### Object Lifecycle
1. Objects created as `sku.Proto` (prototype)
2. Transacted through `sku.Transacted` 
3. Stored via content-addressable SHA
4. Indexed for query and retrieval
5. Can be checked out to filesystem

### Testing Strategy
- Unit tests: `*_test.go` files throughout codebase
- Integration tests: BATS framework in `zz-tests_bats/`
- Test data: Generated fixtures via `test-generate_fixtures`

## Module Import Patterns

When working with this codebase:
- Import paths follow `code.linenisgreat.com/dodder/go/zit/src/{module}/{package}`
- Respect the NATO alphabet dependency hierarchy
- Use existing interfaces rather than concrete types where possible
- Follow established patterns in similar modules

## Important Files for Understanding System
- `main.go`: Application entry point and error handling
- `src/quebec/commands/main.go`: Command dispatch system
- `src/juliett/sku/main.go`: Core object model
- `src/mike/store/main.go`: Primary storage implementation
- `src/november/local_working_copy/main.go`: Working copy management