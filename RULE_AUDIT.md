# tddcheck rule audit

This audit classifies the current checks as mechanical architecture rules for
layered Go projects. They complement, but do not replace, standard Go linters
such as `go vet`, `staticcheck`, `govulncheck`, and `golangci-lint`.

## Mainstream alignment

- Package names matching directory names align with common Go package guidance.
- Import direction checks align with Clean Architecture and hexagonal
  architecture dependency rules when the project uses explicit layer folders.
- Error variables named `Err*`, error helpers in `errors.go`, and `Error`/`Is`/
  `As`/`Unwrap` methods align with idiomatic Go error APIs.
- Keeping `context.WithValue` behind package helpers aligns with Go's guidance
  to avoid ad hoc context values.
- Keeping DTO, mapper, handler, repository, validation, and schema code in
  narrow files aligns with a strict separation-of-responsibilities policy.

## Project-specific policy

These rules are stronger than general Go community standards and should remain
configurable or documented as local architecture policy:

- Mandatory filenames such as `entity.go`, `dto.go`, `mapper.go`,
  `validation.go`, `service.go`, `repository.go`, and `schema.go`.
- CQRS suffixes such as `Command`, `Query`, `Result`, `UseCase`,
  `CommandHandler`, and `QueryHandler`.
- Huma-specific handler and validation signatures.
- Bun/GORM/database-specific forbidden imports and calls.
- DTO and schema suffix requirements.
- `utils.go` allowing only private `util*` helpers.

## Gaps checked in this pass

- `layer` previously assumed module-internal imports under
  `<module>/internal/...` and treated only the first relative path segment as the
  source layer. It now also recognizes imports under the configured scan root and
  detects layer folders below wrappers such as `internal/domain`.
- `context` previously detected only `context.WithValue` and
  `context.Context` when the import name was exactly `context`. It now also
  handles aliases such as `stdctx "context"`.

## Remaining gaps

- Most rules parse syntax only and do not type-check. This keeps the checks
  fast, but aliases through variables, wrapper methods, generated code, and
  type aliases can still evade detection.
- Import checks use exact import paths. Consider prefix or module-group matching
  if framework packages are commonly imported from subpackages.
- Database test checks still use string needles. They are intentionally narrow
  and should be migrated to AST matching if more database helpers are added.
- Handler checks are heavily Huma/httpapi-specific. Projects using chi, Gin,
  Echo, Connect, or gRPC should override or disable those defaults.
- The tool does not cover mainstream security and correctness checks such as SQL
  injection, unchecked errors, concurrency races, shadowing, or vulnerable
  dependencies. Run dedicated tools for those concerns.

## Recommended baseline

Use `tddcheck` for architecture boundaries, and pair it with:

- `go test ./...`
- `go vet ./...`
- `staticcheck ./...`
- `govulncheck ./...`
- a configured `golangci-lint run`
