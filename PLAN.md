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

## Tech Stack

- **Runtime:** Bun
- **TUI Framework:** Ink (React for CLI)
- **State:** Zustand
- **Icons:** Nerd Fonts (required)
- **Cache:** ~/.hustlemc/

## File Structure

```
mission-control/
├── src/
│   ├── index.tsx           # Entry point
│   ├── app.tsx             # Main app component
│   ├── components/
│   │   ├── StatusTop.tsx   # Top status line
│   │   ├── StatusBottom.tsx
│   │   ├── SearchBar.tsx
│   │   ├── ProjectList.tsx
│   │   ├── ProjectRow.tsx
│   │   ├── ChatBar.tsx     # OpenClaw integration
│   │   └── DetailView.tsx  # Single project view
│   ├── hooks/
│   │   ├── useProjects.ts
│   │   ├── useVercel.ts
│   │   ├── useGit.ts
│   │   └── useGitHub.ts
│   ├── lib/
│   │   ├── discover.ts     # Project discovery
│   │   ├── cache.ts        # ~/.hustlemc/ management
│   │   ├── caddy.ts        # Caddy integration
│   │   └── openclaw.ts     # Gateway client
│   └── store/
│       └── index.ts        # Zustand store
├── package.json
├── tsconfig.json
├── README.md
├── PLAN.md
├── REQUIREMENTS.md
├── STANDARDS.md
└── TODO.md
```

## Phases

### Phase 1: Foundation
- [ ] Project scaffold (Bun + Ink + TypeScript)
- [ ] Basic layout with all 5 zones
- [ ] Vim keybindings (hjkl, gg, G, /search)
- [ ] Project discovery (find .vercel, .xcodeproj, package.json)

### Phase 2: Status Integration
- [ ] Git status crawling
- [ ] Vercel status via `vl`
- [ ] GitHub issues/PRs via `gh`
- [ ] Swift build status

### Phase 3: Actions
- [ ] Play/Pause (bun dev + Caddy hostname)
- [ ] Open in browser
- [ ] Edit docs (nvim README, TODO, PLAN, etc.)
- [ ] OpenClaw TUI launch

### Phase 4: OpenClaw Chat
- [ ] Gateway client integration
- [ ] Context-aware commands (project cwd)
- [ ] Response display

### Phase 5: Polish
- [ ] p10k-style transitions
- [ ] Loading states
- [ ] Error handling
- [ ] Performance optimization

## Data Flow

```
Discovery → Cache → UI
    ↓         ↓
  .hustlemc/  Zustand Store
  projects.json   ↓
                Render
```

## Caching Strategy

- **~/.hustlemc/projects.json** — Project metadata
- **~/.hustlemc/status.json** — Cached statuses (TTL: 30s)
- **${project}/.hustlemc/project.env** — Per-project config
- **${project}/.hustlemc/CONTEXT.md** — AI context

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j/k` | Navigate down/up |
| `h/l` | Collapse/expand or prev/next pane |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `{n}j` | Move down n rows |
| `/` | Focus search |
| `Enter` | Open project detail |
| `o` | Open in browser / build+run |
| `r` | Edit README |
| `R` | Edit ROADMAP |
| `p` | Edit PLAN |
| `t` | Edit TODO |
| `c` | OpenClaw TUI (project cwd) |
| `C` | OpenClaw TUI (parent folder) |
| `q` | Quit / back |
| `Esc` | Clear search / back |

## Success Criteria

1. Zero config — auto-discovers everything
2. Instant startup (<500ms)
3. Real-time status updates
4. Seamless OpenClaw integration
5. Works on Nerd Font terminals only (no fallback)
