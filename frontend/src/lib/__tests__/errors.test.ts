import { describe, it, expect } from "vitest";
import { mapAuthError, BeeboxApiError } from "../beebox";

describe("mapAuthError", () => {
  it("maps 401 to credentials error", () => {
    const err = new BeeboxApiError("UNAUTHENTICATED", "bad", 401);
    const mapped = mapAuthError(err);
    expect(mapped.kind).toBe("credentials");
    expect(mapped.title).toMatch(/incorrect/i);
  });
  it("maps 503 to unavailable", () => {
    const err = new BeeboxApiError("SERVICE_UNAVAILABLE", "down", 503);
    expect(mapAuthError(err).kind).toBe("unavailable");
  });
  it("maps 409 to validation", () => {
    const err = new BeeboxApiError("CONFLICT", "exists", 409);
    expect(mapAuthError(err).kind).toBe("validation");
  });
  it("falls back for unknown errors", () => {
    expect(mapAuthError(new Error("boom")).kind).toBe("generic");
  });
});
