import { describe, expect, it } from "vitest";
import { cn } from "./utils";

describe("cn", () => {
  it("should merge class names", () => {
    const result = cn("px-2 py-1", "bg-red-500");
    expect(result).toBe("px-2 py-1 bg-red-500");
  });

  it("should handle conflicting Tailwind classes", () => {
    const result = cn("px-2", "px-4");
    expect(result).toBe("px-4");
  });

  it("should handle conditional classes", () => {
    const isActive = true;
    const result = cn("base-class", isActive && "active-class");
    expect(result).toBe("base-class active-class");
  });

  it("should filter out falsy values", () => {
    const result = cn("base", false, null, undefined, "end");
    expect(result).toBe("base end");
  });
});
