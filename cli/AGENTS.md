# AGENTS.md

Instructions for AI coding agents working in the Gel CLI project.

## Scope

- This file applies only to the `cli/` subtree.
- Do not use this file as guidance for the repository-level workspace or `server/`.
- If a task requires changes outside `cli/`, ask before proceeding.
- Treat explicit user instructions as higher priority than this file.

## Project Summary

Gel is a Git-inspired version control CLI written in Go. It is a portfolio and learning project, not a production Git replacement.

Core characteristics:

- Language: Go 1.26.
- CLI framework: Cobra.
- Config format: TOML via `github.com/BurntSushi/toml`.
- Metadata directory: `.gel/` instead of `.git/`.
- Object hashing: SHA-256 instead of SHA-1.
- Default branch: `main`.
- Binary name: `gel`.


## Common Commands

Run commands from `cli/`.

```bash
make run    # go run ./cmd/gel
make fmt    # go fmt ./...
make test   # go test ./...
make vet    # go vet ./...
make build  # go build -o gel ./cmd/gel
```

Before finalizing Go code changes, run:

```bash
make fmt
make test
make vet
make build
```

For documentation-only changes, validation commands are optional. If a required command cannot be run, state why in the final response.

## External Documentation

- Context7 is available through OpenCode MCP for current third-party library documentation.
- Use Context7 when working with external APIs or libraries such as Cobra, TOML parsing, or Go tooling behavior.
- Prefer local source code, tests, and project docs for Gel-specific architecture and behavior.


Package responsibilities:

- `cmd/gel`: executable entrypoint; calls `cli.Execute()`.
- `internal/cli`: Cobra commands, flag parsing, user-facing output, service wiring.
- `internal/domain`: domain models, typed hashes/paths/file modes, repository layout, normalized file stats, canonical object/index codecs, and validation; no internal package imports.
- `internal/storage`: filesystem persistence for objects, index, and config; depends only on `domain`.


## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:

- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use
  `graphify path "<A>" "<B>"` for relationships and
  `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

