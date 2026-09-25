<script lang="ts">
  // The moment a game starts. From v0.4 this shows the hooks (saves,
  // add-ons, controller) running; for now it starts the game and confirms.
  import GameArt from "../components/GameArt.svelte";
  import { api } from "../lib/api";
  import { feedback, useInput } from "../lib/input.svelte";
  import { errText } from "../lib/store.svelte";
  import { title, type Game } from "../lib/types";
  import Hints from "./Hints.svelte";

  let { game, onclose }: { game: Game; onclose: () => void } = $props();

  let phase = $state<"starting" | "started" | "failed">("starting");
  let error = $state("");
  let logoFailed = $state(false);

  $effect(() => {
    let closeTimer: ReturnType<typeof setTimeout> | undefined;
    api
      .play(game.id)
      .then(() => {
        phase = "started";
        feedback.confirm();
        closeTimer = setTimeout(onclose, 3500);
      })
      .catch((e) => {
        phase = "failed";
        error = errText(e);
        feedback.error();
      });
    return () => clearTimeout(closeTimer);
  });

  $effect(() =>
    useInput((i) => {
      if (i === "back" || (i === "confirm" && phase !== "starting")) onclose();
    }),
  );
</script>

<div class="launch">
  <div class="art"><GameArt {game} kind="hero" /></div>
  <div class="shade"></div>
  <div class="content">
    <span class="label">{phase === "failed" ? "Couldn't start" : phase === "started" ? "Started" : "Starting"}</span>
    {#if game.meta?.logo && !logoFailed}
      <img class="logo" src={game.meta.logo} alt={title(game)} onerror={() => (logoFailed = true)} />
    {:else}
      <h1>{title(game)}</h1>
    {/if}
    {#if phase === "starting"}
      <div class="bar"><span></span></div>
    {:else if phase === "started"}
      <p>Have fun. WaterLauncher stays out of the way until you come back.</p>
    {:else}
      <p class="err">{error}</p>
    {/if}
  </div>
  <div class="hints"><Hints hints={phase === "starting" ? [{ button: "back", label: "Cancel" }] : [{ button: "confirm", label: "Back to library" }]} /></div>
</div>

<style>
  .launch {
    position: absolute;
    inset: 0;
    z-index: 40;
    background: #05080b;
    animation: in 0.35s ease both;
  }
  @keyframes in {
    from {
      opacity: 0;
    }
  }
  .art {
    position: absolute;
    inset: 0;
    opacity: 0.55;
    animation: kb 12s ease-out both;
  }
  @keyframes kb {
    from {
      transform: scale(1.06);
    }
  }
  .shade {
    position: absolute;
    inset: 0;
    background: linear-gradient(0deg, rgba(5, 8, 11, 0.97) 0%, rgba(5, 8, 11, 0.55) 55%, rgba(5, 8, 11, 0.25) 100%);
  }
  .content {
    position: absolute;
    left: 140px;
    bottom: 170px;
    width: 1000px;
    display: flex;
    flex-direction: column;
    gap: 26px;
    color: #f3f5f7;
  }
  .label {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: oklch(0.8 0.12 205);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 110px;
    line-height: 0.95;
  }
  .logo {
    max-width: 760px;
    max-height: 260px;
    object-fit: contain;
    object-position: left bottom;
    filter: drop-shadow(0 6px 30px rgba(0, 0, 0, 0.6));
  }
  .bar {
    width: 520px;
    height: 8px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.12);
    overflow: hidden;
  }
  .bar span {
    display: block;
    height: 100%;
    width: 40%;
    border-radius: 4px;
    background: oklch(0.8 0.12 205);
    animation: slide 1.1s ease-in-out infinite;
  }
  @keyframes slide {
    from {
      transform: translateX(-110%);
    }
    to {
      transform: translateX(260%);
    }
  }
  p {
    margin: 0;
    font-size: 24px;
    line-height: 1.5;
    color: #c9d3dc;
  }
  .err {
    color: #ffb3b3;
  }
  .hints {
    position: absolute;
    right: 96px;
    bottom: 48px;
  }
</style>
