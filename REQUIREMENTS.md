# Mission Control — Requirements

## Functional Requirements

### FR-01: Project Discovery
- **FR-01.1:** Auto-discover projects in configurable root (default: ~/Projects)
- **FR-01.2:** Detect project types:
  - Vercel: `.vercel/` directory exists
  - Swift: `.xcodeproj` or `Package.swift` exists
  - CLI: Executable with shebang or `bin/` directory
  - Node: `package.json` without `.vercel/`
- **FR-01.3:** Recursive search with configurable depth (default: 4)
- **FR-01.4:** Exclude patterns: `node_modules/`, `.git/`, `archive/`, `.templates/`

### FR-02: Status Display
- **FR-02.1:** Top status line shows aggregated counts:
  - Vercel: ready ◬, building 󱫟, queued ⨻, failed 
  - Swift: succeeded 󰸞, failed ✘
  - Git: total , untracked , modified , issues , PRs 
- **FR-02.2:** Bottom status line shows global totals
- **FR-02.3:** Per-project row shows individual counts
- **FR-02.4:** Status refresh interval: 30 seconds (configurable)

### FR-03: Project List
- **FR-03.1:** Scrollable list with vim navigation
- **FR-03.2:** Each row displays:
  - Play/pause state indicator
  - Project type icon
  - Project name
  - Git/GitHub status counts
  - Action button indicators
- **FR-03.3:** Elastic gap fills remaining width
- **FR-03.4:** Selection highlight with cursor

### FR-04: Search
- **FR-04.1:** `/` activates search mode
- **FR-04.2:** Fuzzy matching on project name
- **FR-04.3:** Real-time filtering as you type
- **FR-04.4:** `Esc` clears search and exits search mode
- **FR-04.5:** `Enter` selects first match

### FR-05: Navigation
- **FR-05.1:** Vim keybindings: `hjkl`, `gg`, `G`, `Ctrl+d/u`
- **FR-05.2:** Numeric prefix: `5j` = down 5 rows
- **FR-05.3:** `Enter` opens project detail view
- **FR-05.4:** `q` or `Esc` goes back / quits

### FR-06: Actions
- **FR-06.1:** `o` — Open/run:
  - Vercel: `bun dev`, configure Caddy, open browser
  - Swift: `swift build && swift run`
  - CLI: Show `--help` output
- **FR-06.2:** `󰑢` — Open production URL
- **FR-06.3:** `` — Open in nvim
- **FR-06.4:** `r/R/p/t` — Edit README/ROADMAP/PLAN/TODO
- **FR-06.5:** `c/C` — Launch OpenClaw TUI in project/parent

### FR-07: OpenClaw Integration
- **FR-07.1:** Chat bar at bottom with `>` prompt
- **FR-07.2:** Commands sent to OpenClaw gateway
- **FR-07.3:** Responses displayed inline
- **FR-07.4:** Context includes current project cwd

### FR-08: Caddy Integration
- **FR-08.1:** Auto-generate Caddyfile entries for dev servers
- **FR-08.2:** Pattern: `{project-name}.localhost` → `localhost:{port}`
- **FR-08.3:** Reload Caddy on changes

### FR-09: Caching
- **FR-09.1:** Cache project list in `~/.hustlemc/projects.json`
- **FR-09.2:** Cache status data with 30s TTL
- **FR-09.3:** Per-project config in `.hustlemc/project.env`
- **FR-09.4:** Context file in `.hustlemc/CONTEXT.md`

### FR-10: Detail View
- **FR-10.1:** Title changes to `🚀 mc:${name}`
- **FR-10.2:** Shows full project status
- **FR-10.3:** Recent commits
- **FR-10.4:** Open issues/PRs
- **FR-10.5:** Action buttons with labels

## Non-Functional Requirements

### NFR-01: Performance
- **NFR-01.1:** Startup time < 500ms
- **NFR-01.2:** UI response time < 16ms (60fps)
- **NFR-01.3:** Background status refresh (non-blocking)

### NFR-02: Compatibility
- **NFR-02.1:** macOS only (for now)
- **NFR-02.2:** Requires Nerd Fonts — no fallback
- **NFR-02.3:** Terminal: Ghostty, iTerm2, Kitty, Alacritty

### NFR-03: Dependencies
- **NFR-03.1:** Bun runtime
- **NFR-03.2:** External CLIs: `git`, `gh`, `vl`, `caddy`, `nvim`
- **NFR-03.3:** Optional: `swift`, `openclaw`

### NFR-04: Configuration
- **NFR-04.1:** First-run wizard asks for project root
- **NFR-04.2:** Config stored in `~/.hustlemc/config.json`
- **NFR-04.3:** Zero config after initial setup

### NFR-05: Error Handling
- **NFR-05.1:** Graceful degradation if CLI missing
- **NFR-05.2:** Status shows  if project errored
- **NFR-05.3:** Never crash — show error inline

## Icon Reference

| Icon | Meaning |
|------|---------|
| 🚀 | Mission Control |
| ◬ | Vercel ready |
| 󱫟 | Building/queued |
| ⨻ |  Failed |
| 󰸞 | Swift success |
| ✘ | Swift failed |
| 󰐎 | Vercel project |
| 󰣪 | Swift project |
|  | CLI project |
|  | Files |
|  | Untracked |
|  | Modified |
|  | Issues |
|  | Pull requests |
| 󰑢 | Production link |
|  | Editor |
| 󱔘 | Roadmap |
| 󱐏 | OpenClaw |
| ▶ | Running |
| 󰏤 | Paused |
