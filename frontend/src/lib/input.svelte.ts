// Big picture input: controller actions (from the Go side) and keyboard keys
// become one set of intents, delivered to the top-most screen that wants
// them. Screens register with useInput() inside an $effect, so an overlay
// opened on top gets input first and gives it back when it closes.
import { api } from "./api";
import { lib } from "./store.svelte";
import type { PadState } from "./types";

export type Intent =
  | "up"
  | "down"
  | "left"
  | "right"
  | "confirm"
  | "back"
  | "action"
  | "info"
  | "menu"
  | "view"
  | "home"
  | "lb"
  | "rb"
  | "lt"
  | "rt";

/** A handler returns false to let the screen below handle the intent. */
export type Handler = (intent: Intent, repeat: boolean) => boolean | void;

const stack: Handler[] = [];

export function useInput(h: Handler): () => void {
  stack.push(h);
  return () => {
    const i = stack.lastIndexOf(h);
    if (i >= 0) stack.splice(i, 1);
  };
}

let base: Handler | null = null;

/** The handler under every screen (menu, home), set by the big picture shell. */
export function setBase(h: Handler | null) {
  base = h;
}

export function dispatch(intent: Intent, repeat = false) {
  for (let i = stack.length - 1; i >= 0; i--) {
    if (stack[i](intent, repeat) !== false) return;
  }
  base?.(intent, repeat);
}

const KEYS: Record<string, Intent> = {
  ArrowUp: "up",
  ArrowDown: "down",
  ArrowLeft: "left",
  ArrowRight: "right",
  Enter: "confirm",
  " ": "confirm",
  Escape: "back",
  Backspace: "back",
  x: "action",
  X: "action",
  i: "info",
  I: "info",
  y: "info",
  Y: "info",
  q: "menu",
  Q: "menu",
  m: "menu",
  M: "menu",
  Home: "home",
  PageUp: "lb",
  PageDown: "rb",
  "[": "lt",
  "]": "rt",
  "-": "lt",
  "=": "rt",
  "+": "rt",
};

/** Maps a key to an intent, unless the user is typing in a field. */
export function keyIntent(e: KeyboardEvent): Intent | null {
  const t = e.target as HTMLElement | null;
  if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA") && e.key !== "Escape" && e.key !== "ArrowDown" && e.key !== "ArrowUp") return null;
  if (e.ctrlKey || e.altKey || e.metaKey) return null;
  return KEYS[e.key] ?? null;
}

// ---- controller state and feedback ----

export const pad = $state<PadState>({ connected: false, name: "", kind: "other", dualSense: false, battery: -1, wireless: false });

/** Which glyph set to show: the user's choice, else the controller in use. */
export function glyphSet(): "playstation" | "xbox" {
  const g = lib.settings?.glyphs ?? "auto";
  if (g === "playstation" || g === "xbox") return g;
  return pad.connected && pad.kind === "playstation" ? "playstation" : pad.connected ? "xbox" : "playstation";
}

let audio: AudioContext | null = null;
function beep(freq: number, ms: number, gain = 0.035) {
  if (!lib.settings?.sounds) return;
  try {
    audio ??= new AudioContext();
    const o = audio.createOscillator();
    const g = audio.createGain();
    o.type = "sine";
    o.frequency.value = freq;
    g.gain.setValueAtTime(gain, audio.currentTime);
    g.gain.exponentialRampToValueAtTime(0.0001, audio.currentTime + ms / 1000);
    o.connect(g).connect(audio.destination);
    o.start();
    o.stop(audio.currentTime + ms / 1000);
  } catch {
    /* no audio device */
  }
}

/** Feedback for moving focus, confirming and errors: rumble and sound. */
export const feedback = {
  move() {
    api.pad.rumble("tick");
    beep(1250, 30);
  },
  confirm() {
    api.pad.rumble("confirm");
    beep(880, 60, 0.05);
    setTimeout(() => beep(1320, 70, 0.04), 55);
  },
  error() {
    api.pad.rumble("error");
    beep(220, 140, 0.05);
  },
};

let lastLight = "";
/** Tints the DualSense lightbar (the Go side checks the user's setting). */
export function setLight(css: string | undefined) {
  const hex = css ? toHex(css) : "";
  if (!hex || hex === lastLight) return;
  lastLight = hex;
  api.pad.setLight(hex);
}

let ctx: CanvasRenderingContext2D | null = null;
/** Any CSS colour (oklch too) as #rrggbb. */
export function toHex(css: string): string {
  if (/^#[0-9a-f]{6}$/i.test(css)) return css.toLowerCase();
  ctx ??= document.createElement("canvas").getContext("2d", { willReadFrequently: true });
  if (!ctx) return "";
  ctx.clearRect(0, 0, 1, 1);
  ctx.fillStyle = "#000";
  ctx.fillStyle = css;
  ctx.fillRect(0, 0, 1, 1);
  const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data;
  return "#" + [r, g, b].map((v) => v.toString(16).padStart(2, "0")).join("");
}
