import { describe, expect, it } from "vitest";
import { ago, bytes, playtime, clock } from "./format";

describe("format", () => {
  it("playtime", () => {
    expect(playtime(0)).toBe("Not played");
    expect(playtime(600)).toBe("10 min");
    expect(playtime(3.5 * 3600)).toBe("3.5 h");
    expect(playtime(297 * 3600)).toBe("297 h");
  });
  it("bytes", () => {
    expect(bytes(83.2e9)).toBe("83.2 GB");
    expect(bytes(512)).toBe("512 B");
    expect(bytes(0)).toBe("");
  });
  it("ago", () => {
    const now = new Date(2026, 8, 25, 21, 0).getTime() / 1000;
    expect(ago(now - 60, now)).toBe("Today");
    expect(ago(now - 86400, now)).toBe("Yesterday");
    expect(ago(now - 3 * 86400, now)).toBe("3 days ago");
    expect(ago(now - 400 * 86400, now)).toBe("Last year");
    expect(ago(0, now)).toBe("Never");
  });
});

describe("clock", () => {
  it("formats a running time", () => {
    expect(clock(0)).toBe("0:00");
    expect(clock(65)).toBe("1:05");
    expect(clock(3723)).toBe("1:02:03");
  });
});
