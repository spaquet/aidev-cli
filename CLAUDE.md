# AIDev CLI Development Guide

**AIDev CLI** is a Go/Bubble Tea TUI for managing AI development cloud VMs. Device flow auth, SSH integration, instance management.

## Key Info

- **Language**: Go 1.21+
- **UI**: Bubble Tea + Lipgloss (TUI framework)
- **Auth**: OAuth2 device flow (no passwords stored)
- **Main entry**: `cmd/aidev/main.go`
- **Core packages**: `internal/{api,auth,commands,models,ssh,tui}`
- **Frontend**: `internal/tui` (Bubble Tea models + views)
- **Backend**: Rails 8.1 at `https://api.sandbox.example.com`

## Current Focus

**Device flow authentication** replaced email/password login. Welcome screen shows before device flow initiates.

## Code Patterns

- **No emojis in TUI text**: Use plain ASCII `[OK]`, `[ERROR]`, `[!]`, `[i]`
- **Styles**: `internal/tui/views/styles.go` defines shared colors/styles
- **Async in Bubble Tea**: Commands return `func() tea.Msg` closures for background work
- **State machines**: Use iota constants (e.g., `LoginStateWelcome`)
- **Errors**: Validate at API boundaries only; trust internal code

## Common Tasks

**Add a new command**: `internal/commands/newcmd.go` → add to `main.go` → import/register

**Add TUI screen**: Create `internal/tui/views/newscreen.go` with Model/Init/Update/View

**Update docs**: `docs/{auth-spec,rails-api-spec,tui-design,ARCHITECTURE}.md` (keep in sync with code)

## Testing

```bash
go test ./...           # Run all tests
go build ./...          # Verify compilation
aidev tui              # Test TUI locally
```

No network calls until device flow explicitly initiated (welcome screen).

## No-nos

- ❌ Commit built binaries (bin/ in .gitignore)
- ❌ Add features beyond what's asked
- ❌ Skip error handling at API boundaries
- ❌ Use hardcoded credentials/tokens
- ❌ Add premature abstractions

---

**Last updated**: 2026-03-27
