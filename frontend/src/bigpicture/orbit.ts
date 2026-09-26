// Orbit layout geometry: games on a honeycomb around the most recently
// played one, seen through a soft lens centred on the selected game. The
// middle of the screen is flat; towards the edges bubbles move closer and
// shrink, always keeping a gap, and fade out.

export type Cell = { q: number; r: number };
export type Dir = "left" | "right" | "up" | "down";

const ROW = Math.sqrt(3) / 2; // height of a honeycomb row, in cell widths

/** Hex cells in rings around the centre (ring 0 has 1, ring r has 6r), enough for n games. */
export function spiral(n: number): Cell[] {
  const out: Cell[] = n > 0 ? [{ q: 0, r: 0 }] : [];
  // The six sides of a ring, in order, starting from its lower-left corner.
  const dirs = [
    [1, 0],
    [1, -1],
    [0, -1],
    [-1, 0],
    [-1, 1],
    [0, 1],
  ];
  for (let ring = 1; out.length < n; ring++) {
    let q = -ring,
      r = ring; // start at a corner, walk the six sides
    for (let side = 0; side < 6; side++) {
      for (let s = 0; s < ring; s++) {
        if (out.length >= n) break;
        out.push({ q, r });
        q += dirs[side][0];
        r += dirs[side][1];
      }
    }
  }
  return out;
}

/** A cell's place on the flat honeycomb, in cell widths. */
export const flat = (c: Cell) => ({ x: c.q + c.r / 2, y: c.r * ROW });

export const hexDist = (a: Cell, b: Cell) => (Math.abs(a.q - b.q) + Math.abs(a.r - b.r) + Math.abs(a.q + a.r - b.q - b.r)) / 2;

/**
 * The cell a move lands on: the nearest one ahead. Left and right keep to
 * the row. Up and down keep to the column the last sideways move was in
 * (anchorX), so a path runs straight instead of drifting to one side, and
 * holes or the edge of the honeycomb don't stop it while a game is ahead.
 * Ties go towards the middle. Null when nothing is that way.
 */
export function nextCell(cells: Cell[], from: number, dir: Dir, anchorX: number): number | null {
  if (!cells[from]) return null;
  const f = flat(cells[from]);
  const across = dir === "left" || dir === "right";
  const sign = dir === "left" || dir === "up" ? -1 : 1;
  let best: number | null = null;
  let bestScore = Infinity;
  let bestMid = Infinity;
  for (let k = 0; k < cells.length; k++) {
    if (k === from) continue;
    const c = flat(cells[k]);
    const dx = c.x - f.x;
    const dy = c.y - f.y;
    const ahead = sign * (across ? dx : dy);
    if (ahead < 0.45) continue;
    let score: number;
    if (across) {
      if (Math.abs(dy) > ahead * 1.8 + 0.01) continue; // at most one row up or down per step
      score = ahead + 3 * Math.abs(dy);
    } else {
      if (Math.abs(dx) > ahead * 1.5 + 0.6) continue;
      score = Math.round(ahead / ROW) * 100 + Math.abs(c.x - anchorX); // nearest row first
    }
    const mid = Math.abs(c.x) + Math.abs(c.y);
    if (score < bestScore - 1e-6 || (score < bestScore + 1e-6 && mid < bestMid)) {
      best = k;
      bestScore = score;
      bestMid = mid;
    }
  }
  return best;
}

export type Lens = {
  /** Half the width and height the honeycomb may fill, in pixels. */
  ax: number;
  ay: number;
  /** Share of that which stays flat (0–1). */
  flat: number;
};

/** A bubble's place relative to the selected one, in pixels, with its scale (1 = normal size) and opacity. */
export type Place = { x: number; y: number; scale: number; opacity: number };

const NEIGHBOURS = [
  [1, 0],
  [-1, 0],
  [0.5, ROW],
  [-0.5, ROW],
  [0.5, -ROW],
  [-0.5, -ROW],
];

/**
 * Where the bubble at (dx, dy) cell widths from the selected one shows up.
 * `size` is a bubble's diameter and `spacing` the distance between cell
 * centres, both in pixels; the selected bubble is `focus` times as big and
 * its neighbours move out to make room. A bubble's scale is set by how close
 * its neighbouring cells end up, so bubbles never overlap and the gaps stay
 * in proportion however hard the lens squeezes.
 */
export function place(dx: number, dy: number, size: number, spacing: number, focus: number, lens: Lens): Place {
  if (dx === 0 && dy === 0) return { x: 0, y: 0, scale: focus, opacity: 1 };
  const push = (size * (focus - 1)) / 2 + size * 0.04;
  const at = (x: number, y: number): [number, number, number] => {
    let px = x * spacing;
    let py = y * spacing;
    const d = Math.hypot(px, py);
    if (d < 1e-6) return [0, 0, 0];
    px *= (d + push) / d;
    py *= (d + push) / d;
    const r = Math.hypot(px / lens.ax, py / lens.ay);
    if (r <= lens.flat) return [px, py, r];
    const t = lens.flat + (1 - lens.flat) * Math.tanh((r - lens.flat) / (1 - lens.flat));
    return [(px * t) / r, (py * t) / r, t];
  };
  const [x, y, t] = at(dx, dy);
  let room = Infinity;
  for (const [nx, ny] of NEIGHBOURS) {
    const [x2, y2] = at(dx + nx, dy + ny);
    room = Math.min(room, Math.hypot(x2 - x, y2 - y));
  }
  const scale = Math.min(1, room / spacing);
  // Fade out towards the edge, and bubbles too small to make out.
  const edge = Math.max(0, Math.min(1, (0.97 - t) / 0.12));
  const opacity = scale * size < 14 ? 0 : edge;
  return { x, y, scale, opacity };
}
