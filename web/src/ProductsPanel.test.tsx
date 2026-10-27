import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ProductsPanel } from './ProductsPanel'

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
    fireEvent.click(screen.getByLabelText('Include withdrawn'))
    expect(onIncludeWithdrawn).toHaveBeenCalledWith(true)
  })
})
