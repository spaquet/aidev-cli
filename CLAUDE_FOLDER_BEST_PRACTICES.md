# .claude Folder Best Practices

## What is the `.claude` folder?

The `.claude` folder is created by Claude Code and contains:
- **Memories** (`memory/`) — Project context, user preferences, lessons learned
- **Plans** (`plans/`) — Architecture and implementation plans
- **Hooks** (optional) — Custom shell commands for automation
- **Settings** (optional) — IDE/tool configurations

## Should it be shared in version control?

### ❌ DO NOT commit `.claude` to Git

**Reasons:**
1. **Personal to each developer** — Memories and preferences are specific to one person's workflow
2. **Plans are ephemeral** — They're working documents, not permanent project records
3. **Can contain sensitive info** — API tokens, credentials, or private notes might leak
4. **Clutters the repo** — Not part of the codebase itself

### ✅ Keep it in `.gitignore`

The `.claude` folder should already be in `.gitignore`. If not, add it:

```bash
# .gitignore
.claude/
```

## Best Practices

### For Individual Developers
- Your `.claude/memory/` folder is local to your machine
- It helps Claude understand your project context across conversations
- Keep it updated with important learnings and decisions

### For Team Collaboration
- **Share important context** — Document key decisions in:
  - `GITFLOW.md` (already created)
  - `README.md`
  - Architecture docs
  - Code comments

- **Don't share memories** — Instead, document:
  - Why decisions were made (in commit messages or ADRs)
  - What patterns the project uses (in READMEs)
  - Any gotchas or known issues (in docs)

### For Projects with Multiple Developers
If you're working with a team:
1. Each developer has their own `.claude/` folder (local only)
2. **Store shared context** in Git:
   - `CLAUDE.md` — Project-specific preferences for Claude AI
   - `docs/ARCHITECTURE.md` — System design
   - `docs/DEVELOPMENT.md` — Developer guide
   - Comments in code
   - Commit messages with full context

## Example: CLAUDE.md

Create a `CLAUDE.md` file in your repo root to document project-specific guidance for AI assistants:

```markdown
# Claude Code Guidelines for This Project

## Code Style
- Follow Go idioms: `gofmt`, `go vet`
- Use `make` targets for common tasks
- Keep functions under 40 lines when possible

## Architecture
- TUI runs in separate `internal/tui` package (Bubble Tea)
- Commands in `internal/commands` are stateless
- SSH operations in `internal/ssh` (spawns subprocess)

## Before Making Changes
- Always run `make test` first
- Check `GITFLOW.md` for release process
- Don't modify `.goreleaser.yml` without reviewing existing config

## Known Gotchas
- Windows doesn't have native SSH (test with WSL or skip)
- XDG config paths differ by OS (use `adrg/xdg` package)
- TUI quits before SSH runs (intentional - can't share terminal)
```

## Summary

| Item | Shared? | Location |
|---|---|---|
| `.claude/` folder | ❌ No | `.gitignore` |
| Memories | ❌ No | Local developer machine |
| Plans | ❌ No | Local developer machine |
| Architecture decisions | ✅ Yes | `docs/ARCHITECTURE.md` |
| Development guidelines | ✅ Yes | `CLAUDE.md` or `docs/` |
| Code style | ✅ Yes | Comments, linters, `.golangci.yml` |
| Lessons learned | ✅ Yes | Commit messages, docs/ |

This way, you get the benefits of AI context awareness while keeping shared projects clean and transparent.
