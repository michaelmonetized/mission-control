# Mission Control — Implementation Plan

## Overview

Mission Control is a p10k-inspired TUI project manager for HurleyUS. It provides a unified dashboard for managing Vercel deployments, Swift builds, CLI tools, and git/GitHub status across all projects.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🚀Mission Control  22◬ 2󱫟 8⨻ 3   3󰸞 2✘   42 18 9 14   │ ← Top Status
├─────────────────────────────────────────────────────────────────────────┤
│ /                                                                       │ ← Search Bar
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│ [||] 󰐎 project-name   12 1 2 2 ........................  󰑢 󱔘  󱐏   │
│ [||] 󰐎 another-proj    5 0 1 3 ........................  󰑢 󱔘  󱐏   │
│  ▶  󰣪 swift-app        0 0 0 0 ........................  󰑢 󱔘  󱐏   │
│                                                                         │ ← Project List
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ >                                                                       │ ← OpenClaw Chat
├─────────────────────────────────────────────────────────────────────────┤
│  86 projects   1.2k   412   89   23󱫟   14⨻                      │ ← Bottom Status
└─────────────────────────────────────────────────────────────────────────┘
```

## Layout Zones

### 1. Top Status Line (p10k style)
**Black text on colored backgrounds.**

Left → Right:
- `🚀Mission Control` (green bg) or `🚀 mc:${name}` (project detail view)
- Vercel: `${ready}◬ ${building}󱫟 ${queued}⨻ ${failed}` (yellow bg)
- Swift: `${success}󰸞 ${failed}✘` (magenta bg)
- Git: `${total} ${untracked} ${modified} ${issues} ${prs}` (cyan bg)

### 2. Search Bar
- `/` to focus
- Fuzzy search project names
- Real-time filter

### 3. Project List (scrollable)
Each row:
- `[▶|󰏤]` Play/Pause indicator
- Type icon: `󰐎` Vercel / `󰣪` Swift / `` CLI
- Project name
- Git counts: `untracked modified issues prs commits`
- Elastic gap
- Action buttons: `󰑢` prod / `` nvim / `󱔘` roadmap / `󱐏` openclaw

### 4. OpenClaw Chat Bar
- `>` prompt
- Direct gateway integration
- Commands execute in project context

### 5. Bottom Status Line (p10k style)
- Total projects
- Total files: ``
- Untracked: ``
- Modified: ``
- Building: `󱫟`
- Failed: `⨻`

## Tech Stack (HEAD)

- **Language:** Go (see `go.mod`)
- **TUI Framework:** Charm Bubble Tea + Bubbles + Lip Gloss
- **Icons:** Nerd Fonts (required)
- **Cache / state:** local under `~/.hustlemc/` (as implemented)
- **Not at HEAD:** Bun, Ink (React for CLI), Zustand — those were an earlier plan fiction

## File Structure (HEAD)

```
mission-control/
├── cmd/mc/main.go          # TUI entrypoint (`go build -o mc-tui ./cmd/mc`)
├── pkg/
│   ├── discover/           # Project discovery
│   ├── openclaw/           # Gateway client
│   └── ui/                 # Bubble Tea UI
├── apps/
│   ├── daemon/             # Companion daemon work
│   └── web/                # Web surface (scaffold)
├── services/               # Supporting services
├── go.mod / go.sum
├── README.md
├── PLAN.md
└── PHASE*.md               # Historical phase writeups (docs)
```

## Phases (honesty)

### Phase 1: Foundation — largely landed as Go TUI
- [x] Go + Bubble Tea project scaffold (`cmd/mc`, `pkg/ui`)
- [x] Basic layout zones (status / search / list / chat / bottom)
- [x] Project discovery hooks
- [ ] Full vim keybinding polish (verify vs HEAD before claiming done)

### Phase 2+: Status integrations / daemon / web
See `PHASE2-*.md` … `PHASE7-*.md` for historical writeups. Treat checkbox claims in those docs as **aspirational** unless verified against `pkg/` and tests.

Open work remains around richer Vercel/Swift/GitHub live status, daemon deployment, and production hardening — do not reintroduce an Ink/React/Zustand tree in PLAN.

*PLAN parity sync: 2026-09-08 — replaced Ink/React/Zustand fiction with Go + Bubble Tea HEAD.*
