# Mission Control — TODO

## Phase 1: Foundation
- [ ] Initialize Bun + TypeScript project
- [ ] Install Ink + Zustand dependencies
- [ ] Create folder structure per STANDARDS.md
- [ ] Implement basic layout with 5 zones
- [ ] Top status line (static placeholder)
- [ ] Bottom status line (static placeholder)
- [ ] Search bar component
- [ ] Chat bar component
- [ ] Project list (hardcoded test data)
- [ ] Vim keybindings (hjkl, gg, G, /, q)

## Phase 2: Discovery
- [ ] Project discovery function
- [ ] Detect Vercel projects (.vercel/)
- [ ] Detect Swift projects (.xcodeproj, Package.swift)
- [ ] Detect CLI projects
- [ ] Cache to ~/.hustlemc/projects.json
- [ ] First-run wizard (ask project root)

## Phase 3: Status
- [ ] Git status hook (git status --porcelain)
- [ ] GitHub hook (gh issue list, gh pr list)
- [ ] Vercel status hook (vl --json)
- [ ] Swift build status
- [ ] Background refresh (30s interval)
- [ ] Aggregate counts for top status

## Phase 4: Actions
- [ ] Play/pause toggle
- [ ] Vercel: bun dev + Caddy hostname
- [ ] Swift: build + run
- [ ] CLI: show --help
- [ ] Open production URL
- [ ] Edit docs (nvim r/R/p/t)
- [ ] Launch OpenClaw TUI (c/C)

## Phase 5: Caddy
- [ ] Generate Caddyfile entries
- [ ] Pattern: {name}.localhost → localhost:{port}
- [ ] Reload Caddy on change
- [ ] Port allocation tracking

## Phase 6: OpenClaw Chat
- [ ] Gateway client connection
- [ ] Send message with project context
- [ ] Display response inline
- [ ] Command history

## Phase 7: Detail View
- [ ] Title: 🚀 mc:${name}
- [ ] Full status display
- [ ] Recent commits list
- [ ] Open issues/PRs list
- [ ] File tree preview
- [ ] Action buttons with labels

## Phase 8: Polish
- [ ] p10k-style segment transitions
- [ ] Loading spinners
- [ ] Error states with icons
- [ ] Performance optimization
- [ ] List virtualization
- [ ] Search debounce

## Stretch
- [ ] Multi-select operations
- [ ] Batch git operations
- [ ] Deploy preview links
- [ ] PR review integration
- [ ] Sentry error counts
