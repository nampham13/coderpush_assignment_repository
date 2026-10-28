import { describe, expect, it } from 'vitest'
import { parseCSV, toProducts } from './productRows'

describe('parseCSV', () => {
  it('reads quoted cells, withdrawn flags, and price_cents', () => {
    const rows = parseCSV('ID,Name,Description,Price_Cents,Revision,Withdrawn\r\nTEA-1,"Tea, black","Says ""hi""",1299,3,false\nMUG-1,,,,2,yes\n')
    expect(rows).toEqual([
      { id: 'TEA-1', name: 'Tea, black', description: 'Says "hi"', price: '12.99', revision: '3', withdrawn: false },
      { id: 'MUG-1', name: '', description: '', price: '', revision: '2', withdrawn: true },
    ])
  })

  it('names the missing required columns', () => {
    expect(() => parseCSV('name,price\nTea,1')).toThrow('must include: id, revision')
  })
})

describe('toProducts', () => {
  it('rejects duplicate IDs within a batch', () => {
    const row = { id: 'TEA-1', name: 'Tea', description: '', price: '1', revision: '1', withdrawn: false }
    expect(toProducts([row, row]).errors).toEqual(['Row 2: product ID TEA-1 is already used in another row.'])
  })
})
