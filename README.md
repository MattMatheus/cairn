# Cairn

Cairn is a local-first markdown context layer for small engineering pods. It gives humans and agents a shared, file-readable workspace for capture, promotion, validation, search, sync, and MCP access.

## Start Here

- [Quickstart](docs/user/quickstart.md)
- [Daily Workflows](docs/user/workflows.md)
- [Infra Prep Checklist](deployments/azure-container-apps-indexer/INFRA-PREP.md)
- [Architecture Decisions](docs/adr/README.md)

## Build

```sh
go build -o ./bin/cairn ./cmd/cairn
```

Run from the repository root:

```sh
go run ./cmd/cairn help
```

## Current Shape

Cairn is usable locally without remote infrastructure. Azure Blob sync and the remote indexer are wired behind configuration boundaries, but live remote use still requires Azure resources, auth, and deployment setup.
