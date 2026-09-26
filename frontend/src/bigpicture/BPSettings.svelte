<script lang="ts">
  // Big picture settings, made for a controller: up/down to pick a row,
  // left/right or ✕ to change it.
  import { feedback, useInput } from "../lib/input.svelte";
  import { lib } from "../lib/store.svelte";
  import type { Settings } from "../lib/types";
  import Hints from "./Hints.svelte";
  import { clamp } from "./nav";

  let { onback }: { onback: () => void } = $props();

  type Row =
    | { key: keyof Settings; title: string; detail: string; kind: "toggle" }
    | { key: keyof Settings; title: string; detail: string; kind: "choice"; options: [string, string][] };

  const rows: Row[] = [
    {
      key: "bigPictureLayout",
      title: "Layout",
      detail: "How big picture looks. Deck is the default.",
      kind: "choice",
      options: [
        ["deck", "Deck"],
        ["console", "Console"],
        ["orbit", "Orbit"],
      ],
    },
    {
      key: "glyphs",
      title: "Button prompts",
      detail: "Shapes for PlayStation controllers, letters for Xbox. Auto follows the controller in use.",
      kind: "choice",
      options: [
        ["auto", "Auto"],
        ["playstation", "PlayStation"],
        ["xbox", "Xbox"],
      ],
    },
    { key: "haptics", title: "Haptics while browsing", detail: "Light ticks as you move between games.", kind: "toggle" },
    { key: "lightbar", title: "Lightbar follows the game", detail: "Tints the DualSense to the selected game.", kind: "toggle" },
    { key: "sounds", title: "Navigation sounds", detail: "Soft clicks as you move.", kind: "toggle" },
    { key: "openBigPictureOnController", title: "Open big picture when a controller connects", detail: "Switches over as soon as you pick one up.", kind: "toggle" },
    { key: "startInBigPicture", title: "Start in big picture", detail: "Skip the desktop window when WaterLauncher starts.", kind: "toggle" },
    { key: "psButton", title: "PS / Xbox button opens WaterLauncher", detail: "From other apps, and the overlay in games. Turn off Steam's guide-button shortcut to avoid both opening.", kind: "toggle" },
    { key: "showOwned", title: "Show games you own that aren't installed", detail: "From the store accounts connected in desktop Settings.", kind: "toggle" },
    { key: "syncSavesBefore", title: "Sync saves before playing", detail: "With Syncer: saves come up to date from your other PCs first.", kind: "toggle" },
    { key: "backupSavesAfter", title: "Back up saves after playing", detail: "With Syncer: a backup runs when the game exits.", kind: "toggle" },
    { key: "closeWhilePlaying", title: "Close the interface while playing", detail: "Frees its memory. It comes back when the game exits.", kind: "toggle" },
    { key: "autoUpdate", title: "Keep WaterLauncher up to date", detail: "New versions download in the background and install the next time WaterLauncher starts.", kind: "toggle" },
    {
      key: "padWhilePlaying",
      title: "Controller while playing",
      detail: "Listen: the PS button opens the overlay, nothing is sent to the controller. Off: let go completely.",
      kind: "choice",
      options: [
        ["listen", "Listen"],
        ["off", "Off"],
      ],
    },
  ];

  let i = $state(0);
  const s = $derived(lib.settings);

  function change(dir: 1 | -1) {
    if (!s) return;
    const row = rows[i];
    if (row.kind === "toggle") {
      lib.saveSettings({ ...s, [row.key]: !s[row.key] });
    } else {
      const ids = row.options.map((o) => o[0]);
      const cur = ids.indexOf(String(s[row.key]));
      const next = ids[(cur + dir + ids.length) % ids.length];
      lib.saveSettings({ ...s, [row.key]: next } as Settings);
    }
    feedback.confirm();
  }

  $effect(() =>
    useInput((intent) => {
      if (intent === "up" || intent === "down") {
        const j = clamp(i + (intent === "up" ? -1 : 1), 0, rows.length - 1);
        if (j !== i) {
          i = j;
          feedback.move();
        }
      } else if (intent === "confirm" || intent === "right") change(1);
      else if (intent === "left") change(-1);
      else if (intent === "back") onback();
      else return false;
    }),
  );

  function valueLabel(row: Row): string {
    if (!s) return "";
    if (row.kind === "toggle") return s[row.key] ? "On" : "Off";
    return row.options.find((o) => o[0] === s[row.key])?.[1] ?? "";
  }
</script>

<div class="settings">
  <h1>Settings</h1>
  <div class="rows">
    {#each rows as row, k (row.key)}
      <button type="button" class="row" class:on={k === i} onclick={() => ((i = k), change(1))}>
        <span class="text"><span class="t">{row.title}</span><span class="d">{row.detail}</span></span>
        {#if row.kind === "toggle"}
          <span class="track" class:yes={!!s?.[row.key]}><span class="knob"></span></span>
        {:else}
          <span class="choice"><span class="arrow">‹</span>{valueLabel(row)}<span class="arrow">›</span></span>
        {/if}
      </button>
    {/each}
  </div>
  <p class="more">Library folders and art sources are in desktop mode's settings.</p>
  <div class="hints"><Hints hints={[{ button: "confirm", label: "Change" }, { button: "back", label: "Back" }]} /></div>
</div>

<style>
  .settings {
    position: absolute;
    inset: 0;
    background: #0a0e13;
    color: #e8edf2;
    padding: 60px 110px;
  }
  h1 {
    margin: 0 0 30px;
    font-family: var(--font-display);
    font-size: 52px;
  }
  .rows {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 1300px;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 20px 26px;
    border: 0;
    border-radius: 18px;
    background: #121a23;
    color: #e8edf2;
    text-align: left;
  }
  .row.on {
    background: #1b2632;
    box-shadow: 0 0 0 3px #fff;
  }
  .text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .t {
    font-size: 24px;
    font-weight: 700;
  }
  .d {
    font-size: 18px;
    color: #9ba8b5;
  }
  .choice {
    display: flex;
    align-items: center;
    gap: 14px;
    font-size: 22px;
    font-weight: 700;
    min-width: 220px;
    justify-content: space-between;
  }
  .arrow {
    color: #9ba8b5;
    font-size: 30px;
  }
  .track {
    position: relative;
    width: 62px;
    height: 34px;
    border-radius: 17px;
    background: rgba(255, 255, 255, 0.18);
    flex-shrink: 0;
    transition: background 0.2s;
  }
  .track.yes {
    background: oklch(0.8 0.12 205);
  }
  .knob {
    position: absolute;
    top: 4px;
    left: 4px;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: #fff;
    transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  .track.yes .knob {
    transform: translateX(28px);
  }
  .more {
    margin-top: 26px;
    font-size: 19px;
    color: #9ba8b5;
  }
  .hints {
    position: absolute;
    right: 96px;
    bottom: 40px;
    left: 110px;
  }
</style>
