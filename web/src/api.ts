import type { Product, Update } from './types'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: string } | null
    throw new Error(body?.error || `Request failed (${response.status})`)
  }
  return response.json() as Promise<T>
}

export async function loadDashboard(includeWithdrawn: boolean) {
  const [health, catalog, updates] = await Promise.all([
    request<{ status: string }>('/health'),
    request<{ products: Product[] }>(`/catalog?include_withdrawn=${includeWithdrawn}`),
    request<{ updates: Update[] }>('/updates'),
  ])
  return { online: health.status === 'ok', products: catalog.products, updates: updates.updates }
}

export const createUpdate = (id: string) =>
  request<Update>('/updates', { method: 'POST', body: JSON.stringify({ id }) })
export const runDemo = () => request<Update>('/demo', { method: 'POST', body: '{}' })
export const resumeUpdate = (id: string) =>
  request<Update>(`/updates/${encodeURIComponent(id)}/resume`, { method: 'POST', body: '{}' })
export const completeUpdate = (id: string) =>
  request<Update>(`/updates/${encodeURIComponent(id)}/complete`, { method: 'POST', body: '{}' })
export const reportOutage = (id: string, message: string) =>
  request<Update>(`/updates/${encodeURIComponent(id)}/failures`, {
    method: 'POST', body: JSON.stringify({ message }),
  })
export const submitBatch = (id: string, batch: unknown) =>
  request<Update>(`/updates/${encodeURIComponent(id)}/batches`, {
    method: 'POST', body: JSON.stringify(batch),
  })
export const resolveIssue = (updateID: string, issueID: string) =>
  request<Update>(`/updates/${encodeURIComponent(updateID)}/issues/${encodeURIComponent(issueID)}/resolve`, {
    method: 'POST', body: '{}',
  })
