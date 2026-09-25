<script lang="ts">
  import BigPicture from "./bigpicture/BigPicture.svelte";
  import Toasts from "./components/Toasts.svelte";
  import Desktop from "./desktop/Desktop.svelte";
  import { api } from "./lib/api";
  import { dispatch, pad, type Intent } from "./lib/input.svelte";
  import { errText, lib } from "./lib/store.svelte";

  let failed = $state("");
  let mode = $state<"desktop" | "bigpicture">("desktop");

  function enterBigPicture() {
    if (mode === "bigpicture") return;
    mode = "bigpicture";
    api.window.fullscreen(true);
  }
  function exitBigPicture() {
    if (mode === "desktop") return;
    mode = "desktop";
    api.window.fullscreen(false);
  }

  lib
    .init()
    .then(async () => {
      Object.assign(pad, await api.pad.state());
      if (lib.settings?.startInBigPicture) enterBigPicture();
    })
    .catch((e) => (failed = errText(e)));

  // Controller: in big picture every action goes to the screens; in desktop
  // mode the PS / Xbox button switches to big picture.
  $effect(() =>
    api.pad.onAction((action, repeat) => {
      if (mode === "bigpicture") dispatch(action as Intent, repeat);
      else if (action === "home" && !repeat) enterBigPicture();
    }),
  );
  $effect(() =>
    api.pad.onState((s) => {
      const wasConnected = pad.connected;
      Object.assign(pad, s);
      if (s.connected && !wasConnected && lib.settings?.openBigPictureOnController) enterBigPicture();
    }),
  );

  // Desktop mode follows the Windows theme unless the user picked one.
  let systemDark = $state(window.matchMedia("(prefers-color-scheme: dark)").matches);
  $effect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const on = (e: MediaQueryListEvent) => (systemDark = e.matches);
    mq.addEventListener("change", on);
    return () => mq.removeEventListener("change", on);
  });
  $effect(() => {
    if (mode === "bigpicture") return;
    const pref = lib.settings?.theme ?? "system";
    const dark = pref === "dark" || (pref === "system" && systemDark);
    document.documentElement.dataset.theme = dark ? "dark" : "light";
  });

  function onkeydown(e: KeyboardEvent) {
    if (e.key === "F11") {
      e.preventDefault();
      if (mode === "bigpicture") exitBigPicture();
      else enterBigPicture();
    }
  }

  // No right-click browser menu, except in text fields.
  function oncontextmenu(e: MouseEvent) {
    const t = e.target as HTMLElement;
    if (!t.closest("input, textarea")) e.preventDefault();
  }
</script>

<svelte:window {oncontextmenu} {onkeydown} />

{#if failed}
  <div class="fatal">
    <h1>WaterLauncher couldn't start</h1>
    <p>{failed}</p>
  </div>
{:else if mode === "bigpicture"}
  <BigPicture onexit={exitBigPicture} />
{:else}
  <Desktop onbigpicture={enterBigPicture} />
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
