<script lang="ts">
  // Big picture is designed on a 1920-wide canvas and scaled to the screen,
  // so it looks the same on a TV, a monitor or a laptop. Taller screens get
  // extra height rather than black bars.
  import type { Snippet } from "svelte";

  let { children }: { children: Snippet<[{ height: number }]> } = $props();

  const W = 1920;
  let vw = $state(window.innerWidth);
  let vh = $state(window.innerHeight);
  const scale = $derived(vw / W);
  const height = $derived(Math.max(1080, Math.round(vh / scale)));
  const fits = $derived(height * scale <= vh + 1);
  const s = $derived(fits ? scale : vh / 1080);
</script>

<svelte:window bind:innerWidth={vw} bind:innerHeight={vh} />

<div class="viewport">
  <div class="stage" style:width="{W}px" style:height="{fits ? height : 1080}px" style:transform="translate(-50%, -50%) scale({s})">
    {@render children({ height: fits ? height : 1080 })}
  </div>
</div>

<style>
  .viewport {
    position: fixed;
    inset: 0;
    overflow: hidden;
    background: #05070a;
  }
  .stage {
    position: absolute;
    left: 50%;
    top: 50%;
    transform-origin: 50% 50%;
    overflow: hidden;
  }
</style>
