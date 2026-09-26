import { describe, expect, it } from "vitest";
import { savesSummary } from "./saves";
import type { SaveFolder, Saves } from "./types";

const f = (p: Partial<SaveFolder>): SaveFolder => ({
  id: "g", label: "Game", path: "", sync: true, backup: true, state: "idle", needBytes: 0, errors: 0, conflicts: 0,
  exists: true, modified: "", backedUp: new Date().toISOString(), newerOn: "", newerAt: "", ...p,
});
const s = (folders: SaveFolder[], p: Partial<Saves> = {}): Saves => ({ installed: true, available: true, known: folders.length > 0, folders, ...p });

describe("savesSummary", () => {
  it("explains each state", () => {
    expect(savesSummary(null)).toBeNull();
    expect(savesSummary(s([], { installed: false }))?.action).toBe("get");
    expect(savesSummary(s([], { available: false }))?.tone).toBe("warn");
    expect(savesSummary(s([], { outdated: true, available: false }))?.action).toBe("get");
    expect(savesSummary(s([]))?.text).toBe("Not in Syncer");
    expect(savesSummary(s([f({ conflicts: 1 })]))?.text).toBe("2 versions of a save");
    expect(savesSummary(s([f({ newerOn: "TV-PC" })]))?.text).toBe("Newer save on TV-PC");
    expect(savesSummary(s([f({})]))?.text).toBe("Synced · backed up today");
    expect(savesSummary(s([f({ state: "syncing" })]))?.text).toMatch(/^Syncing…/);
    expect(savesSummary(s([f({ sync: false, state: "backup-only" })]))?.text).toBe("Backed up today · not synced");
  });
});
