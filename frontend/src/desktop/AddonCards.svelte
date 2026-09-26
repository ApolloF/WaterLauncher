<script lang="ts">
  // What the enabled add-ons say about a game: badges, facts and actions.
  import { api } from "../lib/api";
  import { lib } from "../lib/store.svelte";
  import type { AddonAction, AddonGame, Game } from "../lib/types";

  let { game }: { game: Game } = $props();

  let items = $state<AddonGame[]>([]);
  let busy = $state<string | null>(null); // addon/action running
  let progress = $state("");
  let confirming = $state<{ addon: string; action: AddonAction } | null>(null);

  function load(id: number) {
    let live = true;
    api.addons
      .forGame(id)
      .then((r) => live && (items = r))
      .catch(() => {});
    return () => (live = false);
  }
  $effect(() => {
    const id = game.id;
    items = [];
    confirming = null;
    const t = setTimeout(() => load(id), 300);
    return () => clearTimeout(t);
  });
  $effect(() =>
    api.addons.onProgress((a, text) => {
      if (busy?.startsWith(a + "/")) progress = text;
    }),
  );

  async function run(addon: string, action: AddonAction) {
    if (action.confirm && confirming?.action.id !== action.id) {
      confirming = { addon, action };
      return;
    }
    confirming = null;
    busy = `${addon}/${action.id}`;
    progress = "";
    const msg = await lib.run(() => api.addons.runAction(game.id, addon, action.id));
    busy = null;
    progress = "";
    if (msg) lib.toast(msg);
    load(game.id);
  }
</script>

{#each items as a (a.addon)}
  <div class="card addon">
    <div class="card-head">
      <span class="grow">{a.name}</span>
      {#each a.badges as b (b.text)}
        <span class="badge {b.tone ?? 'info'}" title={b.tooltip}>{b.text}</span>
      {/each}
    </div>
    {#if a.error}
      <p class="err">{a.error}</p>
    {/if}
    {#if a.lines.length}
      <dl>
        {#each a.lines as l (l.label)}
          <dt>{l.label}</dt>
          <dd>{l.value}</dd>
        {/each}
      </dl>
    {/if}
    {#if confirming?.addon === a.addon}
      <div class="confirm">
        <p>{confirming.action.confirm}</p>
        <div class="row">
          <button type="button" class="btn primary" onclick={() => run(a.addon, confirming!.action)}>{confirming.action.label}</button>
          <button type="button" class="btn" onclick={() => (confirming = null)}>Cancel</button>
        </div>
      </div>
    {:else if a.actions.length}
      <div class="row">
        {#each a.actions as act (act.id)}
          <button type="button" class="btn" disabled={!!busy} title={act.description} onclick={() => run(a.addon, act)}>
            {busy === `${a.addon}/${act.id}` ? "Working…" : act.label}
          </button>
        {/each}
      </div>
    {/if}
    {#if busy?.startsWith(a.addon + "/") && progress}
      <p class="progress">{progress}</p>
    {/if}
  </div>
{/each}

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    border-radius: var(--radius);
    background: var(--surface-2);
  }
  .card-head {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-size: 15px;
    font-weight: 700;
    color: var(--text);
  }
  .grow {
    flex: 1;
  }
  .badge {
    font-size: 12px;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 999px;
    background: var(--surface-3);
    color: var(--text-2);
  }
  .badge.ok {
    background: var(--accent-soft);
    color: var(--accent-text);
  }
  .badge.warn {
    background: color-mix(in oklab, var(--warn) 18%, transparent);
    color: var(--warn);
  }
  dl {
    margin: 0;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 4px 14px;
    font-size: 13.5px;
  }
  dt {
    color: var(--muted);
  }
  dd {
    margin: 0;
    color: var(--text-2);
  }
  p {
    margin: 0;
    font-size: 13.5px;
    line-height: 1.45;
    color: var(--text-2);
  }
  .err {
    color: var(--danger);
  }
  .progress {
    color: var(--muted);
  }
  .row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .btn {
    height: 32px;
    padding: 0 12px;
    border-radius: var(--radius-s);
    border: 1px solid var(--line-strong);
    background: transparent;
    color: var(--text);
    font-size: 13px;
    font-weight: 700;
  }
  .btn:hover:not(:disabled) {
    background: var(--surface-3);
  }
  .btn:disabled {
    opacity: 0.6;
  }
  .btn.primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-ink);
  }
  .confirm {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
</style>
