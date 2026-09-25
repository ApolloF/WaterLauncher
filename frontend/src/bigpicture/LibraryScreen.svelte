<script lang="ts">
  // Every game as a cover grid, with tabs switched by L1/R1. Rows outside
  // the view aren't rendered.
  import GameArt from "../components/GameArt.svelte";
  import { byRecent, byTitle, newFinds } from "../lib/bp";
  import { feedback, useInput } from "../lib/input.svelte";
  import { lib } from "../lib/store.svelte";
  import { title, type Game } from "../lib/types";
  import Hints from "./Hints.svelte";
  import { gridStep } from "./nav";

  let {
    height,
    games: only,
    heading = "Library",
    onplay,
    oninfo,
    onfocus,
    onback,
  }: {
    height: number;
    games?: Game[];
    heading?: string;
    onplay: (g: Game) => void;
    oninfo: (g: Game) => void;
    onfocus: (g: Game | null) => void;
    onback: () => void;
  } = $props();

  const tabs = [
    { id: "all", label: "All" },
    { id: "fav", label: "Favorites" },
    { id: "recent", label: "Recently played" },
    { id: "new", label: "New" },
  ] as const;
  let tab = $state(0);
  const games = $derived.by(() => {
    if (only) return only;
    const base = lib.base.filter((g) => g.installed);
    switch (tabs[tab].id) {
      case "fav":
        return base.filter((g) => g.favorite).sort(byTitle);
      case "recent":
        return base.filter((g) => (g.lastPlayed ?? 0) > 0 || (g.storeLastPlayed ?? 0) > 0).sort(byRecent);
      case "new":
        return newFinds(base, 100);
      default:
        return [...base].sort(byTitle);
    }
  });

  const COLS = 7;
  const GAP = 26;
  const W = (1920 - 2 * 110 - GAP * (COLS - 1)) / COLS;
  const H = W * 1.5;
  const ROW = H + 64;
  let i = $state(0);
  $effect(() => {
    games;
    i = Math.min(i, Math.max(0, games.length - 1));
  });
  const row = $derived(Math.floor(i / COLS));
  const visibleRows = $derived(Math.ceil((height - 260) / ROW));
  const firstRow = $derived(Math.max(0, Math.min(row - 1, Math.ceil(games.length / COLS) - visibleRows)));
  const shown = $derived(games.slice(Math.max(0, firstRow - 1) * COLS, (firstRow + visibleRows + 1) * COLS).map((g, k) => ({ g, k: Math.max(0, firstRow - 1) * COLS + k })));
  $effect(() => onfocus(games[i] ?? null));

  $effect(() =>
    useInput((intent) => {
      const g = games[i];
      switch (intent) {
        case "lb":
        case "rb":
          if (only) return;
          tab = (tab + (intent === "lb" ? tabs.length - 1 : 1)) % tabs.length;
          i = 0;
          feedback.move();
          return;
        case "confirm":
          if (g) onplay(g);
          return;
        case "info":
          if (g) oninfo(g);
          return;
        case "back":
          onback();
          return;
        case "up":
        case "down":
        case "left":
        case "right": {
          const j = gridStep(i, games.length, COLS, intent);
          if (j !== null) {
            i = j;
            feedback.move();
          } else if (intent === "up" && row === 0) return;
          return;
        }
      }
      return false;
    }),
  );
</script>

<div class="lib">
  <div class="head">
    <h1>{heading}</h1>
    <span class="count">{games.length} {games.length === 1 ? "game" : "games"}</span>
    {#if !only}
      <div class="tabs">
        {#each tabs as t, k (t.id)}
          <button type="button" class:on={k === tab} onclick={() => ((tab = k), (i = 0))}>{t.label}</button>
        {/each}
      </div>
    {/if}
  </div>
  {#if games.length === 0}
    <p class="empty">{tabs[tab].id === "fav" ? "No favorites yet. Press △ on a game to add it." : "Nothing here yet."}</p>
  {/if}
  <div class="grid" style:transform="translateY({-firstRow * ROW}px)" style:height="{Math.ceil(games.length / COLS) * ROW}px">
    {#each shown as { g, k } (g.id)}
      <button
        type="button"
        class="cell"
        class:on={k === i}
        style:width="{W}px"
        style:height="{H}px"
        style:left="{110 + (k % COLS) * (W + GAP)}px"
        style:top="{Math.floor(k / COLS) * ROW}px"
        onclick={() => (k === i ? onplay(g) : (i = k))}
        aria-label={title(g)}
      >
        <GameArt game={g} kind="cover" />
        {#if !g.meta?.cover}<span class="ct">{title(g)}</span>{/if}
      </button>
      <span class="name" class:on={k === i} style:width="{W}px" style:left="{110 + (k % COLS) * (W + GAP)}px" style:top="{Math.floor(k / COLS) * ROW + H + 12}px">{title(g)}</span>
    {/each}
  </div>
  <div class="hints">
    <Hints
      hints={[
        { button: "confirm", label: "Play" },
        { button: "info", label: "Details" },
        ...(only ? [] : [{ button: "rb" as const, label: "Tabs" }]),
        { button: "back", label: "Back" },
      ]}
    />
  </div>
</div>

<style>
  .lib {
    position: absolute;
    inset: 0;
    background: #0a0e13;
    color: #e8edf2;
    overflow: hidden;
  }
  .head {
    position: absolute;
    left: 110px;
    right: 110px;
    top: 50px;
    display: flex;
    align-items: baseline;
    gap: 18px;
    z-index: 2;
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 52px;
  }
  .count {
    font-size: 22px;
    color: #9ba8b5;
  }
  .tabs {
    margin-left: auto;
    display: flex;
    gap: 6px;
  }
  .tabs button {
    height: 48px;
    padding: 0 22px;
    border: 0;
    border-radius: 24px;
    background: transparent;
    color: #9ba8b5;
    font-size: 20px;
    font-weight: 700;
  }
  .tabs button.on {
    background: rgba(79, 209, 232, 0.16);
    color: oklch(0.88 0.09 205);
  }
  .grid {
    position: absolute;
    left: 0;
    right: 0;
    top: 150px;
    transition: transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  .cell {
    position: absolute;
    padding: 0;
    border: 0;
    border-radius: 14px;
    overflow: hidden;
    background: #121a23;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
    transition:
      transform 0.22s cubic-bezier(0.2, 0.8, 0.2, 1),
      box-shadow 0.22s;
  }
  .cell.on {
    transform: scale(1.06);
    box-shadow:
      0 0 0 4px #fff,
      0 18px 44px rgba(0, 0, 0, 0.55);
    z-index: 1;
  }
  .ct {
    position: absolute;
    left: 12px;
    right: 12px;
    top: 18px;
    text-align: center;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 30px;
    line-height: 1;
    color: #fff;
    text-shadow: 0 2px 12px rgba(0, 0, 0, 0.5);
  }
  .name {
    position: absolute;
    font-size: 19px;
    font-weight: 600;
    color: #9ba8b5;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .name.on {
    color: #fff;
  }
  .empty {
    position: absolute;
    left: 110px;
    top: 170px;
    font-size: 24px;
    color: #9ba8b5;
  }
  .hints {
    position: absolute;
    right: 96px;
    bottom: 40px;
    left: 110px;
    z-index: 2;
  }
</style>
