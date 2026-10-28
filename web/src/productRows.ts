import type { IncomingProduct } from './types'

// A product row as typed by staff: every field is text until it is validated.
export type ProductRow = { id: string; name: string; description: string; price: string; revision: string; withdrawn: boolean }

export const emptyRow = (): ProductRow => ({ id: '', name: '', description: '', price: '', revision: '', withdrawn: false })

export const exampleRows = (): ProductRow[] => [
  { id: 'SKU-101', name: 'Example product', description: 'Optional description', price: '12.99', revision: '1', withdrawn: false },
  { id: 'SKU-102', name: '', description: '', price: '', revision: '2', withdrawn: true },
]

// toProducts converts rows into API products, or returns plain-language errors per row.
export function toProducts(rows: ProductRow[]): { products: IncomingProduct[]; errors: string[] } {
  const errors: string[] = []
  const products: IncomingProduct[] = []
  const seen = new Set<string>()
  rows.forEach((row, index) => {
    const label = `Row ${index + 1}`
    const id = row.id.trim()
    const revision = Number(row.revision.trim())
    const price = Number(row.price.trim())
    const problems: string[] = []
    if (!id) problems.push('enter a product ID')
    else if (seen.has(id)) problems.push(`product ID ${id} is already used in another row`)
    if (!Number.isInteger(revision) || revision < 1) problems.push('revision must be a whole number of 1 or more')
    if (!row.withdrawn) {
      if (!row.name.trim()) problems.push('enter a name')
      if (!row.price.trim() || !Number.isFinite(price) || price < 0) problems.push('price must be a number of 0 or more, such as 12.99')
    }
    if (problems.length) { errors.push(`${label}: ${problems.join('; ')}.`); return }
    seen.add(id)
    products.push(row.withdrawn
      ? { id, revision, withdrawn: true }
      : { id, name: row.name.trim(), description: row.description.trim(), price_cents: Math.round(price * 100), revision })
  })
  if (!rows.length) errors.push('Add at least one product.')
  return { products, errors }
}

const csvColumns = ['id', 'name', 'description', 'price', 'revision', 'withdrawn']
export const csvTemplate = `${csvColumns.join(',')}\nSKU-101,Example product,Optional description,12.99,1,false\n`

// parseCSV reads a CSV with a header row containing id and revision (other columns optional).
export function parseCSV(text: string): ProductRow[] {
  const records = splitCSV(text).filter((record) => record.some((cell) => cell.trim()))
  if (!records.length) throw new Error('The CSV file is empty.')
  const header = records[0].map((cell) => cell.trim().toLowerCase())
  const missing = ['id', 'revision'].filter((column) => !header.includes(column))
  if (missing.length) throw new Error(`The CSV header must include: ${missing.join(', ')}. Expected columns: ${csvColumns.join(', ')}.`)
  const cell = (record: string[], column: string) => {
    const index = header.indexOf(column)
    return index < 0 ? '' : (record[index] ?? '').trim()
  }
  return records.slice(1).map((record) => {
    const cents = cell(record, 'price_cents')
    return {
      id: cell(record, 'id'), name: cell(record, 'name'), description: cell(record, 'description'),
      price: cell(record, 'price') || (cents && Number.isFinite(Number(cents)) ? String(Number(cents) / 100) : cents),
      revision: cell(record, 'revision'),
      withdrawn: ['true', 'yes', '1', 'y'].includes(cell(record, 'withdrawn').toLowerCase()),
    }
  })
}

function splitCSV(text: string): string[][] {
  const records: string[][] = []
  let record: string[] = [], cell = '', quoted = false
  for (let index = 0; index < text.length; index++) {
    const char = text[index]
    if (quoted) {
      if (char === '"' && text[index + 1] === '"') { cell += '"'; index++ }
      else if (char === '"') quoted = false
      else cell += char
    } else if (char === '"') quoted = true
    else if (char === ',') { record.push(cell); cell = '' }
    else if (char === '\n' || char === '\r') {
      if (char === '\r' && text[index + 1] === '\n') index++
      record.push(cell); records.push(record); record = []; cell = ''
    } else cell += char
  }
  if (cell || record.length) { record.push(cell); records.push(record) }
  return records
}
