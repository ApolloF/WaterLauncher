<script lang="ts">
  // A game's page in big picture (△ / Y on a game): details and the few
  // choices worth making from the couch.
  import GameArt from "../components/GameArt.svelte";
  import Icon from "../components/Icon.svelte";
  import { api } from "../lib/api";
  import { metaLine, padSummary } from "../lib/bp";
  import { feedback, useInput } from "../lib/input.svelte";
  import { lib } from "../lib/store.svelte";
  import { savesSummary } from "../lib/saves";
  import { title, type Game, type Saves } from "../lib/types";
  import Hints from "./Hints.svelte";
  import { clamp } from "./nav";

  let { game, onplay, onclose }: { game: Game; onplay: () => void; onclose: () => void } = $props();

  const modes = ["", "native", "steam"] as const;
  const modeLabel: Record<string, string> = { "": "Auto", native: "Native", steam: "Steam Input" };

  let b = $state(0);
  let logoFailed = $state(false);

  let saves = $state<Saves | null>(null);
  const savesInfo = $derived(savesSummary(saves));
  $effect(() => {
    const id = game.id;
    let live = true;
    if (game.installed) api.saves.get(id).then((s) => live && (saves = s)).catch(() => {});
    return () => (live = false);
  });
  const buttons = $derived([
    { id: "play", label: game.installed ? "Play" : "Not installed" },
    { id: "fav", label: game.favorite ? "Favorite" : "Add to favorites" },
    { id: "pad", label: `Controller: ${modeLabel[game.padMode ?? ""]}` },
  ]);

  function press(id: string) {
    feedback.confirm();
    if (id === "play" && game.installed) onplay();
    else if (id === "fav") lib.run(() => api.setFavorite(game.id, !game.favorite));
    else if (id === "pad") {
      const next = modes[(modes.indexOf((game.padMode ?? "") as (typeof modes)[number]) + 1) % modes.length];
      lib.run(() => api.setPadMode(game.id, next));
    }
  }

  $effect(() =>
    useInput((i) => {
      if (i === "left" || i === "right") {
        const j = clamp(b + (i === "left" ? -1 : 1), 0, buttons.length - 1);
        if (j !== b) {
          b = j;
          feedback.move();
        }
      } else if (i === "confirm") press(buttons[b].id);
      else if (i === "back" || i === "info") onclose();
      else if (i === "menu" || i === "home") return false;
    }),
  );
  const pad = $derived(padSummary(game));
  const facts = $derived(
    [game.meta?.developers?.[0], game.meta?.releaseYear ? String(game.meta.releaseYear) : "", game.meta?.genres?.slice(0, 3).join(" · ")].filter(Boolean).join("  ·  "),
  );
</script>

<div class="sheet">
  <div class="art"><GameArt {game} kind="hero" /></div>
  <div class="shade"></div>
  <div class="content">
    <span class="src">{game.sourceLabel}</span>
    {#if game.meta?.logo && !logoFailed}
      <img class="logo" src={game.meta.logo} alt={title(game)} onerror={() => (logoFailed = true)} />
    {:else}
      <h1>{title(game)}</h1>
    {/if}
    <div class="line">{metaLine(game)}{facts ? `  ·  ${facts}` : ""}</div>
    {#if game.meta?.description}<p class="desc">{game.meta.description}</p>{/if}
    <div class="buttons">
      {#each buttons as bt, k (bt.id)}
        <button type="button" class="btn" class:primary={bt.id === "play"} class:on={k === b} onclick={() => ((b = k), press(bt.id))}>
          {#if bt.id === "play"}<Icon name="play" size={22} />{:else if bt.id === "fav"}<Icon name="star" size={22} />{:else}<Icon name="pad" size={24} stroke={1.8} />{/if}
          {bt.label}
        </button>
      {/each}
    </div>
    <div class="note">{pad.long}</div>
    {#if savesInfo && saves?.installed}
      <div class="note saves" class:warn={savesInfo.tone === "warn"}>Saves: {savesInfo.text}</div>
    {/if}
  </div>
  <div class="hints"><Hints hints={[{ button: "confirm", label: "Select" }, { button: "back", label: "Back" }]} /></div>
</div>

<style>
  .sheet {
    position: absolute;
    inset: 0;
    z-index: 25;
    background: #06080b;
    color: #f3f5f7;
    animation: in 0.3s ease both;
  }
  @keyframes in {
    from {
      opacity: 0;
      transform: scale(1.01);
    }
  }
  .art {
    position: absolute;
    inset: 0;
  }
  .shade {
    position: absolute;
    inset: 0;
    background:
      linear-gradient(90deg, rgba(6, 8, 11, 0.95) 0%, rgba(6, 8, 11, 0.7) 40%, rgba(6, 8, 11, 0.1) 75%),
      linear-gradient(0deg, rgba(6, 8, 11, 0.9) 0%, rgba(6, 8, 11, 0) 50%);
  }
  .content {
    position: absolute;
    left: 110px;
    bottom: 150px;
    width: 1000px;
    display: flex;
    flex-direction: column;
    gap: 22px;
  }
  .src {
    width: fit-content;
    padding: 7px 14px;
    border-radius: 999px;
    border: 1px solid rgba(255, 255, 255, 0.24);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 110px;
    line-height: 0.95;
  }
  .logo {
    max-width: 720px;
    max-height: 240px;
    object-fit: contain;
    object-position: left bottom;
    filter: drop-shadow(0 6px 30px rgba(0, 0, 0, 0.6));
  }
  .line {
    font-size: 22px;
    color: rgba(243, 245, 247, 0.82);
  }
  .desc {
    margin: 0;
    font-size: 21px;
    line-height: 1.5;
    color: rgba(243, 245, 247, 0.75);
    max-width: 860px;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .buttons {
    display: flex;
    gap: 16px;
    margin-top: 6px;
  }
  .btn {
    height: 72px;
    padding: 0 30px;
    border-radius: 36px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(14, 18, 24, 0.6);
    color: #f3f5f7;
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 22px;
    font-weight: 700;
  }
  .btn.primary {
    background: #f3f5f7;
    color: #06080b;
    border-color: transparent;
    padding: 0 40px;
  }
  .btn.on {
    box-shadow:
      0 0 0 4px #06080b,
      0 0 0 7px #f3f5f7;
  }
  .note {
    font-size: 18px;
    color: rgba(243, 245, 247, 0.6);
  }
  .note.saves {
    margin-top: -12px;
  }
  .note.warn {
    color: #ffd28a;
  }
  .hints {
    position: absolute;
    right: 96px;
    bottom: 48px;
  }
</style>
