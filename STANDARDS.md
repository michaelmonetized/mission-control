# Mission Control — Coding Standards

## Language & Runtime

- **TypeScript** — Strict mode, no `any`
- **Bun** — Runtime and package manager
- **React/Ink** — TUI framework

## File Organization

```
src/
├── index.tsx           # Entry: render <App />
├── app.tsx             # Root component, layout orchestration
├── components/         # UI components (PascalCase.tsx)
├── hooks/              # Custom hooks (useCamelCase.ts)
├── lib/                # Pure utilities (camelCase.ts)
├── store/              # Zustand store
└── types/              # Type definitions (camelCase.ts)
```

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Components | PascalCase | `StatusTop.tsx` |
| Hooks | useCamelCase | `useProjects.ts` |
| Utilities | camelCase | `discover.ts` |
| Constants | SCREAMING_SNAKE | `MAX_DEPTH` |
| Types/Interfaces | PascalCase | `Project`, `VercelStatus` |

## Component Structure

```tsx
// StatusTop.tsx
import { Box, Text } from "ink";
import { useProjects } from "../hooks/useProjects";

interface StatusTopProps {
  viewMode: "list" | "detail";
  projectName?: string;
}

export function StatusTop({ viewMode, projectName }: StatusTopProps) {
  const { stats } = useProjects();
  
  const title = viewMode === "detail" 
    ? `🚀 mc:${projectName}` 
    : "🚀Mission Control";

  return (
    <Box>
      <Text>{title}</Text>
      {/* ... */}
    </Box>
  );
}
```

## State Management

- **Zustand** for global state
- **Local state** for component-specific UI state
- **No prop drilling** — use hooks

```ts
// store/index.ts
import { create } from "zustand";

interface MissionControlStore {
  projects: Project[];
  selectedIndex: number;
  searchQuery: string;
  viewMode: "list" | "detail";
  // Actions
  setSelectedIndex: (i: number) => void;
  setSearchQuery: (q: string) => void;
  openDetail: (project: Project) => void;
}

export const useStore = create<MissionControlStore>((set) => ({
  projects: [],
  selectedIndex: 0,
  searchQuery: "",
  viewMode: "list",
  setSelectedIndex: (i) => set({ selectedIndex: i }),
  setSearchQuery: (q) => set({ searchQuery: q }),
  openDetail: (project) => set({ viewMode: "detail", currentProject: project }),
}));
```

## Icon Usage

- **Always use Nerd Font icons** — no emoji fallbacks (except 🚀 in title)
- **Define icons as constants:**

```ts
// lib/icons.ts
export const ICONS = {
  // Status
  VERCEL_READY: "◬",
  BUILDING: "󱫟",
  FAILED: "⨻",
  SWIFT_OK: "󰸞",
  SWIFT_FAIL: "✘",
  
  // Project types
  TYPE_VERCEL: "󰐎",
  TYPE_SWIFT: "󰣪",
  TYPE_CLI: "",
  
  // Git
  FILES: "",
  UNTRACKED: "",
  MODIFIED: "",
  ISSUES: "",
  PRS: "",
  
  // Actions
  PROD_LINK: "󰑢",
  EDITOR: "",
  ROADMAP: "󱔘",
  OPENCLAW: "󱐏",
  
  // Play state
  PLAYING: "▶",
  PAUSED: "󰏤",
} as const;
```

## Color Palette

p10k style: **black text on colored backgrounds**.

```ts
// lib/colors.ts
export const COLORS = {
  // All foreground text is black
  fg: "black",
  
  // Status segment backgrounds
  title: { bg: "green", fg: "black" },       // 🚀Mission Control
  vercel: { bg: "yellow", fg: "black" },     // ◬ 󱫟 ⨻
  swift: { bg: "magenta", fg: "black" },     // 󰸞 ✘
  git: { bg: "cyan", fg: "black" },          //   
  
  // Project states (background colors)
  ready: { bg: "green", fg: "black" },
  building: { bg: "blue", fg: "black" },
  queued: { bg: "gray", fg: "black" },
  failed: { bg: "red", fg: "black" },
  
  // UI
  selected: { bg: "cyan", fg: "black" },
  prompt: { bg: "green", fg: "black" },
  search: { bg: "yellow", fg: "black" },
} as const;
```

**Rule:** Colored backgrounds, black text — always.

## Async Patterns

- **Non-blocking** — Never freeze UI
- **Background refresh** — Use intervals
- **Graceful errors** — Catch and display

```ts
// hooks/useVercelStatus.ts
export function useVercelStatus() {
  const [status, setStatus] = useState<VercelStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetch = async () => {
      try {
        const result = await $`vl --json`.text();
        setStatus(JSON.parse(result));
        setError(null);
      } catch (e) {
        setError("vl failed");
      }
    };

    fetch();
    const interval = setInterval(fetch, 30_000);
    return () => clearInterval(interval);
  }, []);

  return { status, error };
}
```

## Shell Commands

Use Bun's `$` shell:

```ts
import { $ } from "bun";

// Good
const output = await $`git status --porcelain`.text();

// Bad
const output = execSync("git status --porcelain").toString();
```

## Testing

- Unit tests for `lib/` utilities
- Integration tests for hooks
- Manual testing for TUI components

```ts
// lib/discover.test.ts
import { expect, test } from "bun:test";
import { discoverProjects } from "./discover";

test("discovers vercel projects", async () => {
  const projects = await discoverProjects("./fixtures");
  expect(projects.some(p => p.type === "vercel")).toBe(true);
});
```

## Error Messages

- Use Nerd Font icons:  for errors,  for warnings
- Keep messages short
- Include actionable info

```ts
// Bad
console.error("An error occurred while trying to fetch the Vercel status");

// Good
log(" vl failed — is Vercel CLI installed?");
```

## Git Commits

Follow conventional commits:

```
feat(discovery): add swift project detection
fix(status): handle vl timeout
refactor(store): simplify project state
docs(readme): add installation steps
```

## Performance Rules

1. **Lazy load** — Don't fetch status until visible
2. **Debounce search** — 150ms delay
3. **Virtualize list** — Only render visible rows
4. **Cache aggressively** — 30s TTL for status data
5. **Batch updates** — Use Zustand's `set` properly
