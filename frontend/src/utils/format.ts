export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes)) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB'] as const
  let v = Math.max(0, bytes)
  let idx = 0
  while (v >= 1024 && idx < units.length - 1) {
    v /= 1024
    idx++
  }
  const digits = idx === 0 ? 0 : v < 10 ? 2 : v < 100 ? 1 : 0
  return `${v.toFixed(digits)} ${units[idx]}`
}

export function formatNumber(v: number): string {
  if (!Number.isFinite(v)) return '—'
  return new Intl.NumberFormat(undefined).format(v)
}

export function formatDateTime(tsMs: number): string {
  if (!Number.isFinite(tsMs)) return '—'
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(tsMs))
}
