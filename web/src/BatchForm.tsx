import { useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { csvTemplate, emptyRow, exampleRows, parseCSV, toProducts, type ProductRow } from './productRows'
import type { IncomingProduct } from './types'

const templateHref = `data:text/csv;charset=utf-8,${encodeURIComponent(csvTemplate)}`

export function BatchForm({ disabled, onSubmit }: {
  disabled: boolean
  onSubmit: (id: string, products: IncomingProduct[]) => Promise<boolean>
}) {
  const [id, setID] = useState('')
  const [rows, setRows] = useState<ProductRow[]>([emptyRow()])
  const [errors, setErrors] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const fileInput = useRef<HTMLInputElement>(null)

  const updateRow = (index: number, patch: Partial<ProductRow>) =>
    setRows((current) => current.map((row, position) => position === index ? { ...row, ...patch } : row))

  async function importCSV(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return
    try {
      const imported = parseCSV(await file.text())
      if (!imported.length) throw new Error('The CSV file has a header but no products.')
      setRows(imported); setErrors([])
    } catch (reason) {
      setErrors([reason instanceof Error ? reason.message : 'Could not read the CSV file.'])
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    const { products, errors: problems } = toProducts(rows)
    if (!id.trim()) problems.unshift('Enter a batch ID.')
    setErrors(problems)
    if (problems.length) return
    setSaving(true)
    try {
      if (await onSubmit(id.trim(), products)) { setID(''); setRows([emptyRow()]) }
    } finally { setSaving(false) }
  }

  return <form className="stack batch-form" onSubmit={submit} noValidate>
    <div className="form-heading"><h3>Submit a batch</h3>
      <div className="form-tools">
        <button type="button" className="link" onClick={() => { setRows(exampleRows()); setErrors([]) }}>Use example</button>
        <button type="button" className="link" onClick={() => fileInput.current?.click()}>Import CSV</button>
        <a className="link" href={templateHref} download="products-template.csv">CSV template</a>
        <input ref={fileInput} type="file" accept=".csv,text/csv" hidden aria-label="CSV file" onChange={(event) => void importCSV(event)} />
      </div></div>
    <label>Batch ID<input value={id} onChange={(event) => setID(event.target.value)} placeholder="supplier-batch-001" /></label>
    <ol className="product-rows">{rows.map((row, index) =>
      <ProductRowFields key={index} row={row} index={index} canRemove={rows.length > 1}
        onChange={(patch) => updateRow(index, patch)}
        onRemove={() => setRows((current) => current.filter((_, position) => position !== index))} />)}
    </ol>
    <button type="button" className="secondary" onClick={() => setRows((current) => [...current, emptyRow()])}>Add product</button>
    {errors.length > 0 && <div className="inline-error" role="alert"><strong>Fix these before submitting:</strong>
      <ul>{errors.map((error) => <li key={error}>{error}</li>)}</ul></div>}
    <button disabled={disabled || saving}>{saving ? 'Submitting…' : `Submit batch (${rows.length} ${rows.length === 1 ? 'product' : 'products'})`}</button>
  </form>
}

function ProductRowFields({ row, index, canRemove, onChange, onRemove }: {
  row: ProductRow; index: number; canRemove: boolean
  onChange: (patch: Partial<ProductRow>) => void; onRemove: () => void
}) {
  const n = index + 1
  return <li className="product-row">
    <div className="row-heading"><strong>Product {n}</strong>
      {canRemove && <button type="button" className="link" aria-label={`Remove product ${n}`} onClick={onRemove}>Remove</button>}</div>
    <div className="row-grid">
      <label>Product ID<input aria-label={`Product ${n} ID`} value={row.id} onChange={(event) => onChange({ id: event.target.value })} placeholder="SKU-101" /></label>
      <label>Revision<input aria-label={`Product ${n} revision`} inputMode="numeric" value={row.revision} onChange={(event) => onChange({ revision: event.target.value })} placeholder="1" /></label>
      <label className="wide">Name<input aria-label={`Product ${n} name`} value={row.name} disabled={row.withdrawn} onChange={(event) => onChange({ name: event.target.value })} /></label>
      <label>Price (USD)<input aria-label={`Product ${n} price`} inputMode="decimal" value={row.price} disabled={row.withdrawn} onChange={(event) => onChange({ price: event.target.value })} placeholder="12.99" /></label>
      <label className="checkbox"><input type="checkbox" aria-label={`Withdraw product ${n}`} checked={row.withdrawn} onChange={(event) => onChange({ withdrawn: event.target.checked })} /> Withdraw product</label>
      <label className="wide"><span>Description <span className="optional">(optional)</span></span><input aria-label={`Product ${n} description`} value={row.description} disabled={row.withdrawn} onChange={(event) => onChange({ description: event.target.value })} /></label>
    </div>
  </li>
}
