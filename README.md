# 🚀 Mission Control

**A p10k-inspired TUI for managing all your projects.**

Mission Control is a terminal dashboard that unifies Vercel deployments, Swift builds, git status, and GitHub activity across your entire portfolio. Zero config. Instant overview. Full keyboard control.

![Status: Phase 1 Complete](https://img.shields.io/badge/status-phase%201%20complete-green)
![Tests: 18 passing](https://img.shields.io/badge/tests-18%20passing-brightgreen)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8)

---

## Quick Start

```bash
# Clone
git clone https://github.com/michaelmonetized/mission-control
cd mission-control

# Build
go build -o mc-tui ./cmd/mc

# Install
ln -sf "$(pwd)/mc-tui" ~/.local/bin/mc

# Run
mc
```

---

## Layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🚀Mission Control   22◬ 2󱫟 8⨻ 3    3󰸞 2✘        42 18 9 14   │
├─────────────────────────────────────────────────────────────────────────┤
│ /                                                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│ [▶] 󰐎 bestwnc.com      12  1  2  2 ..................  󰑢 󱔘  󱐏   │
│ [󰏤] 󰐎 ileague.golf      5  0  1  3 ..................  󰑢 󱔘  󱐏   │
│ [▶] 󰣪 whisper-app        0  0  0  0 ..................  󰑢 󱔘  󱐏   │
│ [󰏤]  nfglyph            3  2  0  1 ..................  󰑢 󱔘  󱐏   │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ >                                                                       │
├─────────────────────────────────────────────────────────────────────────┤
│  86 projects   1.2k   412   89   23󱫟   14⨻                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### Zones

| Zone | Description |
|------|-------------|
| **Top Status** | Aggregated Vercel/Swift/Git stats (p10k style) |
| **Search Bar** | `/` to filter projects |
| **Project List** | Scrollable with vim nav, status icons |
| **Chat Bar** | OpenClaw gateway integration |
| **Bottom Status** | Global totals |

---

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `g/G` | Top/bottom |
| `5j` | Down 5 (vim motions) |
| `Ctrl+d/u` | Page down/up |
| `/` | Search projects |
| `Enter` | Open detail view |
| `o` | Open in nvim |
| `l` | Open lazygit |
| `d` | Open production URL |
| `c` | Launch OpenClaw TUI |
| `r` | Edit README.md |
| `R` | Edit ROADMAP.md |
| `p` | Edit PLAN.md |
| `t` | Edit TODO.md |
| `?` | Show help |
| `Ctrl+r` | Refresh all |
| `q/Esc` | Back/Quit |

---

## Shell Scripts

Mission Control includes a suite of shell scripts for CLI access:

| Script | Purpose |
|--------|---------|
| `mc-discover` | Find all projects in ~/Projects |
| `mc-git-status` | Git status for a project |
| `mc-gh-status` | GitHub issues/PRs count |
| `mc-vl-status` | Vercel deploy status |
| `mc-swift-status` | Swift build status |
| `mc-stats` | Aggregate all stats |
| `mc-cache` | Cache management |
| `mc-dev` | Start/stop dev servers |
| `mc-caddy` | Caddy proxy config |

All scripts support `--json` output:

```bash
./bin/mc-git-status ~/Projects/my-app --json
# {"branch":"main","untracked":2,"modified":1,"staged":0,"ahead":0,"behind":0}
```

---

## Testing

```bash
cd mission-control
./test/run-tests.sh
```

**18 tests** covering all shell scripts and Go integration.

---

## Requirements

- **Go 1.21+** — TUI runtime
- **Nerd Fonts** — Required for icons
- **macOS** — Primary target (Linux untested)
- **CLIs:** `git`, `gh`, `vercel`, `jq`

---

## Project Types

| Type | Icon | Detection |
|------|------|-----------|
| Vercel | 󰐎 | `.vercel/` directory |
| Swift | 󰣪 | `Package.swift` or `*.xcodeproj` |
| CLI | | `package.json` with `bin` field |
| Git | | `.git/` directory |

---

## Configuration

Config stored in `~/.hustlemc/`:

```
~/.hustlemc/
├── config.json      # Settings
├── projects.json    # Discovered projects cache
├── status.json      # Status cache
├── caddy/           # Caddy configs
├── pids/            # Dev server PIDs
└── logs/            # Dev server logs
```

---

## Roadmap

### Phase 1: Local TUI ✅
- [x] Go + Bubble Tea TUI
- [x] Project discovery
- [x] Git/GitHub/Vercel/Swift integration
- [x] Shell script ecosystem
- [x] 18 passing tests
- [ ] OpenClaw chat integration
- [ ] v1.0.0 release

### Phase 2: Cloud Platform
See [PHASE2.md](./PHASE2.md) for the cloud SaaS roadmap:
- GitHub OAuth sign-in
- Connect public/private repos
- BYO Claude Code subscription
- Isolated VPS VMs per user
- Pay-as-you-go compute billing

---

## Documentation

| Doc | Purpose |
|-----|---------|
| [PLAN.md](./PLAN.md) | Architecture & implementation |
| [PHASE2.md](./PHASE2.md) | Cloud platform roadmap |
| [REQUIREMENTS.md](./REQUIREMENTS.md) | Functional requirements |
| [STANDARDS.md](./STANDARDS.md) | Coding conventions |

---

## License

MIT © HurleyUS
