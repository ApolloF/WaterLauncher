<script lang="ts">
  import Glyph from "./Glyph.svelte";

  type B = "confirm" | "back" | "action" | "info" | "menu" | "view" | "lb" | "rb" | "lt" | "rt" | "home";
  type H = { button: B; also?: B; label: string };
  let { hints, left = "" }: { hints: H[]; left?: string } = $props();
</script>

<div class="hints">
  {#if left}<span class="left">{left}</span>{/if}
  {#each hints as h (h.button + h.label)}
    <span class="h"><span class="gs"><Glyph button={h.button} />{#if h.also}<Glyph button={h.also} />{/if}</span>{h.label}</span>
  {/each}
</div>

<style>
  .hints {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 34px;
    font-size: 20px;
    font-weight: 600;
    color: #e8edf2;
  }
  .left {
    margin-right: auto;
    color: #9ba8b5;
    font-weight: 500;
  }
  .h {
    display: inline-flex;
    align-items: center;
    gap: 12px;
  }
  .gs {
    display: inline-flex;
    gap: 6px;
  }
</style>
