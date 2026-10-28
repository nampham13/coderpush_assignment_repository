import { useMemo, useState } from 'react'
import { completeUpdate, createUpdate, reportOutage, resolveIssue, resumeUpdate, runDemo, submitBatch } from './api'
import { ProductsPanel } from './ProductsPanel'
import { ToastRegion, useToasts, type ToastTone } from './Toasts'
import { UpdatePanel } from './UpdatePanel'
import type { IncomingProduct, Update } from './types'
import { useDashboard } from './useDashboard'

export default function App() {
  const [includeWithdrawn, setIncludeWithdrawn] = useState(false)
  const [selectedID, setSelectedID] = useState('')
  const [busy, setBusy] = useState(false)
  const toasts = useToasts()
  const dashboard = useDashboard(includeWithdrawn)
  const selected = useMemo(() => dashboard.updates.find((item) => item.id === selectedID), [dashboard.updates, selectedID])
  const activeProducts = dashboard.products.filter((product) => !product.withdrawn).length

  async function perform(action: () => Promise<unknown>, message: string, tone: ToastTone = 'success') {
    setBusy(true); toasts.clear()
    try { await action(); await dashboard.refresh(); if (message) toasts.show(tone, message); return true }
    catch (reason) { toasts.show('error', reason instanceof Error ? reason.message : 'Action failed.'); return false }
    finally { setBusy(false) }
  }

  async function submitBatchAndReport(updateID: string, batchID: string, products: IncomingProduct[]) {
    let summary: Update['batches'][string] | undefined
    const saved = await perform(async () => {
      summary = (await submitBatch(updateID, { id: batchID, products })).batches?.[batchID]
    }, '')
    if (saved && summary) {
      toasts.show(summary.rejected ? 'warning' : 'success', `Batch ${batchID} saved: ${summary.applied} applied, ${summary.unchanged} unchanged, ${summary.stale} stale, ${summary.rejected} rejected.` +
        (summary.rejected ? ' See the issues list for details.' : ''))
    }
    return saved
  }

  async function refresh() {
    toasts.clear()
    await dashboard.refresh()
  }

  return <main className="page">
    <header className="topbar"><div><p className="eyebrow">SUPPLIER OPERATIONS</p><h1>Catalog Desk</h1></div>
      <div className="top-actions"><span className={`connection ${dashboard.online ? 'online' : ''}`}><i />{dashboard.online ? 'API connected' : 'API offline'}</span>
        <button className="secondary" disabled={busy || dashboard.loading} aria-busy={dashboard.loading}
          onClick={() => void refresh()}>{dashboard.loading ? 'Refreshing…' : 'Refresh'}</button></div></header>
    <p className="intro">Review products, apply supplier batches, and resolve catalog issues.</p>
    {dashboard.refreshError && <div className="alert error" role="alert">Could not load the dashboard: {dashboard.refreshError}</div>}
    <div className="summary"><Summary label="Active products" value={activeProducts} />
      <Summary label="Updates" value={dashboard.updates.length} />
      <Summary label="Open issues" value={dashboard.updates.reduce((total, item) => total + (item.issues || []).filter((issue) => !issue.resolved).length, 0)} />
      <button disabled={busy || !dashboard.online} onClick={() => void perform(async () => {
        const update = await runDemo(); setSelectedID(update.id)
      }, 'Demo catalog update finished.')}>Run demo</button></div>
    <div className="content-grid">
      <ProductsPanel products={dashboard.products} loading={dashboard.loading} onIncludeWithdrawn={setIncludeWithdrawn} />
      <UpdatePanel updates={dashboard.updates} selected={selected} selectedID={selectedID} busy={busy}
        onSelect={setSelectedID} onCreate={(id) => perform(async () => {
          const update = await createUpdate(id); setSelectedID(update.id)
        }, 'Update created.')}
        onResume={() => selected && void perform(() => resumeUpdate(selected.id), 'Update resumed.')}
        onComplete={() => selected && void perform(() => completeUpdate(selected.id), 'Update completed.')}
        onOutage={(message) => selected && void perform(() => reportOutage(selected.id, message), 'Outage recorded; this update is paused until you resume it.', 'warning')}
        onBatch={(id, products) => submitBatchAndReport(selected!.id, id, products)}
        onResolve={(id) => selected && void perform(() => resolveIssue(selected.id, id), 'Issue resolved.')} />
    </div>
    <ToastRegion toasts={toasts.toasts} onDismiss={toasts.dismiss} />
  </main>
}

function Summary({ label, value }: { label: string; value: number }) {
  return <div className="summary-item"><span>{label}</span><strong>{value}</strong></div>
}
