import { useMemo, useState } from 'react'
import { formatPrice } from './format'
import type { Product } from './types'

type Props = { products: Product[]; loading: boolean; onIncludeWithdrawn: (value: boolean) => void }
type SortKey = 'name' | 'id' | 'price_cents' | 'revision' | 'withdrawn'
type Sort = { key: SortKey; ascending: boolean }

const pageSize = 20
const columns: { key: SortKey; label: string }[] = [
  { key: 'name', label: 'Product' }, { key: 'id', label: 'ID' }, { key: 'price_cents', label: 'Price' },
  { key: 'revision', label: 'Revision' }, { key: 'withdrawn', label: 'Status' },
]

function compare(a: Product, b: Product, key: SortKey) {
  const left = a[key], right = b[key]
  if (typeof left === 'string' && typeof right === 'string') return left.localeCompare(right, undefined, { numeric: true })
  return Number(left) - Number(right)
}

export function ProductsPanel({ products, loading, onIncludeWithdrawn }: Props) {
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<Sort>({ key: 'name', ascending: true })
  const [page, setPage] = useState(0)
  const visible = useMemo(() => {
    const query = search.trim().toLowerCase()
    const matches = products.filter((product) => `${product.id} ${product.name}`.toLowerCase().includes(query))
    return matches.sort((a, b) => (sort.ascending ? 1 : -1) * compare(a, b, sort.key))
  }, [products, search, sort])
  const pages = Math.max(1, Math.ceil(visible.length / pageSize))
  const current = Math.min(page, pages - 1)

  function changeSort(key: SortKey) {
    setSort((previous) => ({ key, ascending: previous.key === key ? !previous.ascending : true }))
    setPage(0)
  }

  return (
    <section className="panel" aria-labelledby="products-heading">
      <div className="panel-heading"><div><p className="eyebrow">CATALOG</p><h2 id="products-heading">Products</h2></div>
        <span className="count">{search.trim() ? `${visible.length} of ${products.length}` : products.length}</span></div>
      <div className="toolbar">
        <label className="search-label">Search products
          <input value={search} onChange={(event) => { setSearch(event.target.value); setPage(0) }} placeholder="Name or product ID" />
        </label>
        <label className="checkbox"><input type="checkbox" onChange={(event) => { onIncludeWithdrawn(event.target.checked); setPage(0) }} /> Include withdrawn</label>
      </div>
      {loading && !products.length ? <p className="empty">Loading products…</p>
        : !visible.length ? <p className="empty">No products match this view.</p>
        : <>
          <ProductTable products={visible.slice(current * pageSize, (current + 1) * pageSize)} sort={sort} onSort={changeSort} />
          <Pagination page={current} pages={pages} total={visible.length} onPage={setPage} />
        </>}
    </section>
  )
}

function ProductTable({ products, sort, onSort }: { products: Product[]; sort: Sort; onSort: (key: SortKey) => void }) {
  return <div className="table-scroll"><table><thead><tr>{columns.map((column) =>
    <th key={column.key} aria-sort={sort.key === column.key ? (sort.ascending ? 'ascending' : 'descending') : 'none'}>
      <button type="button" className="sort" onClick={() => onSort(column.key)}>{column.label}
        <span aria-hidden="true">{sort.key === column.key ? (sort.ascending ? ' ↑' : ' ↓') : ''}</span></button></th>)}
  </tr></thead>
    <tbody>{products.map((product) => <tr key={product.id}><td>{product.name}</td><td>{product.id}</td>
      <td>{formatPrice(product.price_cents)}</td>
      <td>{product.revision}</td><td><span className={`badge ${product.withdrawn ? 'muted' : ''}`}>{product.withdrawn ? 'Withdrawn' : 'Active'}</span></td></tr>)}</tbody>
  </table></div>
}

function Pagination({ page, pages, total, onPage }: { page: number; pages: number; total: number; onPage: (page: number) => void }) {
  const first = page * pageSize + 1, last = Math.min(total, (page + 1) * pageSize)
  return <nav className="pagination" aria-label="Product pages">
    <span>Showing {first}–{last} of {total}</span>
    {pages > 1 && <div className="actions">
      <button type="button" className="secondary" disabled={page === 0} onClick={() => onPage(page - 1)}>Previous</button>
      <span>Page {page + 1} of {pages}</span>
      <button type="button" className="secondary" disabled={page >= pages - 1} onClick={() => onPage(page + 1)}>Next</button>
    </div>}
  </nav>
}
