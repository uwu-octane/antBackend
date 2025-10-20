# Repository Guidelines

## Project Structure & Module Organization
- `go.work` links `api`, `auth`, `gateway`, `user`, `common`, `cmd`; run tooling from the repo root so replacements resolve.
- `gateway/` holds the HTTP edge (`gateway.go`, `internal/`, `util/`) plus the spec `gateway/gateway.api` and configs in `gateway/etc/`.
- RPC services live in `auth/` and `user/`; keep handlers in `internal/logic`, transport structs in `internal/types`, and runtime YAML under each `etc/`.
- Shared code and ops artifacts sit in `common/` (`db/migrations` for goose SQL, `envloader` for `.env`), while `cmd/boot` wires all services.

## Build, Test, and Development Commands
- `go work sync` — refresh workspace entries after adding modules or pulling dependencies.
- `go test ./...` — execute unit suites across active modules; run from the root to respect workspace replaces.
- `go run ./cmd/boot -gateway gateway/etc/gateway-api.yaml -auth auth/etc/auth.yaml -user user/etc/user.yaml` — start gateway + RPC trio with defaults.
- `./test_ratelimit.sh` — manual gateway login throttling check; requires services on `http://127.0.0.1:28256`.

## Coding Style & Naming Conventions
- Keep Go code `gofmt`/`goimports` clean; tabs for indentation and canonical GoZero naming (`Auth`, `User`, `Gateway` prefixes).
- Place business logic under `internal/logic`, shared DTOs in `internal/types`, and keep package-private helpers close to use.
- Generate APIs and RPCs with `goctl` (configured via `.goctl.yaml`); avoid hand-editing generated files without template updates.
- Config files follow `snake_case.yaml`; prefer injecting env via `common/envloader` rather than `os.Getenv` directly.

## Testing Guidelines
- Add table-driven `*_test.go` beside new logic (see `auth/internal/logic` for structure and mocks).
- Exercise error paths and cache/DB behaviour with in-memory fakes; document external dependencies inside test comments.
- Follow the integration flow in `TEST_AUTH.md` when changing auth, and rerun the rate-limit script after gateway throttling edits.

## Commit & Pull Request Guidelines
- Mirror history: use lowercase conventional commit prefixes (`feat:`, `fix:`, `refactor:`) plus a short imperative subject.
- Scope commits to one service or shared concern; call out touched modules in the subject or body when cross-cutting.
- Pull requests should include a summary, linked issue, config/migration callouts, and evidence of manual or scripted tests.

## Configuration & Ops Notes
- `.env` files load automatically through `common/envloader`; supply service overrides locally without committing secrets.
- Sync `auth/etc`, `user/etc`, and `gateway/etc` when updating registries, ports, or JWT settings to keep services aligned.
- Place new migrations under `common/db/migrations/<service>` and note required `goose postgres <DSN> up` steps in your PR.
