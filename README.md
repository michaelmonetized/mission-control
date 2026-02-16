# 🚀 Mission Control

**A p10k-inspired TUI for managing all your projects.**

Mission Control is a terminal dashboard that unifies Vercel deployments, Swift builds, git status, and GitHub activity across your entire portfolio. Zero config. Instant overview. Full keyboard control.

![Status: Planning](https://img.shields.io/badge/status-planning-blue)

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
| **Project List** | Scrollable with vim nav, status icons, action buttons |
| **Chat Bar** | Direct OpenClaw gateway integration |
| **Bottom Status** | Global totals |

### Colors (p10k style)

**Black text on colored backgrounds — always.**

| Segment | Background |
|---------|------------|
| Title | Green |
| Vercel | Yellow |
| Swift | Magenta |
| Git | Cyan |

---

## Features

### 󰐎 Project Types
- **Vercel** — Next.js apps with deploy status
- **󰣪 Swift** — iOS/macOS apps with build status
- ** CLI** — Command-line tools

### 󰊤 Status Integration
- Git: untracked, modified, commits
- GitHub: issues, pull requests
- Vercel: ready, building, failed
- Swift: build success/failure

### ⌨️ Vim Keybindings
| Key | Action |
|-----|--------|
| `j/k` | Navigate |
| `gg/G` | Top/bottom |
| `5j` | Down 5 |
| `/` | Search |
| `Enter` | Open detail |
| `o` | Run/open browser |
| `r/R/p/t` | Edit docs |
| `c` | OpenClaw TUI |
| `q` | Quit |

### 󱐏 OpenClaw Integration
Chat bar sends commands to OpenClaw gateway with project context.

### 󰒍 Caddy Integration  
Auto-generates `*.localhost` hostnames for dev servers.

---

## Requirements

- **Bun** — Runtime
- **Nerd Fonts** — Required for icons (no fallback)
- **macOS** — Primary target
- **CLIs:** `git`, `gh`, `vl`, `nvim`, `caddy`

---

## Installation

```bash
# Clone
git clone https://github.com/michaelmonetized/mission-control
cd mission-control

# Install
bun install

# Run
bun start
```

---

## Configuration

On first run, Mission Control asks for your project root (default: `~/Projects`).

Config stored in `~/.hustlemc/config.json`.

---

## Documentation

| Doc | Purpose |
|-----|---------|
| [PLAN.md](./PLAN.md) | Architecture & implementation phases |
| [REQUIREMENTS.md](./REQUIREMENTS.md) | Functional & non-functional requirements |
| [STANDARDS.md](./STANDARDS.md) | Coding conventions |
| [TODO.md](./TODO.md) | Task tracking |

---

## Status Icons Reference

| Icon | Meaning |
|------|---------|
| ◬ | Vercel ready |
| 󱫟 | Building |
| ⨻ | Failed |
| 󰸞 | Swift success |
| ✘ | Swift failed |
| 󰐎 | Vercel project |
| 󰣪 | Swift project |
|  | CLI project |
|  | Files |
|  | Untracked |
|  | Modified |
|  | Issues |
|  | PRs |
| ▶ | Running |
| 󰏤 | Paused |

---

## Title States

| Context | Title |
|---------|-------|
| List view | `🚀Mission Control` |
| Detail view | `🚀 mc:${project-name}` |

---

## License

MIT © HurleyUS
