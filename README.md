# Cairn

Cairn is a local-first markdown context layer for small engineering pods. It gives humans and agents a shared, file-readable workspace for capture, promotion, validation, search, sync, and MCP access.

## Start Here

- [Quickstart](docs/user/quickstart.md)
- [Install](docs/user/install.md)
- [Daily Workflows](docs/user/workflows.md)
- [Pilot Guide](docs/user/pilot.md)
- [Local Development Harness](deployments/local-dev/README.md)
- [Architecture Decisions](docs/adr/README.md)

## Install

Pilots should install a released binary into userland:

```sh
curl -fsSL https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
cairn version
```

## Build

```sh
go build -o ./bin/cairn ./cmd/cairn
```

Run from the repository root:

```sh
go run ./cmd/cairn help
```

## Current Shape

Cairn Core v1 is local-first: documents are visible markdown files, search/indexing is local SQLite metadata and full text, and remote sharing is blob-backed sync with conflict refusal. The default quickstart does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer.

For no-service sync development, use the `local_fs` remote-store example in the quickstart. `deployments/local-dev/` contains the no-service local smoke and sample config for that path.

Run the no-service Cairn Core smoke with:

```sh
deployments/local-dev/core-smoke.sh
```

Run the full pilot readiness check with:

```sh
make pilot-check
```
