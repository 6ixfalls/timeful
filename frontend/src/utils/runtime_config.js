export const DEFAULT_PRIVACY_POLICY_URL =
  "https://docs.google.com/document/d/1WfcSi7FkNACWsm31sUiIcgpqIWWJ0mMm7qqma_x0kIM/preview"

export const getPrivacyPolicyUrl = (configs = globalThis.window?.configs) => {
  const url = configs?.VUE_APP_PRIVACY_POLICY_URL
  const customUrl = typeof url === "string" ? url.trim() : ""
  return customUrl || DEFAULT_PRIVACY_POLICY_URL
}
