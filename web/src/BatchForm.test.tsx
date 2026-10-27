import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { BatchForm } from './BatchForm'

describe('BatchForm', () => {
  it('submits a valid product batch', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    render(<BatchForm disabled={false} onSubmit={onSubmit} />)
    fireEvent.change(screen.getByLabelText('Batch ID'), { target: { value: 'batch-1' } })
    fireEvent.click(screen.getByRole('button', { name: 'Submit batch' }))
    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith('batch-1', [
      { id: 'SKU-101', name: 'Example product', description: 'Optional', price_cents: 1299, revision: 1 },
    ]))
  })

  it('rejects malformed JSON without calling the API', async () => {
    const onSubmit = vi.fn()
    render(<BatchForm disabled={false} onSubmit={onSubmit} />)
    fireEvent.change(screen.getByLabelText('Batch ID'), { target: { value: 'batch-1' } })
    fireEvent.change(screen.getByLabelText('Products (JSON array)'), { target: { value: '{bad' } })
    fireEvent.click(screen.getByRole('button', { name: 'Submit batch' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('JSON')
    expect(onSubmit).not.toHaveBeenCalled()
  })
})
