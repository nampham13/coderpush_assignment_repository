import { useState, type FormEvent } from 'react'

const example = '[\n  {"id":"SKU-101","name":"Example product","description":"Optional","price_cents":1299,"revision":1}\n]'

export function BatchForm({ disabled, onSubmit }: {
  disabled: boolean
  onSubmit: (id: string, products: unknown[]) => Promise<void>
}) {
  const [id, setID] = useState('')
  const [json, setJSON] = useState(example)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    try {
      const products: unknown = JSON.parse(json)
      if (!Array.isArray(products)) throw new Error('Products must be a JSON array.')
      if (!id.trim()) throw new Error('Enter a batch ID.')
      setSaving(true); setError('')
      await onSubmit(id.trim(), products)
      setID('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Invalid batch.')
    } finally { setSaving(false) }
  }

  return <form className="stack batch-form" onSubmit={submit}>
    <label>Batch ID<input required value={id} onChange={(event) => setID(event.target.value)} placeholder="supplier-batch-001" /></label>
    <label>Products (JSON array)<textarea required rows={7} value={json} onChange={(event) => setJSON(event.target.value)} spellCheck={false} /></label>
    {error && <p className="inline-error" role="alert">{error}</p>}
    <button disabled={disabled || saving}>{saving ? 'Submitting…' : 'Submit batch'}</button>
  </form>
}
