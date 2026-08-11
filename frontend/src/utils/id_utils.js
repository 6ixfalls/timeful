export const isLongId = (value) => /^[0-9a-f]{24}$/i.test(value)

export const dateFromLongId = (value) => {
  if (!isLongId(value)) return null
  return new Date(parseInt(value.slice(0, 8), 16) * 1000)
}

// Keeps Timeful's established 24-hex-character public ID format without
// depending on a database-specific client package.
export const createLongId = () => {
  const bytes = new Uint8Array(12)
  const seconds = Math.floor(Date.now() / 1000)
  bytes[0] = seconds >>> 24
  bytes[1] = seconds >>> 16
  bytes[2] = seconds >>> 8
  bytes[3] = seconds
  crypto.getRandomValues(bytes.subarray(4))
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("")
}
