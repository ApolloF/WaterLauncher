// Generated placeholder art: a small landscape per game (mountains, sea,
// synthwave, space, forest, dunes or a city), seeded by the game so it is
// always the same. Used until real art is downloaded, and for games no
// art source knows.

export interface Art {
  sky: string;
  stars: string;
  starOp: number;
  sun: string;
  sunMask: string;
  sunL: string;
  sunT: string;
  sunW: string;
  far: string;
  farClip: string;
  mid: string;
  midClip: string;
  near: string;
  nearClip: string;
  fog: string;
  floor: string;
  floorH: string;
  accent: string;
  accentSoft: string;
}

const SCENES = ["mountains", "sea", "synth", "space", "forest", "dunes", "city"] as const;
const NONE = "polygon(0 0,0 0,0 0)";

export function hash(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

function rng(a: number) {
  return () => {
    a |= 0;
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const ok = (l: number, c: number, h: number, a?: number) =>
  `oklch(${l} ${c} ${(((h % 360) + 360) % 360).toFixed(0)}${a != null ? ` / ${a}` : ""})`;

const poly = (p: [number, number][]) =>
  `polygon(0% 100%,${p.map(([x, y]) => `${x.toFixed(1)}% ${y.toFixed(1)}%`).join(",")},100% 100%)`;

function peaks(r: () => number, b: number, a: number, s: number) {
  const p: [number, number][] = [];
  let x = 0;
  let up = r() > 0.5;
  while (x < 100) {
    p.push([x, up ? b - a * (0.45 + r() * 0.55) : b + a * 0.15 * r()]);
    up = !up;
    x += s * (0.6 + r() * 0.8);
  }
  p.push([100, b - a * 0.3 * r()]);
  return poly(p);
}

function waves(r: () => number, b: number, a: number, f: number) {
  const p: [number, number][] = [];
  const ph = r() * 6.28;
  for (let x = 0; x <= 100; x += 2) p.push([x, b + Math.sin((x * f) / 10 + ph) * a + Math.sin((x * f) / 3.7 + ph * 1.7) * a * 0.3]);
  return poly(p);
}

function teeth(r: () => number, b: number, a: number) {
  const p: [number, number][] = [];
  for (let x = 0; x < 100; x += 1.8) {
    p.push([x, b + a * 0.25]);
    p.push([x + 0.9, b - a * (0.35 + r() * 0.65)]);
  }
  p.push([100, b]);
  return poly(p);
}

function city(r: () => number, b: number, a: number) {
  const p: [number, number][] = [];
  let x = 0;
  while (x < 100) {
    const w = 1.8 + r() * 5;
    const y = b - a * (0.15 + r() * 0.85);
    p.push([x, y]);
    x = Math.min(100, x + w);
    p.push([x, y]);
    x += r() * 0.6;
  }
  return poly(p);
}

function stars(r: () => number, n: number, maxY: number) {
  const s: string[] = [];
  for (let i = 0; i < n; i++) {
    const z = (0.6 + r() * 1.6).toFixed(1);
    s.push(`radial-gradient(${z}px ${z}px at ${(r() * 100).toFixed(1)}% ${(r() * maxY).toFixed(1)}%,rgba(255,255,255,${(0.45 + r() * 0.55).toFixed(2)}) 50%,transparent 51%)`);
  }
  return s.join(",");
}

const cache = new Map<string, Art>();

/** Art for a game, keyed by something stable (its install key). */
export function artFor(key: string): Art {
  const hit = cache.get(key);
  if (hit) return hit;
  const seed = hash(key);
  const r = rng(seed);
  // Draw hue and scene from the generator, not the raw hash: keys that
  // differ only at the end (…\Game 1, …\Game 2) must still look different.
  r();
  const h = Math.floor(r() * 360);
  const scene = SCENES[Math.floor(r() * SCENES.length)];
  const a = {} as Art;
  a.sky = `linear-gradient(180deg,${ok(0.22, 0.06, h)} 0%,${ok(0.45, 0.11, h + 25)} 55%,${ok(0.82, 0.1, h + 60)} 100%)`;
  a.stars = "none";
  a.starOp = 0;
  let sl = 38 + r() * 32;
  let st = 10 + r() * 14;
  let sw = 30;
  a.sun = `radial-gradient(circle closest-side,${ok(0.97, 0.05, h + 70)} 0 42%,${ok(0.9, 0.12, h + 55, 0.5)} 48%,transparent 100%)`;
  a.sunMask = "none";
  a.far = ok(0.52, 0.08, h + 15);
  a.mid = ok(0.34, 0.07, h + 5);
  a.near = ok(0.16, 0.035, h - 10);
  a.fog = `linear-gradient(0deg,${ok(0.78, 0.09, h + 50, 0.55)} 0%,transparent 45%)`;
  a.floorH = "0%";
  a.floor = "none";
  a.farClip = a.midClip = a.nearClip = NONE;
  switch (scene) {
    case "mountains":
      a.farClip = peaks(r, 56, 26, 9);
      a.midClip = peaks(r, 70, 20, 7);
      a.nearClip = peaks(r, 86, 12, 6);
      break;
    case "forest":
      a.farClip = peaks(r, 54, 18, 12);
      a.midClip = teeth(r, 70, 12);
      a.nearClip = teeth(r, 88, 16);
      st = 8;
      break;
    case "dunes":
      a.sky = `linear-gradient(180deg,${ok(0.55, 0.09, h + 200)} 0%,${ok(0.8, 0.08, h + 20)} 60%,${ok(0.9, 0.09, h + 10)} 100%)`;
      a.far = ok(0.72, 0.1, h);
      a.mid = ok(0.6, 0.12, h - 5);
      a.near = ok(0.45, 0.12, h - 12);
      a.farClip = waves(r, 64, 4, 1.2);
      a.midClip = waves(r, 74, 5, 0.9);
      a.nearClip = waves(r, 88, 4, 0.7);
      a.fog = `linear-gradient(0deg,${ok(0.92, 0.06, h + 20, 0.35)} 0%,transparent 40%)`;
      sw = 24;
      st = 16;
      break;
    case "city":
      a.sky = `linear-gradient(180deg,${ok(0.12, 0.03, h)} 0%,${ok(0.26, 0.09, h + 15)} 60%,${ok(0.52, 0.16, h + 35)} 100%)`;
      a.stars = stars(r, 30, 50);
      a.starOp = 0.8;
      sw = 11;
      sl = 66 + r() * 14;
      st = 9;
      a.sun = "radial-gradient(circle closest-side,#f3efe6 0 62%,rgba(243,239,230,.22) 68%,transparent 100%)";
      a.far = ok(0.3, 0.06, h + 10);
      a.mid = ok(0.19, 0.05, h);
      a.near = ok(0.1, 0.02, h);
      a.farClip = city(r, 60, 26);
      a.midClip = city(r, 74, 24);
      a.nearClip = "polygon(0% 100%,0% 94%,100% 94%,100% 100%)";
      a.fog = `linear-gradient(0deg,${ok(0.62, 0.2, h + 30, 0.5)} 0%,transparent 42%)`;
      break;
    case "sea": {
      a.farClip = peaks(r, 60, 12, 14);
      a.fog = "none";
      a.floorH = "40%";
      st = 26 + r() * 8;
      const cx = (sl + sw / 2 + 50) / 2;
      a.floor = `radial-gradient(ellipse 5% 75% at ${cx.toFixed(1)}% 0%,${ok(0.96, 0.08, h + 60, 0.75)},transparent 70%),repeating-linear-gradient(180deg,rgba(255,255,255,.09) 0 1px,transparent 1px 7px),linear-gradient(180deg,${ok(0.56, 0.1, h + 20)},${ok(0.17, 0.05, h)})`;
      break;
    }
    case "synth": {
      a.sky = `linear-gradient(180deg,${ok(0.14, 0.06, h - 40)} 0%,${ok(0.3, 0.15, h)} 62%,${ok(0.55, 0.2, h + 20)} 100%)`;
      a.stars = stars(r, 30, 55);
      a.starOp = 0.9;
      sl = 33;
      st = 12;
      sw = 34;
      a.sun = `linear-gradient(180deg,${ok(0.93, 0.14, 95)} 0%,${ok(0.78, 0.19, 40)} 45%,${ok(0.66, 0.25, h + 20)} 100%)`;
      a.sunMask = "linear-gradient(180deg,#000 0 50%,transparent 50% 54%,#000 54% 62%,transparent 62% 67%,#000 67% 74%,transparent 74% 80%,#000 80% 86%,transparent 86% 92%,#000 92%)";
      a.far = ok(0.24, 0.12, h - 10);
      a.farClip = peaks(r, 58, 14, 8);
      a.fog = "none";
      a.floorH = "42%";
      const P = ok(0.72, 0.24, h + 10, 0.8);
      a.floor = `linear-gradient(180deg,transparent 0 2%,${P} 2% 2.8%,transparent 2.8% 7%,${P} 7% 8%,transparent 8% 15%,${P} 15% 16.4%,transparent 16.4% 27%,${P} 27% 28.8%,transparent 28.8% 44%,${P} 44% 46.2%,transparent 46.2% 66%,${P} 66% 68.8%,transparent 68.8% 92%,${P} 92% 95.5%,transparent 95.5%),repeating-conic-gradient(from 180deg at 50% -30%,${P} 0deg 0.35deg,transparent 0.35deg 3.2deg),linear-gradient(180deg,${ok(0.24, 0.13, h)},${ok(0.1, 0.05, h - 30)})`;
      break;
    }
    case "space":
      a.sky = `radial-gradient(120% 90% at 25% 20%,${ok(0.3, 0.1, h)} 0%,${ok(0.13, 0.05, h + 20)} 55%,#05060a 100%)`;
      a.stars = stars(r, 40, 100);
      a.starOp = 1;
      sl = 48;
      st = 34;
      sw = 72;
      a.sun = `radial-gradient(circle at 30% 28%,${ok(0.88, 0.08, h + 40)} 0%,${ok(0.55, 0.13, h + 10)} 28%,${ok(0.24, 0.09, h - 15)} 55%,${ok(0.1, 0.03, h - 30)} 72%)`;
      a.fog = "none";
      a.near = ok(0.12, 0.02, h);
      a.nearClip = waves(r, 92, 2.5, 0.5);
      break;
  }
  a.sunL = `${sl.toFixed(1)}%`;
  a.sunT = `${st.toFixed(1)}%`;
  a.sunW = `${sw}%`;
  a.accent = ok(0.8, 0.13, h + 40);
  a.accentSoft = ok(0.8, 0.13, h + 40, 0.45);
  cache.set(key, a);
  return a;
}
