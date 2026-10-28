export type Product = {
  id: string
  name: string
  description?: string
  price_cents: number
  revision: number
  withdrawn: boolean
}

export type Issue = {
  id: string
  code: string
  message: string
  product_id?: string
  resolved: boolean
}

export type Change = {
  product_id: string
  kind: string
  after: Product
}

export type Update = {
  id: string
  status: string
  created_at: string
  batches: Record<string, { applied: number; stale: number; unchanged: number; rejected: number }>
  changes: Change[]
  issues: Issue[]
}

export type IncomingProduct = {
  id: string
  name?: string
  description?: string
  price_cents?: number
  revision: number
  withdrawn?: boolean
}
