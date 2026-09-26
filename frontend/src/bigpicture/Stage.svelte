<script lang="ts">
  // Big picture is designed for a 1920×1080 screen and zoomed to fit, so it
  // looks the same on a TV, a monitor or a laptop. A taller screen (16:10)
  // gets extra height and a wider one (21:9) extra width, instead of bars.
  //
  // CSS zoom rather than a transform: the page is laid out and drawn at
  // the screen's own resolution, so text and art stay sharp at 4K and at
  // 720p instead of being a 1080p picture stretched or squeezed.
  import type { Snippet } from "svelte";

  let { children, clear = false }: { children: Snippet<[{ width: number; height: number }]>; clear?: boolean } = $props();

  const W = 1920;
  const H = 1080;
  let vw = $state(window.innerWidth);
  let vh = $state(window.innerHeight);
  const s = $derived(Math.max(0.1, Math.min(vw / W, vh / H)));
  // At least the design size; the rest of the screen in design pixels.
  const width = $derived(Math.max(W, Math.ceil(vw / s)));
  const height = $derived(Math.max(H, Math.ceil(vh / s)));
</script>

<svelte:window bind:innerWidth={vw} bind:innerHeight={vh} />

<div class="viewport" class:clear>
  <div class="stage" style:width="{width}px" style:height="{height}px" style:zoom={s}>
    {@render children({ width, height })}
  </div>
</div>

<style>
  .viewport {
    position: fixed;
    inset: 0;
    overflow: hidden;
    background: #05070a;
  }
  .viewport.clear {
    background: transparent;
  }
  .stage {
    position: absolute;
    left: 0;
    top: 0;
    overflow: hidden;
  }
</style>
