<script lang="ts">
  // Console layout: one row of games over full-screen art. Down shows the
  // selected game's details.
  import GameArt from "../components/GameArt.svelte";
  import Icon from "../components/Icon.svelte";
  import Logo from "../components/Logo.svelte";
  import { metaLine, padSummary, recentFirst } from "../lib/bp";
  import { bytes } from "../lib/format";
  import { feedback, pad, useInput } from "../lib/input.svelte";
  import { isFresh, lib } from "../lib/store.svelte";
  import { title, type Game } from "../lib/types";
  import Hints from "./Hints.svelte";

  type Props = {
    height: number;
    onplay: (g: Game) => void;
    oninfo: (g: Game) => void;
    onfocus: (g: Game | null) => void;
    onlibrary: () => void;
    onsearch: () => void;
    onsettings: () => void;
    onfound: () => void;
    onmenu: () => void;
    ondesktop: () => void;
  };
  let p: Props = $props();

  const games = $derived(recentFirst(lib.base).slice(0, 40));
  let i = $state(0);
  let details = $state(false);
  const n = $derived(games.length + 1); // + "All games"
  const g = $derived(games[i] ?? null);
  $effect(() => p.onfocus(g));
  $effect(() => {
    if (i >= n) i = Math.max(0, n - 1);
  });

  // Keep the last two heroes mounted so they can crossfade.
  let heroes = $state<Game[]>([]);
  $effect(() => {
    if (g && heroes[heroes.length - 1]?.id !== g.id) heroes = [...heroes.filter((h) => h.id !== g.id).slice(-1), g];
  });

  $effect(() =>
    useInput((intent) => {
      switch (intent) {
        case "left":
          if (i > 0) ((i -= 1), feedback.move());
          return;
        case "right":
          if (i < n - 1) ((i += 1), feedback.move());
          return;
        case "down":
          if (g && !details) ((details = true), feedback.move());
          return;
        case "up":
          if (details) ((details = false), feedback.move());
          return;
        case "back":
          if (details) {
            details = false;
            feedback.move();
            return;
          }
          return false;
        case "info":
          if (g) ((details = !details), feedback.move());
          return;
        case "confirm":
          if (g) p.onplay(g);
          else (feedback.confirm(), p.onlibrary());
          return;
      }
      return false;
    }),
  );

  const TILE = 172;
  const GAP = 20;
  const rowX = $derived(-Math.max(0, i - 1) * (TILE + GAP));
  const logoSize = $derived(g ? Math.round(Math.min(130, 1500 / (title(g).length * 0.6))) : 100);
  let logoFailed = $state(false);
  $effect(() => {
    g?.id;
    logoFailed = false;
  });
  const clock = $state({ v: "" });
  $effect(() => {
    const tick = () => {
      const d = new Date();
      clock.v = `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
    };
    tick();
    const t = setInterval(tick, 15000);
    return () => clearInterval(t);
  });
</script>

<div class="console">
  <div class="heroes">
    {#each heroes as h (h.id)}
      <div class="hero" class:show={h.id === g?.id}><GameArt game={h} kind="hero" /></div>
    {/each}
  </div>
  <div class="scrim-l"></div>
  <div class="scrim-v"></div>
  <div class="dim" class:on={details}></div>

  <div class="top">
    <Logo size={32} />
    <span class="tab on">Home</span>
    <button type="button" class="tab" onclick={p.onlibrary}>Library</button>
    <button type="button" class="tab" onclick={p.onsearch}>Search</button>
    <div class="grow"></div>
    {#if pad.connected}<span class="status"><Icon name="pad" size={26} stroke={1.8} />{pad.battery >= 0 ? `${pad.battery}%` : ""}</span>{/if}
    <span class="clock">{clock.v}</span>
  </div>

  <div class="slide" style:transform="translateY({details ? -300 : 0}px)">
    <div class="tiles" class:hide={details}>
      <div class="strip" style:transform="translateX({rowX}px)">
        {#each games as t, k (t.id)}
          {@const on = k === i}
          <button type="button" class="tile" class:on style:width="{on ? 400 : TILE}px" style:height="{on ? 225 : TILE}px" onclick={() => (on ? p.onplay(t) : (i = k))} aria-label={title(t)}>
            <GameArt game={t} kind={on ? "hero" : "cover"} />
            {#if on}<span class="tname">{title(t)}</span>{/if}
            {#if isFresh(t)}<span class="new">NEW</span>{/if}
          </button>
        {/each}
        <button type="button" class="tile all" class:on={i === games.length} style:width="{TILE}px" style:height="{TILE}px" onclick={p.onlibrary}>
          <Icon name="grid" size={36} stroke={1.7} />
          <span>All games</span>
        </button>
      </div>
    </div>

    {#if g}
      <div class="info">
        <span class="pill">{g.sourceLabel}</span>
        {#if g.meta?.logo && !logoFailed}
          <img class="logo" src={g.meta.logo} alt={title(g)} onerror={() => (logoFailed = true)} />
        {:else}
          <div class="title" style:font-size="{logoSize}px">{title(g)}</div>
        {/if}
        <div class="line">{metaLine(g)}{g.meta?.genres?.length ? ` · ${g.meta.genres.slice(0, 2).join(", ")}` : ""}</div>
        <div class="chips">
          <span class="chip"><Icon name="pad" size={22} stroke={1.8} />{padSummary(g).short}</span>
          {#if g.favorite}<span class="chip"><Icon name="star" size={20} />Favorite</span>{/if}
          {#if g.needsReview}<span class="chip warn"><Icon name="warn" size={20} />Needs a check</span>{/if}
        </div>
        <div class="buttons">
          <button type="button" class="play" onclick={() => p.onplay(g)}>
            <span class="pic"><Icon name="play" size={22} /></span>
            Play
          </button>
          <button type="button" class="round" aria-label="Details" onclick={() => (details = !details)}><Icon name="info" size={30} stroke={1.8} /></button>
        </div>
      </div>

      <div class="cards" class:show={details}>
        <div class="card">
          <div class="ch"><Icon name="info" size={22} stroke={1.8} /><span>About</span></div>
          <div class="big">{g.meta?.developers?.[0] ?? g.sourceLabel}{g.meta?.releaseYear ? ` · ${g.meta.releaseYear}` : ""}</div>
          <p>{g.meta?.description ?? "No store description for this game."}</p>
        </div>
        <div class="card">
          <div class="ch"><Icon name="pad" size={24} stroke={1.8} /><span>Controller</span></div>
          <div class="big">{padSummary(g).short}</div>
          <p>{padSummary(g).long}</p>
        </div>
        <div class="card">
          <div class="ch"><Icon name="scan" size={22} stroke={1.8} /><span>This copy</span></div>
          <dl>
            <dt>Found</dt>
            <dd>{g.how}</dd>
            <dt>Identified</dt>
            <dd>{g.matchHow}</dd>
            {#if g.sizeBytes}<dt>Size</dt><dd>{bytes(g.sizeBytes)}</dd>{/if}
          </dl>
        </div>
      </div>
    {/if}
  </div>

  <div class="hints">
    <Hints
      hints={details
        ? [
            { button: "confirm", label: "Play" },
            { button: "back", label: "Back" },
          ]
        : [
            { button: "confirm", label: g ? "Play" : "Open" },
            { button: "info", label: "Details" },
            { button: "view", label: "Search" },
            { button: "menu", label: "Quick access" },
          ]}
    />
  </div>
</div>

<style>
  .console {
    position: absolute;
    inset: 0;
    overflow: hidden;
    background: #06080b;
    color: #f3f5f7;
  }
  .heroes {
    position: absolute;
    inset: 0;
  }
  .hero {
    position: absolute;
    inset: 0;
    opacity: 0;
    transition: opacity 0.7s ease;
  }
  .hero.show {
    opacity: 1;
    animation: kb 16s ease-out both;
  }
  @keyframes kb {
    from {
      transform: scale(1.07);
    }
  }
  .scrim-l {
    position: absolute;
    inset: 0;
    background: linear-gradient(90deg, rgba(6, 8, 11, 0.94) 0%, rgba(6, 8, 11, 0.62) 36%, rgba(6, 8, 11, 0) 68%);
  }
  .scrim-v {
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, rgba(6, 8, 11, 0.82) 0%, rgba(6, 8, 11, 0) 30%, rgba(6, 8, 11, 0) 60%, rgba(6, 8, 11, 0.92) 100%);
  }
  .dim {
    position: absolute;
    inset: 0;
    background: rgba(6, 8, 11, 0.72);
    opacity: 0;
    transition: opacity 0.5s;
  }
  .dim.on {
    opacity: 1;
  }
  .top {
    position: absolute;
    left: 96px;
    right: 96px;
    top: 34px;
    height: 60px;
    display: flex;
    align-items: center;
    gap: 10px;
    z-index: 5;
  }
  .tab {
    height: 48px;
    padding: 0 24px;
    border: 0;
    border-radius: 24px;
    background: transparent;
    color: rgba(243, 245, 247, 0.72);
    font-size: 20px;
    font-weight: 600;
    display: flex;
    align-items: center;
  }
  .tab.on {
    background: rgba(243, 245, 247, 0.16);
    color: #f3f5f7;
    font-weight: 700;
    margin-left: 26px;
  }
  .grow {
    flex: 1;
  }
  .status {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
    margin-right: 20px;
  }
  .clock {
    font-size: 24px;
    font-weight: 700;
  }
  .slide {
    position: absolute;
    inset: 0;
    transition: transform 0.6s cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  .tiles {
    position: absolute;
    left: 96px;
    right: 0;
    top: 128px;
    height: 240px;
    transition: opacity 0.5s;
  }
  .tiles.hide {
    opacity: 0;
  }
  .strip {
    display: flex;
    gap: 20px;
    align-items: flex-start;
    transition: transform 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  .tile {
    position: relative;
    flex-shrink: 0;
    padding: 0;
    border: 0;
    border-radius: 22px;
    overflow: hidden;
    background: #11151b;
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.45);
    transition:
      width 0.45s cubic-bezier(0.2, 0.8, 0.2, 1),
      height 0.45s cubic-bezier(0.2, 0.8, 0.2, 1),
      box-shadow 0.3s;
  }
  .tile.on {
    box-shadow:
      0 0 0 4px #f3f5f7,
      0 20px 50px rgba(0, 0, 0, 0.55),
      0 0 70px color-mix(in oklab, var(--accent-game) 45%, transparent);
  }
  .tname {
    position: absolute;
    left: 20px;
    right: 20px;
    bottom: 16px;
    text-align: left;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 30px;
    color: #fff;
    text-shadow: 0 2px 16px rgba(0, 0, 0, 0.7);
  }
  .new {
    position: absolute;
    top: 12px;
    right: 12px;
    padding: 4px 10px;
    border-radius: 999px;
    background: #f3f5f7;
    color: #06080b;
    font-size: 13px;
    font-weight: 800;
    letter-spacing: 0.06em;
  }
  .tile.all {
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(18, 22, 28, 0.6);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #f3f5f7;
    font-size: 18px;
    font-weight: 700;
  }
  .info {
    position: absolute;
    left: 96px;
    top: 436px;
    width: 1040px;
    display: flex;
    flex-direction: column;
    gap: 22px;
  }
  .pill {
    width: fit-content;
    padding: 7px 14px;
    border-radius: 999px;
    border: 1px solid rgba(255, 255, 255, 0.24);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .logo {
    max-width: 800px;
    max-height: 230px;
    object-fit: contain;
    object-position: left bottom;
    filter: drop-shadow(0 6px 30px rgba(0, 0, 0, 0.55));
  }
  .title {
    font-family: var(--font-display);
    font-weight: 700;
    line-height: 0.95;
    max-width: 1000px;
    text-wrap: balance;
    text-shadow: 0 4px 30px rgba(0, 0, 0, 0.45);
  }
  .line {
    font-size: 22px;
    color: rgba(243, 245, 247, 0.82);
  }
  .chips {
    display: flex;
    gap: 12px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    height: 44px;
    padding: 0 18px;
    border-radius: 999px;
    background: rgba(14, 18, 24, 0.62);
    border: 1px solid rgba(255, 255, 255, 0.1);
    font-size: 18px;
    font-weight: 600;
  }
  .chip.warn {
    color: #ffd28a;
  }
  .buttons {
    display: flex;
    gap: 16px;
    margin-top: 6px;
  }
  .play {
    height: 80px;
    padding: 0 46px 0 14px;
    border-radius: 40px;
    border: 0;
    background: #f3f5f7;
    color: #06080b;
    display: flex;
    align-items: center;
    gap: 18px;
    font-size: 28px;
    font-weight: 800;
    box-shadow: 0 14px 44px color-mix(in oklab, var(--accent-game) 45%, transparent);
  }
  .pic {
    width: 54px;
    height: 54px;
    border-radius: 50%;
    background: #06080b;
    color: #f3f5f7;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .round {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(14, 18, 24, 0.55);
    color: #f3f5f7;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .cards {
    position: absolute;
    left: 96px;
    right: 96px;
    top: 950px;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 24px;
    opacity: 0;
    transition: opacity 0.5s;
  }
  .cards.show {
    opacity: 1;
  }
  .card {
    height: 300px;
    padding: 28px;
    border-radius: 28px;
    background: rgba(14, 18, 24, 0.74);
    border: 1px solid rgba(255, 255, 255, 0.09);
    display: flex;
    flex-direction: column;
    gap: 10px;
    overflow: hidden;
  }
  .ch {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 18px;
    font-weight: 700;
    color: rgba(243, 245, 247, 0.78);
  }
  .big {
    font-size: 30px;
    font-weight: 800;
  }
  .card p {
    margin: 0;
    font-size: 18px;
    line-height: 1.45;
    color: rgba(243, 245, 247, 0.78);
    display: -webkit-box;
    -webkit-line-clamp: 5;
    line-clamp: 5;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  dl {
    display: grid;
    grid-template-columns: 120px minmax(0, 1fr);
    gap: 9px 12px;
    margin: 0;
    font-size: 17px;
  }
  dt {
    color: rgba(243, 245, 247, 0.6);
  }
  dd {
    margin: 0;
  }
  .hints {
    position: absolute;
    right: 96px;
    bottom: 40px;
    left: 96px;
    z-index: 5;
  }
</style>
