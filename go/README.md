# Dodder

A distributed knowledge management system implementing Git-like version control for Zettelkasten-style note-taking.

## Overview

Dodder provides a content-addressable storage system with an append-only object graph that enables multiple versions of knowledge objects to coexist and satisfy concurrent dependency graphs. Built in Go, it offers cryptographic signing and verification of objects across distributed repositories, ensuring data integrity and authenticity in collaborative environments.

## Key Features

- **Versioned Knowledge Storage**: Append-only object graph with content-addressable storage supporting concurrent version histories
- **Cryptographic Security**: Object signing and verification across repositories with ZSTD compression and Age encryption support
- **Advanced Querying**: Sophisticated search capabilities including tags, types, full-text search, and embedded Lua scripting
- **Optimized Performance**: Custom binary index format with highly-optimized IO operations and pool-based memory management
- **Distributed Architecture**: HTTP/Unix socket API server supporting mTLS, remote synchronization, and cross-repository operations
- **Familiar Interface**: Git-style CLI commands (clone, push, pull, diff, checkout) with shell completion support

## Architecture

The system employs a highly-modular architecture with strict dependency layering, enabling clean separation of concerns and maintainable code. The custom binary index format and optimized IO subsystem ensure efficient handling of large knowledge bases while maintaining fast query performance.

## Testing

Comprehensive test coverage through Go unit tests and BATS integration testing framework ensures reliability and correctness of distributed operations.