// A controller in desktop mode: the stick and D-pad move keyboard focus
// around the window by position (the nearest control that way), ✕ / A
// presses what has focus, ○ / B backs out (Escape). What already reacts
// to keys (the cover grid, menus, dialogs) gets the matching key first.
import { feedback, type Intent } from "./input.svelte";

const FOCUSABLE = 'button:not([disabled]), a[href], input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]';

/** Sends a key to the focused element; true when something handled it. */
function key(k: string, init: KeyboardEventInit = {}): boolean {
  const el = (document.activeElement as HTMLElement | null) ?? document.body;
  const e = new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true, ...init });
  el.dispatchEvent(e);
  return e.defaultPrevented;
}

function visible(el: HTMLElement): boolean {
  if (el.tabIndex < 0 || el.closest("[inert], [aria-hidden='true']")) return false;
  const r = el.getBoundingClientRect();
  if (r.width < 2 || r.height < 2 || r.bottom < 0 || r.right < 0 || r.top > innerHeight || r.left > innerWidth) return false;
  const st = getComputedStyle(el);
  return st.visibility !== "hidden" && st.opacity !== "0";
}

/** Where focus may go: inside the topmost open dialog or menu, else anywhere. */
function scope(): ParentNode {
  const dialogs = document.querySelectorAll<HTMLElement>('[role="dialog"][aria-modal="true"], [role="menu"]');
  return dialogs.length ? dialogs[dialogs.length - 1] : document;
}

function candidates(): HTMLElement[] {
  return [...scope().querySelectorAll<HTMLElement>(FOCUSABLE)].filter(visible);
}

type Dir = "up" | "down" | "left" | "right";

/** The control nearest to `from` in a direction, preferring ones in line with it. */
export function nearest(from: DOMRect, dir: Dir, rects: { el: HTMLElement; r: DOMRect }[]): HTMLElement | null {
  const cx = from.left + from.width / 2;
  const cy = from.top + from.height / 2;
  let best: HTMLElement | null = null;
  let bestScore = Infinity;
  for (const { el, r } of rects) {
    let ahead: number, side: number;
    switch (dir) {
      case "up":
        ahead = from.top - r.bottom;
        side = Math.max(0, r.left - from.right, from.left - r.right);
        break;
      case "down":
        ahead = r.top - from.bottom;
        side = Math.max(0, r.left - from.right, from.left - r.right);
        break;
      case "left":
        ahead = from.left - r.right;
        side = Math.max(0, r.top - from.bottom, from.top - r.bottom);
        break;
      default:
        ahead = r.left - from.right;
        side = Math.max(0, r.top - from.bottom, from.top - r.bottom);
    }
    // Overlapping a little still counts as that way, if the centre is.
    const centre = dir === "up" ? r.top + r.height / 2 < cy : dir === "down" ? r.top + r.height / 2 > cy : dir === "left" ? r.left + r.width / 2 < cx : r.left + r.width / 2 > cx;
    if (ahead < -8 || !centre) continue;
    const score = Math.max(0, ahead) + side * 3;
    if (score < bestScore) {
      bestScore = score;
      best = el;
    }
  }
  return best;
}

function start(): HTMLElement | null {
  return (
    document.querySelector<HTMLElement>('[role="dialog"][aria-modal="true"] [tabindex="0"], [role="dialog"][aria-modal="true"] button') ??
    document.querySelector<HTMLElement>('[role="listbox"]') ??
    candidates()[0] ??
    null
  );
}

function focus(el: HTMLElement) {
  el.focus({ preventScroll: true });
  el.scrollIntoView({ block: "nearest", inline: "nearest" });
}

function move(dir: Dir) {
  const cur = document.activeElement as HTMLElement | null;
  if (!cur || cur === document.body || !visible(cur)) {
    const el = start();
    if (el) (focus(el), feedback.move());
    return;
  }
  // A list or grid moves its own selection first.
  const arrow = { up: "ArrowUp", down: "ArrowDown", left: "ArrowLeft", right: "ArrowRight" }[dir];
  if (cur.tagName === "SELECT" && (dir === "left" || dir === "right")) {
    const s = cur as HTMLSelectElement;
    const j = s.selectedIndex + (dir === "left" ? -1 : 1);
    if (j >= 0 && j < s.options.length) {
      s.selectedIndex = j;
      s.dispatchEvent(new Event("change", { bubbles: true }));
      feedback.move();
    } else feedback.edge();
    return;
  }
  if (cur.tagName !== "INPUT" && key(arrow)) {
    feedback.move();
    return;
  }
  const rects = candidates()
    .filter((el) => el !== cur && !cur.contains(el) && !el.contains(cur))
    .map((el) => ({ el, r: el.getBoundingClientRect() }));
  const next = nearest(cur.getBoundingClientRect(), dir, rects);
  if (next) {
    focus(next);
    feedback.move();
  } else feedback.edge();
}

function press() {
  const el = document.activeElement as HTMLElement | null;
  if (!el || el === document.body) return move("down");
  const role = el.getAttribute("role");
  if (el.tagName === "BUTTON" || el.tagName === "A" || role === "switch" || role === "menuitem" || role === "tab" || (el.tagName === "INPUT" && ((el as HTMLInputElement).type === "checkbox" || (el as HTMLInputElement).type === "radio"))) {
    feedback.confirm();
    el.click();
  } else if (key("Enter")) feedback.confirm();
}

/** Scrolls whatever scrolls around the focused control by most of a page. */
function page(d: -1 | 1) {
  let el = document.activeElement as HTMLElement | null;
  while (el && el !== document.body) {
    const st = getComputedStyle(el);
    if (/(auto|scroll)/.test(st.overflowY) && el.scrollHeight > el.clientHeight) {
      el.scrollBy({ top: d * el.clientHeight * 0.8, behavior: "smooth" });
      return;
    }
    el = el.parentElement;
  }
}

/** Steps through the sidebar's views (All games, Favorites, …). */
function sidebar(d: -1 | 1) {
  const items = [...document.querySelectorAll<HTMLElement>(".side nav button.nav")];
  const at = items.findIndex((b) => b.classList.contains("active"));
  const next = items[at + d];
  if (next) {
    next.click();
    feedback.move();
  } else feedback.edge();
}

let showing = false;
function showFocus(on: boolean) {
  if (on === showing) return;
  showing = on;
  document.documentElement.classList.toggle("pad-nav", on);
}
if (typeof window !== "undefined") {
  // The mouse takes over again: no focus ring.
  window.addEventListener("mousemove", (e) => (e.movementX || e.movementY) && showFocus(false), { passive: true });
  window.addEventListener("mousedown", () => showFocus(false), { passive: true });
}

/** Handles one controller action in desktop mode. */
export function desktopPad(intent: Intent, repeat: boolean) {
  showFocus(true);
  switch (intent) {
    case "up":
    case "down":
    case "left":
    case "right":
      return move(intent);
    case "confirm":
      if (!repeat) press();
      return;
    case "back":
      if (!repeat && !key("Escape")) {
        // Nothing to close: back to the games.
        const grid = document.querySelector<HTMLElement>('[role="listbox"]');
        if (grid && document.activeElement !== grid) (focus(grid), feedback.move());
      }
      return;
    case "view": // search
      if (!repeat) {
        const s = document.querySelector<HTMLInputElement>('input[type="search"]');
        if (s) (focus(s), feedback.move());
      }
      return;
    case "menu": // settings
      if (!repeat) key(",", { ctrlKey: true });
      return;
    case "lb":
    case "rb":
      if (!repeat) sidebar(intent === "lb" ? -1 : 1);
      return;
    case "lt":
    case "rt":
      return page(intent === "lt" ? -1 : 1);
  }
}
