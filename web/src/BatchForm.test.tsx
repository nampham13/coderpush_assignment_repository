import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { BatchForm } from './BatchForm'

describe('BatchForm', () => {
  it('submits products entered row by row with prices in dollars', async () => {
    const onSubmit = vi.fn().mockResolvedValue(true)
    render(<BatchForm disabled={false} onSubmit={onSubmit} />)
    fireEvent.change(screen.getByLabelText('Batch ID'), { target: { value: 'batch-1' } })
    fireEvent.change(screen.getByLabelText('Product 1 ID'), { target: { value: 'TEA-1' } })
    fireEvent.change(screen.getByLabelText('Product 1 revision'), { target: { value: '3' } })
    fireEvent.change(screen.getByLabelText('Product 1 name'), { target: { value: 'Breakfast Tea' } })
    fireEvent.change(screen.getByLabelText('Product 1 price'), { target: { value: '12.99' } })
    fireEvent.click(screen.getByRole('button', { name: 'Add product' }))
    fireEvent.change(screen.getByLabelText('Product 2 ID'), { target: { value: 'MUG-1' } })
    fireEvent.change(screen.getByLabelText('Product 2 revision'), { target: { value: '2' } })
    fireEvent.click(screen.getByLabelText('Withdraw product 2'))
    fireEvent.click(screen.getByRole('button', { name: /Submit batch/ }))
    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith('batch-1', [
      { id: 'TEA-1', name: 'Breakfast Tea', description: '', price_cents: 1299, revision: 3 },
      { id: 'MUG-1', revision: 2, withdrawn: true },
    ]))
    await waitFor(() => expect(screen.getByLabelText('Batch ID')).toHaveValue(''))
  })

  it('starts empty and fills sample rows only on request', () => {
    render(<BatchForm disabled={false} onSubmit={vi.fn()} />)
    expect(screen.getByLabelText('Product 1 ID')).toHaveValue('')
    fireEvent.click(screen.getByRole('button', { name: 'Use example' }))
    expect(screen.getByLabelText('Product 1 ID')).toHaveValue('SKU-101')
  })

  it('explains invalid rows in plain language without calling the API', async () => {
    const onSubmit = vi.fn()
    render(<BatchForm disabled={false} onSubmit={onSubmit} />)
    fireEvent.change(screen.getByLabelText('Product 1 price'), { target: { value: 'abc' } })
    fireEvent.click(screen.getByRole('button', { name: /Submit batch/ }))
    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Enter a batch ID.')
    expect(alert).toHaveTextContent('Row 1: enter a product ID')
    expect(alert).toHaveTextContent('price must be a number')
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('keeps the entered rows when the API rejects the batch', async () => {
    const onSubmit = vi.fn().mockResolvedValue(false)
    render(<BatchForm disabled={false} onSubmit={onSubmit} />)
    fireEvent.click(screen.getByRole('button', { name: 'Use example' }))
    fireEvent.change(screen.getByLabelText('Batch ID'), { target: { value: 'batch-1' } })
    fireEvent.click(screen.getByRole('button', { name: /Submit batch/ }))
    await waitFor(() => expect(onSubmit).toHaveBeenCalled())
    expect(screen.getByLabelText('Batch ID')).toHaveValue('batch-1')
    expect(screen.getByLabelText('Product 1 ID')).toHaveValue('SKU-101')
  })
})
