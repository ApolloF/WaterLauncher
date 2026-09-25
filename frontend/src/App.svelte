<script lang="ts">
  import Toasts from "./components/Toasts.svelte";
  import Desktop from "./desktop/Desktop.svelte";
  import { errText, lib } from "./lib/store.svelte";

  let failed = $state("");
  lib.init().catch((e) => (failed = errText(e)));

  // Desktop mode follows the Windows theme unless the user picked one.
  let systemDark = $state(window.matchMedia("(prefers-color-scheme: dark)").matches);
  $effect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const on = (e: MediaQueryListEvent) => (systemDark = e.matches);
    mq.addEventListener("change", on);
    return () => mq.removeEventListener("change", on);
  });
  $effect(() => {
    const pref = lib.settings?.theme ?? "system";
    const dark = pref === "dark" || (pref === "system" && systemDark);
    document.documentElement.dataset.theme = dark ? "dark" : "light";
  });

  // No right-click browser menu, except in text fields.
  function oncontextmenu(e: MouseEvent) {
    const t = e.target as HTMLElement;
    if (!t.closest("input, textarea")) e.preventDefault();
  }
</script>

<svelte:window {oncontextmenu} />

{#if failed}
  <div class="fatal">
    <h1>WaterLauncher couldn't start</h1>
    <p>{failed}</p>
  </div>
{:else}
  <Desktop />
{/if}
<Toasts />

<style>
  .fatal {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 40px;
    text-align: center;
  }
  .fatal h1 {
    margin: 0;
    font-family: var(--font-display);
  }
  .fatal p {
    margin: 0;
    color: var(--muted);
    user-select: text;
  }
</style>
