import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ProductsPanel } from './ProductsPanel'
import type { Product } from './types'

const product = (n: number): Product => ({ id: `SKU-${n}`, name: `Item ${n}`, price_cents: n * 100, revision: 1, withdrawn: false })

describe('ProductsPanel', () => {
  it('filters products by name or ID and exposes withdrawn option', () => {
    const onIncludeWithdrawn = vi.fn()
    render(<ProductsPanel products={[
      { id: 'TEA-1', name: 'Breakfast Tea', price_cents: 500, revision: 1, withdrawn: false },
      { id: 'MUG-1', name: 'Stoneware Mug', price_cents: 900, revision: 2, withdrawn: true },
    ]} loading={false} onIncludeWithdrawn={onIncludeWithdrawn} />)

    fireEvent.change(screen.getByLabelText('Search products'), { target: { value: 'tea-1' } })
    expect(screen.getByText('Breakfast Tea')).toBeInTheDocument()
    expect(screen.queryByText('Stoneware Mug')).not.toBeInTheDocument()
    expect(screen.getByText('1 of 2')).toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('Include withdrawn'))
    expect(onIncludeWithdrawn).toHaveBeenCalledWith(true)
  })

  it('pages through long catalogs and sorts by column', () => {
    render(<ProductsPanel products={Array.from({ length: 45 }, (_, i) => product(i + 1))} loading={false} onIncludeWithdrawn={vi.fn()} />)
    const firstName = () => within(screen.getAllByRole('row')[1]).getAllByRole('cell')[0].textContent
    expect(screen.getAllByRole('row')).toHaveLength(21)
    expect(screen.getByText('Showing 1–20 of 45')).toBeInTheDocument()
    expect(firstName()).toBe('Item 1')

    fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(screen.getByText('Page 2 of 3')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Price/ }))
    fireEvent.click(screen.getByRole('button', { name: /Price/ }))
    expect(screen.getByRole('columnheader', { name: /Price/ })).toHaveAttribute('aria-sort', 'descending')
    expect(screen.getByText('Page 1 of 3')).toBeInTheDocument()
    expect(firstName()).toBe('Item 45')
  })
})
