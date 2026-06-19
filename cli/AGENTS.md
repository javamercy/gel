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

- Language: Go 1.25.
- CLI framework: Cobra.
- Config format: TOML via `github.com/BurntSushi/toml`.
- Metadata directory: `.gel/` instead of `.git/`.
- Object hashing: SHA-256 instead of SHA-1.
- Default branch: `main`.
- Binary name: `gel`.

Use `docs/ROADMAP.md` as the authoritative source for implemented and planned commands.

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

## Architecture

Gel follows a layered architecture. Keep dependencies flowing in this direction:

```text
CLI commands
Feature services
Core services
Storage
Domain
```

Package responsibilities:

- `cmd/gel`: executable entrypoint; calls `cli.Execute()`.
- `internal/cli`: Cobra commands, flag parsing, user-facing output, service wiring.
- `internal/domain`: pure domain models, value types, serialization, validation; no internal package imports.
- `internal/storage`: filesystem persistence for objects, index, and config; depends only on `domain`.
-
`internal/core`: shared services for objects, refs, config, index, path resolution, tree walking, and change detection.
- `internal/staging`: staging-area commands and index mutations.
- `internal/tree`: tree object commands.
- `internal/commit`: commit and log commands.
- `internal/branch`: branch and switch commands.
- `internal/diff`: diff engine and formatting.
- `internal/inspect`: read-only inspection and restore/status/show services.
- `internal/setup`: repository initialization.
- `internal`: package-level services that do not yet have a narrower feature package, currently including reset.
- `internal/validate`: reusable validation helpers.

Dependency rules:

- `domain` must not import any other `Gel/internal/...` package.
- `storage` may depend on `domain` only.
- `core` may depend on `domain` and `storage`.
- Feature packages may depend on `domain`, `core`, and other feature packages only when the dependency is necessary.
- `cli` may depend on all internal packages and is responsible for dependency injection.
- No package may import `internal/cli`.

## Service Wiring

- Instantiate services in `internal/cli/root.go` inside `initializeServices()`.
- Commands that require an existing repository use the root command `PersistentPreRunE` to initialize services lazily.
- Commands that do not require a repository must be listed in `commandsWithoutRepository`.
- Prefer constructor injection with concrete service types.
- Do not introduce interfaces unless multiple implementations or a consumer-side test seam makes the abstraction necessary.

## Coding Standards

- Follow standard Go formatting and naming conventions.
- Use `MixedCaps` and `mixedCaps`; do not use underscores in Go identifiers.
- Keep acronyms capitalized: `SHA256`, `UserID`, `HTTP`.
- Name constructors `New<Type>`.
- Name service structs `<Noun>Service`.
- Keep business logic out of `internal/cli`; CLI code should parse input, call services, and format output.
- Prefer small, focused functions. Extract helpers when they improve clarity or reuse.
- Do not add dead code, unused exports, or speculative abstractions.
- Do not add external dependencies without explicit user approval.

Documentation rules:

- Every exported type, function, method, and constant must have a Godoc comment starting with the identifier name.
- Comments should explain non-obvious rationale, not restate the code.
- Preserve unrelated existing comments.

## Error Handling

- Return errors instead of panicking.
- Use package-level sentinel errors with `errors.New()` where callers need structured checks.
- Wrap errors with context using `fmt.Errorf("context: %w", err)`.
- Use `errors.Is()` and `errors.As()` for error checks.
- Do not compare error strings.
- Do not use `panic()` or `log.Fatal()`.
- Non-CLI packages must not print user-facing output; return data and errors instead.
-
`internal/cli` may print command output using Cobra output streams or standard formatting helpers already used in the package.

## Domain Model Rules

- Domain types should be deterministic and free of filesystem side effects unless explicitly modeling path or stat data.
- Domain types that store mutable reference data, especially
  `[]byte` or slices, must use defensive copies in constructors and accessors.
- Validate domain invariants at construction boundaries where practical.
- Keep serialization and deserialization behavior stable unless the task explicitly changes the on-disk format.

## Object And Repository Format

Gel repository layout:

```text
.gel/
  HEAD
  config.toml
  index
  objects/<2-char-prefix>/<remaining-hash>
  refs/heads/main
```

Format invariants:

- Objects serialize as `<type> <size>\x00<body>`.
- Objects are zlib-compressed in `.gel/objects`.
- Object IDs are SHA-256 hashes encoded as 64 lowercase hex characters.
- Supported object types are blob, tree, and commit.
- The index uses `DIRC`, version 2, SHA-256 checksums, and 8-byte entry alignment.
- Refs are plain text files containing a hex hash and newline.
- Symbolic refs use `ref: <target>\n`.

If changing repository format or `.gel` layout:

- Update constants in `internal/domain/constants.go` and any dependent code.
- Update serializers, deserializers, readers, and writers together.
- Add or update tests when test infrastructure exists for the changed area.
- Ask before making backward-incompatible format changes.

## Adding Or Changing Commands

Use this workflow for CLI commands:

1. Put business logic in the appropriate feature package under `internal/`.
2. Add or modify domain/core support only when the feature requires it.
3. Add the Cobra command in `internal/cli/`.
4. Wire the service in `internal/cli/root.go`.
5. Add commands that do not require a repository to `commandsWithoutRepository`.
6. Update `docs/ROADMAP.md` when command status changes.
7. Run the standard validation commands for Go changes.

Do not duplicate substantial business logic between CLI commands and services.

## Tests

- Use `go test ./...` through `make test` for repository-wide validation.
- Prefer focused tests near the package being changed when adding test coverage.
- For behavior changes, add tests when the surrounding package already has suitable test patterns or when the change introduces a clear regression risk.
- Do not add broad, brittle tests that only restate implementation details.

## Documentation

- Keep `AGENTS.md` concise and stable.
- Do not paste large source trees, generated command templates, or volatile roadmap content into this file.
- Use `docs/ROADMAP.md` for command status.
- Read files under
  `docs/guides/` only when the task is documentation-related or the guide is directly relevant to the implementation.

## Git Workflow

- Do not commit, amend, tag, push, or open pull requests unless explicitly requested.
- If asked to commit, inspect `git status`, `git diff`, and recent commit history first.
- Commit messages must use Conventional Commits, preferably with a package or feature scope.
- Keep commits focused on the requested change.
- Never revert or overwrite user changes unless explicitly asked.
- Never use destructive git commands such as `git reset --hard` or `git checkout --` without explicit approval.

Commit examples:

```text
feat(branch): add branch deletion safety check
fix(ref): remove refs with os.Remove
refactor(index): normalize ID acronym naming
docs(cli): tighten agent instructions
```

## Safety Boundaries

Ask before:

- Touching files outside `cli/`.
- Adding or replacing external dependencies.
- Making large cross-package refactors.
- Changing the `.gel` on-disk format.
- Introducing backward-incompatible command behavior.
- Removing existing public commands, flags, or serialized fields.

Never:

- Commit secrets, credentials, tokens, or private local paths.
- Edit generated or build-output files as source changes.
- Treat `cli/.gel/` as authoritative project source.
- Hide validation failures or omit relevant command output from the final response.

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

