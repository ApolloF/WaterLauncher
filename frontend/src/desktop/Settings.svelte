<script lang="ts">
  import Icon from "../components/Icon.svelte";
  import Toggle from "../components/Toggle.svelte";
  import { api } from "../lib/api";
  import { lib } from "../lib/store.svelte";
  import type { Settings } from "../lib/types";

  let { onclose }: { onclose: () => void } = $props();

  type Tab = "general" | "library" | "about";
  let tab = $state<Tab>("library");
  const tabs: { id: Tab; label: string }[] = [
    { id: "general", label: "General" },
    { id: "library", label: "Library" },
    { id: "about", label: "About" },
  ];

  let autoFolders = $state<string[]>([]);
  $effect(() => {
    api.autoFolders().then((f) => (autoFolders = f));
  });

  const s = $derived(lib.settings);
  const set = (patch: Partial<Settings>) => s && lib.saveSettings({ ...s, ...patch });

  async function addFolder() {
    const next = await lib.run(() => api.addFolder());
    if (next) lib.settings = next;
  }
  async function removeFolder(p: string) {
    const next = await lib.run(() => api.removeFolder(p));
    if (next) lib.settings = next;
  }

  let dialog: HTMLDivElement | undefined = $state();
  $effect(() => {
    dialog?.focus();
  });
</script>

<div class="scrim" role="presentation" onclick={(e) => e.target === e.currentTarget && onclose()}>
  <div class="dialog" role="dialog" aria-modal="true" aria-label="Settings" tabindex="-1" bind:this={dialog} onkeydown={(e) => e.key === "Escape" && onclose()}>
    <nav class="tabs" aria-label="Settings sections">
      <span class="heading">Settings</span>
      {#each tabs as t (t.id)}
        <button type="button" class:on={tab === t.id} aria-current={tab === t.id ? "page" : undefined} onclick={() => (tab = t.id)}>{t.label}</button>
      {/each}
    </nav>

    <section class="content">
      <div class="top">
        <h2>{tabs.find((t) => t.id === tab)?.label}</h2>
        <button type="button" class="close" aria-label="Close settings" onclick={onclose}><Icon name="close" size={18} stroke={2.2} /></button>
      </div>

      {#if s}
        {#if tab === "general"}
          <div class="group">
            <span class="glabel">Appearance</span>
            <div class="seg" role="group" aria-label="Theme">
              {#each [["system", "Follow Windows"], ["dark", "Dark"], ["light", "Light"]] as [id, label] (id)}
                <button type="button" class:on={s.theme === id} aria-pressed={s.theme === id} onclick={() => set({ theme: id as Settings["theme"] })}>{label}</button>
              {/each}
            </div>
            <p class="hint">Applies to desktop mode. Big picture mode is always dark.</p>
          </div>
        {:else if tab === "library"}
          <div class="group">
            <Toggle checked={s.autoFolders} title="Look in common game folders" detail="Games and repack folders on every drive." onchange={(v) => set({ autoFolders: v })} />
            {#if s.autoFolders && autoFolders.length}
              <ul class="paths auto">
                {#each autoFolders as f (f)}<li><Icon name="folder" size={16} /><span>{f}</span></li>{/each}
              </ul>
            {/if}
            <Toggle checked={s.detectUnofficial} title="Recognise unofficial copies" detail="Steam emulators, cracks and repacks, matched to the right game by their Steam AppID." onchange={(v) => set({ detectUnofficial: v })} />
            <Toggle checked={s.reviewUncertain} title="Let me check uncertain matches" detail="Games matched only by folder name wait in Found on this PC." onchange={(v) => set({ reviewUncertain: v })} />
            <Toggle checked={s.showNotInstalled} title="Show games that aren't installed" detail="Games you uninstalled stay listed with their playtime. Owned store games come in a later version." onchange={(v) => set({ showNotInstalled: v })} />
          </div>
          <div class="group">
            <span class="glabel">Your game folders</span>
            <p class="hint">Every folder inside these counts as a game.</p>
            {#if s.folders.length}
              <ul class="paths">
                {#each s.folders as f (f)}
                  <li>
                    <Icon name="folder" size={16} /><span>{f}</span>
                    <button type="button" class="icon" aria-label={`Stop looking in ${f}`} onclick={() => removeFolder(f)}><Icon name="close" size={14} stroke={2.2} /></button>
                  </li>
                {/each}
              </ul>
            {/if}
            <button type="button" class="btn" onclick={addFolder}><Icon name="plus" size={16} stroke={2.2} />Add folder</button>
          </div>
        {:else}
          <dl class="kv">
            <dt>Version</dt>
            <dd>{lib.info?.version ?? ""}</dd>
            <dt>Game database</dt>
            <dd>{lib.scan.known ? `${lib.scan.known.toLocaleString()} titles (Ludusavi manifest)` : "Downloading…"}</dd>
            <dt>Last scan</dt>
            <dd>{lib.scan.games} games in {(lib.scan.tookMs / 1000).toFixed(1)} s</dd>
            <dt>Data folder</dt>
            <dd class="path">{lib.info?.dataDir ?? ""}</dd>
          </dl>
          <div><button type="button" class="btn" onclick={() => lib.run(() => api.openLog())}><Icon name="file" size={16} />Open log folder</button></div>
        {/if}
      {/if}
    </section>
  </div>
</div>

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: 50;
    background: var(--scrim);
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fade 0.18s ease both;
  }
  @keyframes fade {
    from {
      opacity: 0;
    }
  }
  .dialog {
    width: min(960px, calc(100vw - 64px));
    height: min(660px, calc(100vh - 96px));
    display: flex;
    border-radius: var(--radius-l);
    background: var(--surface);
    border: 1px solid var(--line-strong);
    box-shadow: var(--shadow);
    overflow: hidden;
    outline: none;
    animation: pop 0.22s var(--ease) both;
  }
  @keyframes pop {
    from {
      transform: scale(0.98);
      opacity: 0;
    }
  }
  .tabs {
    width: 210px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 20px 12px;
    background: var(--bg);
    border-right: 1px solid var(--line);
  }
  .heading {
    padding: 0 12px 12px;
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 700;
  }
  .tabs button {
    height: 40px;
    padding: 0 12px;
    border: 0;
    border-radius: 10px;
    background: transparent;
    color: var(--text-2);
    font-size: 15px;
    font-weight: 600;
    text-align: left;
  }
  .tabs button:hover {
    background: var(--surface-2);
  }
  .tabs button.on {
    background: var(--accent-soft);
    color: var(--accent-text);
  }
  .content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 22px 28px 28px;
    overflow-y: auto;
  }
  .top {
    display: flex;
    align-items: center;
  }
  h2 {
    flex: 1;
    margin: 0;
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 700;
  }
  .close {
    width: 36px;
    height: 36px;
    border: 0;
    border-radius: 10px;
    background: var(--surface-2);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .glabel {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-2);
  }
  .hint {
    margin: -4px 0 2px;
    font-size: 13.5px;
    color: var(--muted);
  }
  .seg {
    display: flex;
    gap: 2px;
    padding: 3px;
    width: fit-content;
    border-radius: 10px;
    background: var(--surface-2);
  }
  .seg button {
    height: 34px;
    padding: 0 14px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    font-size: 14px;
    font-weight: 700;
  }
  .seg button.on {
    background: var(--surface-3);
    color: var(--text);
  }
  .paths {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .paths li {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 38px;
    padding: 0 6px 0 12px;
    border-radius: 10px;
    background: var(--surface-2);
    color: var(--text-2);
    font-size: 14px;
  }
  .paths.auto li {
    background: transparent;
    min-height: 28px;
    color: var(--muted);
  }
  .paths span {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .icon {
    width: 30px;
    height: 30px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .icon:hover {
    background: var(--surface-3);
    color: var(--text);
  }
  .btn {
    width: fit-content;
    height: 38px;
    padding: 0 14px;
    border-radius: 10px;
    border: 1px solid var(--line-strong);
    background: transparent;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14.5px;
    font-weight: 700;
  }
  .btn:hover {
    background: var(--surface-2);
  }
  .kv {
    display: grid;
    grid-template-columns: 140px minmax(0, 1fr);
    gap: 10px 16px;
    margin: 0;
    font-size: 14.5px;
  }
  dt {
    color: var(--muted);
  }
  dd {
    margin: 0;
  }
  .path {
    word-break: break-all;
    user-select: text;
    -webkit-user-select: text;
  }
</style>
