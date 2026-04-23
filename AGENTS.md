# AGENTS.md

## About

This is the source for Zen - a system-wide proxy-based ad-blocker and privacy guard. Built using Wails as the application framework, Go for core logic, and TS/React for the UI.

## Commands

Use `task` commands when available.

- Build: `task build-dev`
- Tests (Go only): `task test`
- Lint (Go and frontend): `task lint`
- Lint (frontend only): `task frontend:lint`
- Format (Go): `task fmt-go`
- Format check (frontend): `task frontend:fmt`

## File structure

- `main.go` - main application entrypoint
- `internal/` - core Go application logic
- `frontend/` - UI

## Working conventions

- Prefer `task` commands over manual shell commands
- Run `task lint` after changes
- Run `task test` after Go changes

## Issue and PR guidelines

- Never create an issue
- Never create a PR
