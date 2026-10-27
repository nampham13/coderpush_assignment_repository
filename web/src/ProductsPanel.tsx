import { useMemo, useState } from 'react'
import type { Product } from './types'

type Props = { products: Product[]; loading: boolean; onIncludeWithdrawn: (value: boolean) => void }

export function ProductsPanel({ products, loading, onIncludeWithdrawn }: Props) {
  const [search, setSearch] = useState('')
  const visible = useMemo(() => products.filter((product) =>
    `${product.id} ${product.name}`.toLowerCase().includes(search.toLowerCase())), [products, search])

  return (
    <section className="panel" aria-labelledby="products-heading">
      <div className="panel-heading"><div><p className="eyebrow">CATALOG</p><h2 id="products-heading">Products</h2></div>
        <span className="count">{products.length}</span></div>
      <div className="toolbar">
        <label className="search-label">Search products
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Name or product ID" />
        </label>
        <label className="checkbox"><input type="checkbox" onChange={(event) => onIncludeWithdrawn(event.target.checked)} /> Include withdrawn</label>
      </div>
      <ProductTable products={visible} loading={loading} />
    </section>
  )
}

function ProductTable({ products, loading }: { products: Product[]; loading: boolean }) {
  if (loading) return <p className="empty">Loading products…</p>
  if (!products.length) return <p className="empty">No products match this view.</p>
  return <div className="table-scroll"><table><thead><tr><th>Product</th><th>ID</th><th>Price</th><th>Revision</th><th>Status</th></tr></thead>
    <tbody>{products.map((product) => <tr key={product.id}><td>{product.name}</td><td>{product.id}</td>
      <td>{(product.price_cents / 100).toLocaleString(undefined, { style: 'currency', currency: 'USD' })}</td>
      <td>{product.revision}</td><td><span className={`badge ${product.withdrawn ? 'muted' : ''}`}>{product.withdrawn ? 'Withdrawn' : 'Active'}</span></td></tr>)}</tbody>
  </table></div>
}
