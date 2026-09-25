<script lang="ts">
  // A controller button, drawn for the controller in use (PlayStation
  // shapes or Xbox letters).
  import { glyphSet } from "../lib/input.svelte";

  type Button = "confirm" | "back" | "action" | "info" | "menu" | "view" | "lb" | "rb" | "lt" | "rt" | "home";
  let { button, size = 34 }: { button: Button; size?: number } = $props();

  const ps = $derived(glyphSet() === "playstation");
  const letter: Record<string, string> = { confirm: "A", back: "B", action: "X", info: "Y" };
  const pill: Record<string, [string, string]> = {
    lb: ["L1", "LB"],
    rb: ["R1", "RB"],
    lt: ["L2", "LT"],
    rt: ["R2", "RT"],
    home: ["PS", "⊕"],
  };
</script>

{#if pill[button]}
  <span class="pill" style:height="{size}px" style:font-size="{size * 0.42}px" aria-hidden="true">{pill[button][ps ? 0 : 1]}</span>
{:else}
  <span class="glyph" style:width="{size}px" style:height="{size}px" aria-hidden="true">
    {#if button === "menu" || button === "view"}
      <svg width={size * 0.47} height={size * 0.47} viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        {#if button === "menu"}<path d="M3 4.5h10M3 8h10M3 11.5h10" />{:else}<rect x="3" y="3.5" width="10" height="9" rx="1.5" />{/if}
      </svg>
    {:else if !ps}
      <span class="letter" style:font-size="{size * 0.48}px">{letter[button]}</span>
    {:else}
      <svg width={size * 0.47} height={size * 0.47} viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        {#if button === "confirm"}<path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
        {:else if button === "back"}<circle cx="8" cy="8" r="5.4" />
        {:else if button === "action"}<rect x="3" y="3" width="10" height="10" rx="1.2" />
        {:else}<path d="M8 2.8l5.8 10H2.2z" />{/if}
      </svg>
    {/if}
  </span>
{/if}

<style>
  .glyph {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(243, 245, 247, 0.14);
    border: 1px solid rgba(243, 245, 247, 0.32);
    color: #f3f5f7;
  }
  .letter {
    font-weight: 800;
    line-height: 1;
  }
  .pill {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    padding: 0 0.6em;
    border-radius: 9px;
    background: rgba(243, 245, 247, 0.14);
    border: 1px solid rgba(243, 245, 247, 0.32);
    color: #f3f5f7;
    font-weight: 800;
    letter-spacing: 0.02em;
  }
</style>
