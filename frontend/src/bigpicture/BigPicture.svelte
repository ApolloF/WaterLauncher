<script lang="ts">
  import { accentOf } from "../lib/bp";
  import { dispatch, feedback, keyIntent, setBase, setLight, toHex } from "../lib/input.svelte";
  import { lib } from "../lib/store.svelte";
  import type { Game } from "../lib/types";
  import BPSettings from "./BPSettings.svelte";
  import Console from "./Console.svelte";
  import Deck from "./Deck.svelte";
  import GameSheet from "./GameSheet.svelte";
  import Launching from "./Launching.svelte";
  import LibraryScreen from "./LibraryScreen.svelte";
  import Orbit from "./Orbit.svelte";
  import QuickAccess from "./QuickAccess.svelte";
  import SearchScreen from "./SearchScreen.svelte";
  import Stage from "./Stage.svelte";

  let { onexit }: { onexit: () => void } = $props();

  type Screen = "home" | "library" | "search" | "settings" | "found";
  let screen = $state<Screen>("home");
  let qa = $state(false);
  let sheetId = $state<number | null>(null);
  let launchId = $state<number | null>(null);
  let focused = $state<Game | null>(null);

  const layout = $derived(lib.settings?.bigPictureLayout ?? "deck");
  const sheet = $derived(sheetId === null ? null : (lib.games.find((g) => g.id === sheetId) ?? null));
  const launching = $derived(launchId === null ? null : (lib.games.find((g) => g.id === launchId) ?? null));
  const found = $derived(lib.base.filter((g) => g.installed && (g.needsReview || (!g.initial && !g.lastPlayed && Date.now() / 1000 - g.addedAt < 7 * 86400))));
  const accent = $derived(accentOf(focused) ?? "oklch(0.8 0.12 205)");
  const lightHex = $derived(toHex(accent));

  $effect(() => setLight(accent));

  const go = (s: Screen) => {
    screen = s;
    qa = false;
    feedback.move();
  };
  const play = (g: Game) => {
    if (!g.installed) return feedback.error();
    sheetId = null;
    launchId = g.id;
  };
  const info = (g: Game) => {
    sheetId = g.id;
    feedback.move();
  };
  const onfocus = (g: Game | null) => (focused = g);

  // Under every screen: the buttons that work everywhere.
  $effect(() => {
    setBase((i) => {
      if (i === "menu" || i === "home") {
        qa = !qa;
        feedback.move();
      } else if (i === "view") go("search");
      else if (i === "back" && screen !== "home") go("home");
    });
    return () => setBase(null);
  });

  // Big picture is always dark.
  $effect(() => {
    const prev = document.documentElement.dataset.theme;
    document.documentElement.dataset.theme = "dark";
    return () => {
      if (prev) document.documentElement.dataset.theme = prev;
    };
  });

  function onkeydown(e: KeyboardEvent) {
    const i = keyIntent(e);
    if (!i) return;
    e.preventDefault();
    dispatch(i, e.repeat);
  }

  const layoutProps = $derived({
    onplay: play,
    oninfo: info,
    onfocus,
    onlibrary: () => go("library"),
    onsearch: () => go("search"),
    onsettings: () => go("settings"),
    onfound: () => go("found"),
    onmenu: () => (qa = true),
    ondesktop: onexit,
  });
</script>

<svelte:window {onkeydown} />

<Stage>
  {#snippet children({ height })}
    <div class="bp" style:--accent-game={accent}>
      {#if screen === "home"}
        {#if layout === "console"}
          <Console {height} {...layoutProps} />
        {:else if layout === "orbit"}
          <Orbit {height} {...layoutProps} />
        {:else}
          <Deck {height} {...layoutProps} />
        {/if}
      {:else if screen === "library"}
        <LibraryScreen {height} onplay={play} oninfo={info} {onfocus} onback={() => go("home")} />
      {:else if screen === "found"}
        <LibraryScreen {height} games={found} heading="Found on this PC" onplay={play} oninfo={info} {onfocus} onback={() => go("home")} />
      {:else if screen === "search"}
        <SearchScreen onplay={play} oninfo={info} {onfocus} onback={() => go("home")} />
      {:else}
        <BPSettings onback={() => go("home")} />
      {/if}

      {#if sheet}
        <GameSheet game={sheet} onplay={() => play(sheet)} onclose={() => (sheetId = null)} />
      {/if}
      {#if qa}
        <QuickAccess light={lightHex} onclose={() => (qa = false)} onsettings={() => go("settings")} ondesktop={onexit} />
      {/if}
      {#if launching}
        <Launching game={launching} onclose={() => (launchId = null)} />
      {/if}
    </div>
  {/snippet}
</Stage>

<style>
  .bp {
    position: absolute;
    inset: 0;
    background: #0a0e13;
    color: #e8edf2;
    font-family: var(--font);
    overflow: hidden;
    user-select: none;
  }
  .bp :global(button) {
    font-family: inherit;
    cursor: pointer;
  }
  .bp :global(button:focus-visible) {
    outline: none;
  }
</style>
