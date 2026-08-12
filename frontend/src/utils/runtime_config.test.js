import { describe, expect, it } from "vitest"
import {
  DEFAULT_PRIVACY_POLICY_URL,
  getPrivacyPolicyUrl,
} from "./runtime_config"

describe("getPrivacyPolicyUrl", () => {
  it("returns the configured URL", () => {
    expect(
      getPrivacyPolicyUrl({
        VUE_APP_PRIVACY_POLICY_URL: "https://example.com/privacy",
      }),
    ).toBe("https://example.com/privacy")
  })

  it("trims whitespace from the configured URL", () => {
    expect(
      getPrivacyPolicyUrl({
        VUE_APP_PRIVACY_POLICY_URL: "  https://example.com/privacy  ",
      }),
    ).toBe("https://example.com/privacy")
  })

  it("returns the bundled policy when no override is configured", () => {
    expect(getPrivacyPolicyUrl({})).toBe(DEFAULT_PRIVACY_POLICY_URL)
  })

  it("returns the bundled policy when the override is blank", () => {
    expect(
      getPrivacyPolicyUrl({ VUE_APP_PRIVACY_POLICY_URL: "   " }),
    ).toBe(DEFAULT_PRIVACY_POLICY_URL)
  })
})
