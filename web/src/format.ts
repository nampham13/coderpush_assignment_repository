const statusLabels: Record<string, string> = {
  in_progress: 'In progress',
  interrupted: 'Interrupted',
  completed: 'Completed',
}

export const statusLabel = (status: string) =>
  statusLabels[status] ?? status.charAt(0).toUpperCase() + status.slice(1).split('_').join(' ')

export const formatPrice = (cents: number) =>
  (cents / 100).toLocaleString(undefined, { style: 'currency', currency: 'USD' })

export const formatDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}
