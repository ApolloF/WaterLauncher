import { describe, expect, it } from "vitest";
import { nearest } from "./desknav";

const rect = (x: number, y: number, w = 100, h = 40) => ({ left: x, top: y, right: x + w, bottom: y + h, width: w, height: h, x, y }) as DOMRect;
const el = (name: string) => ({ name }) as unknown as HTMLElement;

describe("nearest", () => {
  // A sidebar on the left, a grid of two rows on the right.
  const side = [el("nav1"), el("nav2")];
  const grid = [el("g1"), el("g2"), el("g3"), el("g4")];
  const rects = [
    { el: side[0], r: rect(0, 0, 200) },
    { el: side[1], r: rect(0, 50, 200) },
    { el: grid[0], r: rect(300, 0) },
    { el: grid[1], r: rect(420, 0) },
    { el: grid[2], r: rect(300, 60) },
    { el: grid[3], r: rect(420, 60) },
  ];
  it("keeps to the row and column", () => {
    expect(nearest(rects[2].r, "right", rects.slice(3))).toBe(grid[1]);
    expect(nearest(rects[2].r, "down", rects.slice(3))).toBe(grid[2]);
  });
  it("goes left to the sidebar from the grid's first column", () => {
    expect(nearest(rects[4].r, "left", rects.filter((r) => r.el !== grid[2]))).toBe(side[1]);
  });
  it("finds nothing past the edge", () => {
    expect(nearest(rects[0].r, "up", rects.slice(1))).toBeNull();
  });
});
