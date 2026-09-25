<script lang="ts">
  // A setting row: title, explanation and a switch. The whole row is the button.
  let {
    checked,
    title,
    detail = "",
    onchange,
    disabled = false,
  }: { checked: boolean; title: string; detail?: string; onchange: (v: boolean) => void; disabled?: boolean } = $props();
</script>

<button type="button" class="row" role="switch" aria-checked={checked} {disabled} onclick={() => onchange(!checked)}>
  <span class="text">
    <span class="title">{title}</span>
    {#if detail}<span class="detail">{detail}</span>{/if}
  </span>
  <span class="track" class:on={checked}><span class="knob"></span></span>
</button>

<style>
  .row {
    display: flex;
    align-items: center;
    gap: 16px;
    width: 100%;
    padding: 12px 16px;
    border: 0;
    border-radius: var(--radius);
    background: var(--surface-2);
    text-align: left;
  }
  .row:hover:not(:disabled) {
    background: var(--surface-3);
  }
  .text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .title {
    font-weight: 700;
    font-size: 15px;
  }
  .detail {
    font-size: 13.5px;
    color: var(--muted);
    line-height: 1.35;
  }
  .track {
    position: relative;
    flex-shrink: 0;
    width: 44px;
    height: 26px;
    border-radius: 13px;
    background: var(--line-strong);
    transition: background 0.18s var(--ease);
  }
  .track.on {
    background: var(--accent);
  }
  .knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
    transition: transform 0.18s var(--ease);
  }
  .track.on .knob {
    transform: translateX(18px);
  }
</style>
