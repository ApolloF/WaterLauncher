// How a game will start, mirroring route() in internal/app/launch.go, so
// the interface can explain it before the game is played.
import type { Game, PadState } from "./types";

export type Route = "direct" | "store" | "steamInput";

export function routeOf(g: Game, pad: PadState): Route {
  if (g.launchUri) return "store";
  if (g.padMode === "native") return "direct";
  if (g.padMode === "steam") return "steamInput";
  if (!pad.connected || pad.kind !== "playstation" || g.padHint || !g.meta) return "direct";
  const ds = g.meta.dualSense ?? "";
  if (ds && ds !== "no") return "direct";
  const c = (g.meta.controller ?? "").toLowerCase();
  return c === "full" || c === "partial" ? "steamInput" : "direct";
}

/** The controller mode in words: a short label and a sentence. */
export function padExplain(g: Game, pad: PadState): { short: string; long: string } {
  const mode = g.padMode || "auto";
  if (g.launchUri) {
    return g.launchUri.startsWith("steam:")
      ? { short: "Steam", long: "Starts through Steam, which applies its own Steam Input settings." }
      : { short: "Store", long: "Starts through its store; the game handles the controller." };
  }
  if (mode === "native") return { short: "Native", long: "Always starts directly. The game handles the controller itself." };
  if (mode === "steam") return { short: "Steam Input", long: "Always starts through Steam Input, so a DualSense acts as an Xbox controller." };
  const ds = g.meta?.dualSense ?? "";
  if (ds === "yes") return { short: "DualSense", long: "Auto: Steam lists DualSense support, so it starts directly." };
  if (g.padHint === "libScePad") return { short: "DualSense", long: "Auto: the game ships Sony's DualSense library, so it starts directly." };
  if (g.padHint === "SDL") return { short: "Native", long: "Auto: the game uses SDL, which handles a DualSense itself, so it starts directly." };
  if (ds === "dualshock") return { short: "DualShock", long: "Auto: Steam lists DualShock support, so it starts directly." };
  if (routeOf(g, pad) === "steamInput")
    return { short: "Steam Input", long: "Auto: the game supports Xbox controllers but not PlayStation ones, so it starts through Steam Input." };
  const c = (g.meta?.controller ?? "").toLowerCase();
  if (c === "full" || c === "partial")
    return { short: "Auto", long: "Auto: the game supports Xbox controllers. With a PlayStation controller it starts through Steam Input." };
  return { short: "Auto", long: "Auto: no controller support known, so the game starts directly." };
}
