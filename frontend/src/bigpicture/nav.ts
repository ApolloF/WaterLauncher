// Small helpers for moving focus through rows and grids.

export const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v));

/** Next index in a list for a left/right (or up/down) move; null at the edge. */
export function step(i: number, n: number, dir: -1 | 1): number | null {
  const j = i + dir;
  return j < 0 || j >= n ? null : j;
}

/** Next index in a grid of `cols` columns; null when the move leaves the grid. */
export function gridStep(i: number, n: number, cols: number, intent: string): number | null {
  const row = Math.floor(i / cols);
  const col = i % cols;
  switch (intent) {
    case "left":
      return col === 0 ? null : i - 1;
    case "right":
      return i + 1 >= n || col === cols - 1 ? null : i + 1;
    case "up":
      return row === 0 ? null : i - cols;
    case "down": {
      if (row >= Math.ceil(n / cols) - 1) return null;
      return Math.min(n - 1, i + cols);
    }
  }
  return null;
}

/** Scroll offset that keeps item i of size `size` (with `gap`) visible, showing `lead` items before it. */
export function rowOffset(i: number, size: number, gap: number, lead = 1): number {
  return i <= lead ? 0 : -(i - lead) * (size + gap);
}
