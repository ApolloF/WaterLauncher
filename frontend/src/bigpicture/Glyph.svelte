<script lang="ts">
  // A button prompt, drawn for what is in use: PlayStation shapes, Xbox
  // letters, or keyboard keys.
  import { glyphSet, KEY_LABELS } from "../lib/input.svelte";

  type Button = "confirm" | "back" | "action" | "info" | "menu" | "view" | "lb" | "rb" | "lt" | "rt" | "home";
  let { button, size = 34 }: { button: Button; size?: number } = $props();

  const set = $derived(glyphSet());
  const letter: Record<string, string> = { confirm: "A", back: "B", action: "X", info: "Y" };
  const pill: Record<string, [string, string]> = {
    lb: ["L1", "LB"],
    rb: ["R1", "RB"],
    lt: ["L2", "LT"],
    rt: ["R2", "RT"],
  };
  const label = $derived(set === "keyboard" ? KEY_LABELS[button] : "");
</script>

{#if set === "keyboard"}
  <span class="key" style:height="{size}px" style:min-width="{size}px" style:font-size="{size * (label.length > 2 ? 0.4 : 0.46)}px" aria-hidden="true">{label}</span>
{:else if pill[button]}
  <span class="pill" style:height="{size}px" style:font-size="{size * 0.42}px" aria-hidden="true">{pill[button][set === "playstation" ? 0 : 1]}</span>
{:else if button === "menu" || button === "view"}
  <!-- Options / Create on a DualSense, Menu / View on an Xbox controller:
       the small buttons either side of the touchpad or the guide button. -->
  <span class="small" style:height="{size}px" style:width="{size * 1.3}px" aria-hidden="true">
    <svg width={size * 0.52} height={size * 0.52} viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
      {#if button === "menu"}
        <path d="M3.5 4.5h9M3.5 8h9M3.5 11.5h9" />
      {:else if set === "playstation"}
        <path d="M8 13V7.5M8 13 4.2 8.6M8 13l3.8-4.4M8 4.4V3M3.3 5.4l-.9-1M12.7 5.4l.9-1" />
      {:else}
        <rect x="2.5" y="4.5" width="7.5" height="7" rx="1.2" /><path d="M6 4.5V3.3c0-.5.4-.8.8-.8h5.9c.5 0 .8.4.8.8v5.9c0 .5-.4.8-.8.8H10" />
      {/if}
    </svg>
  </span>
{:else if button === "home"}
  <span class="glyph" style:width="{size}px" style:height="{size}px" aria-hidden="true"><span class="letter" style:font-size="{size * 0.36}px">{set === "playstation" ? "PS" : "⊕"}</span></span>
{:else}
  <span class="glyph" style:width="{size}px" style:height="{size}px" aria-hidden="true">
    {#if set === "xbox"}
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
  .glyph,
  .small,
  .pill,
  .key {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(243, 245, 247, 0.14);
    border: 1px solid rgba(243, 245, 247, 0.32);
    color: #f3f5f7;
  }
  .glyph {
    border-radius: 50%;
  }
  .small {
    border-radius: 999px;
  }
  .letter {
    font-weight: 800;
    line-height: 1;
  }
  .pill {
    padding: 0 0.6em;
    border-radius: 9px;
    font-weight: 800;
    letter-spacing: 0.02em;
  }
  .key {
    box-sizing: border-box;
    padding: 0 0.55em;
    border-radius: 8px;
    border-bottom-width: 3px;
    font-weight: 700;
    line-height: 1;
    white-space: nowrap;
  }
</style>
