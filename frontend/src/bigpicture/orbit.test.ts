import { describe, expect, it } from "vitest";
import { flat, hexDist, nextCell, place, spiral, type Dir } from "./orbit";

const at = (cells: ReturnType<typeof spiral>, q: number, r: number) => cells.findIndex((c) => c.q === q && c.r === r);

describe("spiral", () => {
  it("fills rings around the centre", () => {
    const cells = spiral(19);
    expect(cells[0]).toEqual({ q: 0, r: 0 });
    expect(cells.slice(1, 7).every((c) => hexDist(c, cells[0]) === 1)).toBe(true);
    expect(cells.slice(7).every((c) => hexDist(c, cells[0]) === 2)).toBe(true);
    expect(new Set(cells.map((c) => `${c.q},${c.r}`)).size).toBe(19);
  });
});

describe("nextCell", () => {
  const cells = spiral(61); // four full rings

  it("moves to the neighbour in the row", () => {
    expect(nextCell(cells, 0, "right", 0)).toBe(at(cells, 1, 0));
    expect(nextCell(cells, 0, "left", 0)).toBe(at(cells, -1, 0));
  });

  it("goes straight up and down instead of drifting sideways", () => {
    for (const dir of ["up", "down"] as Dir[]) {
      let k = 0;
      const xs: number[] = [];
      for (let s = 0; s < 4; s++) {
        const n = nextCell(cells, k, dir, 0);
        expect(n).not.toBeNull();
        k = n!;
        xs.push(flat(cells[k]).x);
      }
      // A honeycomb column zigzags half a cell either way, never further.
      expect(Math.max(...xs.map(Math.abs))).toBeLessThanOrEqual(0.5);
      expect(Math.abs(flat(cells[k]).y)).toBeGreaterThan(3); // four rows away
    }
  });

  it("comes back the way it went", () => {
    const up = nextCell(cells, 0, "up", 0)!;
    expect(nextCell(cells, up, "down", 0)).toBe(0);
    const right = nextCell(cells, 0, "right", 0)!;
    expect(nextCell(cells, right, "left", flat(cells[right]).x)).toBe(0);
  });

  it("gets past holes and a ragged edge", () => {
    const some = spiral(10); // ring 2 only partly filled
    for (let k = 0; k < some.length; k++) {
      for (const dir of ["left", "right", "up", "down"] as Dir[]) {
        const n = nextCell(some, k, dir, flat(some[k]).x);
        if (n === null) continue;
        const a = flat(some[k]);
        const b = flat(some[n]);
        const ahead = dir === "left" ? a.x - b.x : dir === "right" ? b.x - a.x : dir === "up" ? a.y - b.y : b.y - a.y;
        expect(ahead).toBeGreaterThan(0.4);
      }
    }
    // Nothing further right on the edge of a small honeycomb, or without games.
    expect(nextCell(spiral(1), 0, "right", 0)).toBeNull();
    expect(nextCell(spiral(0), 0, "up", 0)).toBeNull();
  });
});

describe("place", () => {
  const lens = { ax: 880, ay: 390, flat: 0.5 };
  const size = 176;
  const spacing = 198;

  it("keeps the middle flat and makes room around the selected game", () => {
    const n = place(1, 0, size, spacing, 1.25, lens);
    expect(n.scale).toBeCloseTo(1, 5);
    expect(n.opacity).toBe(1);
    expect(n.x).toBeGreaterThan(spacing); // pushed out
    expect(n.x - (1.25 * size) / 2 - size / 2).toBeGreaterThan(spacing - size - 1); // gap at least as wide as elsewhere
  });

  it("never lets bubbles overlap, however hard the lens squeezes", () => {
    const cells = spiral(400);
    for (const focus of [0, 7, 40]) {
      const f = flat(cells[focus]);
      const placed = cells
        .map((c, k) => {
          const p = flat(c);
          return { k, ...place(p.x - f.x, p.y - f.y, size, spacing, 1.25, lens) };
        })
        .filter((b) => b.opacity > 0);
      for (let a = 0; a < placed.length; a++) {
        for (let b = a + 1; b < placed.length; b++) {
          const A = placed[a];
          const B = placed[b];
          const d = Math.hypot(A.x - B.x, A.y - B.y);
          expect(d).toBeGreaterThan(((A.scale + B.scale) * size) / 2 - 0.5);
        }
      }
    }
  });

  it("shrinks and fades bubbles towards the edge", () => {
    const near = place(2, 0, size, spacing, 1.25, lens);
    const far = place(5, 0, size, spacing, 1.25, lens);
    expect(far.scale).toBeLessThan(near.scale);
    expect(place(12, 0, size, spacing, 1.25, lens).opacity).toBe(0);
    expect(place(0, 6, size, spacing, 1.25, lens).opacity).toBe(0);
  });
});
