<script lang="ts">
  // A game's picture: downloaded art when there is some, otherwise the
  // generated landscape. `kind` picks which picture (cover is portrait,
  // hero is wide).
  import { artFor } from "../lib/art";
  import type { Game } from "../lib/types";

  let { game, kind = "cover" }: { game: Game; kind?: "cover" | "hero" } = $props();

  const src = $derived(kind === "cover" ? game.meta?.cover : game.meta?.hero);
  const a = $derived(artFor(game.key));
  let failed = $state(false);
  $effect(() => {
    src;
    failed = false;
  });
</script>

{#if src && !failed}
  <img class="img" {src} alt="" loading="lazy" decoding="async" draggable="false" onerror={() => (failed = true)} />
{:else}
  <div class="art" style:background={a.sky} aria-hidden="true">
    {#if a.starOp}<div class="layer" style:background-image={a.stars} style:opacity={a.starOp}></div>{/if}
    <div class="sun" style:left={a.sunL} style:top={a.sunT} style:width={a.sunW} style:background={a.sun} style:mask-image={a.sunMask}></div>
    <div class="layer" style:background={a.far} style:clip-path={a.farClip}></div>
    <div class="layer" style:background={a.fog}></div>
    <div class="layer" style:background={a.mid} style:clip-path={a.midClip}></div>
    {#if a.floorH !== "0%"}<div class="floor" style:height={a.floorH} style:background={a.floor}></div>{/if}
    <div class="layer" style:background={a.near} style:clip-path={a.nearClip}></div>
  </div>
{/if}

<style>
  .img,
  .art {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }
  .img {
    object-fit: cover;
  }
  .art {
    overflow: hidden;
  }
  .layer {
    position: absolute;
    inset: 0;
  }
  .sun {
    position: absolute;
    aspect-ratio: 1;
    border-radius: 50%;
  }
  .floor {
    position: absolute;
    left: -50%;
    right: -50%;
    bottom: 0;
  }
</style>
