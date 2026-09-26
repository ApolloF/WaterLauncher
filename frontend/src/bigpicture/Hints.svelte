<script lang="ts">
  // The prompts along the bottom. Each one is also a button, for a mouse.
  import { dispatch } from "../lib/input.svelte";
  import Glyph from "./Glyph.svelte";

  type B = "confirm" | "back" | "action" | "info" | "menu" | "view" | "lb" | "rb" | "lt" | "rt" | "home";
  type H = { button: B; also?: B; label: string };
  let { hints, left = "" }: { hints: H[]; left?: string } = $props();
</script>

<div class="hints">
  {#if left}<span class="left">{left}</span>{/if}
  {#each hints as h (h.button + h.label)}
    <button type="button" class="h" tabindex="-1" onclick={() => dispatch(h.button)}>
      <span class="gs"><Glyph button={h.button} />{#if h.also}<Glyph button={h.also} />{/if}</span>{h.label}
    </button>
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
    min-width: 0;
  }
  .left {
    margin-right: auto;
    color: #9ba8b5;
    font-weight: 500;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .h {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    padding: 0;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    white-space: nowrap;
  }
  .gs {
    display: inline-flex;
    gap: 6px;
  }
</style>
