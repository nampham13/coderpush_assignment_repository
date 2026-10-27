import { useCallback, useEffect, useState } from 'react'
import { loadDashboard } from './api'
import type { Product, Update } from './types'

export function useDashboard(includeWithdrawn: boolean) {
  const [products, setProducts] = useState<Product[]>([])
  const [updates, setUpdates] = useState<Update[]>([])
  const [online, setOnline] = useState(false)
  const [loading, setLoading] = useState(true)
  const [refreshError, setRefreshError] = useState('')

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const data = await loadDashboard(includeWithdrawn)
      setProducts(data.products)
      setUpdates(data.updates)
      setOnline(data.online)
      setRefreshError('')
    } catch (error) {
      setOnline(false)
      setRefreshError(error instanceof Error ? error.message : 'Unable to load catalog.')
    } finally {
      setLoading(false)
    }
  }, [includeWithdrawn])

  useEffect(() => { void refresh() }, [refresh])
  return { products, updates, online, loading, refreshError, refresh }
}
