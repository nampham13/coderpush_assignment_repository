import { useMemo, useState } from 'react'
import { completeUpdate, createUpdate, reportOutage, resolveIssue, resumeUpdate, runDemo, submitBatch } from './api'
import { ProductsPanel } from './ProductsPanel'
import { UpdatePanel } from './UpdatePanel'
import { useDashboard } from './useDashboard'

export default function App() {
  const [includeWithdrawn, setIncludeWithdrawn] = useState(false)
  const [selectedID, setSelectedID] = useState('')
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const dashboard = useDashboard(includeWithdrawn)
  const selected = useMemo(() => dashboard.updates.find((item) => item.id === selectedID), [dashboard.updates, selectedID])

  async function perform(action: () => Promise<unknown>, message: string) {
    setBusy(true); setError(''); setNotice('')
    try { await action(); await dashboard.refresh(); setNotice(message) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Action failed.') }
    finally { setBusy(false) }
  }

  return <main className="page">
    <header className="topbar"><div><p className="eyebrow">SUPPLIER OPERATIONS</p><h1>Catalog Desk</h1></div>
      <div className="top-actions"><span className={`connection ${dashboard.online ? 'online' : ''}`}><i />{dashboard.online ? 'API connected' : 'API offline'}</span>
        <button className="secondary" disabled={busy || dashboard.loading} onClick={() => void dashboard.refresh()}>Refresh</button></div></header>
    <p className="intro">Review products, apply supplier batches, and resolve catalog issues.</p>
    {error && <div className="alert error" role="alert">{error}</div>}
    {notice && <div className="alert success" role="status">{notice}</div>}
    {dashboard.refreshError && <div className="alert error" role="alert">{dashboard.refreshError}</div>}
    <div className="summary"><Summary label="Products" value={dashboard.products.length} />
      <Summary label="Updates" value={dashboard.updates.length} />
      <Summary label="Open issues" value={dashboard.updates.reduce((total, item) => total + (item.issues || []).filter((issue) => !issue.resolved).length, 0)} />
      <button disabled={busy || !dashboard.online} onClick={() => void perform(async () => {
        const update = await runDemo(); setSelectedID(update.id)
      }, 'Demo catalog update finished.')}>Run demo</button></div>
    <div className="content-grid">
      <ProductsPanel products={dashboard.products} loading={dashboard.loading} onIncludeWithdrawn={setIncludeWithdrawn} />
      <UpdatePanel updates={dashboard.updates} selected={selected} selectedID={selectedID} busy={busy}
        onSelect={setSelectedID} onCreate={(id) => void perform(async () => {
          const update = await createUpdate(id); setSelectedID(update.id)
        }, 'Update created.')}
        onResume={() => selected && void perform(() => resumeUpdate(selected.id), 'Update resumed.')}
        onComplete={() => selected && void perform(() => completeUpdate(selected.id), 'Update completed.')}
        onOutage={(message) => selected && void perform(() => reportOutage(selected.id, message), 'Outage recorded; this update was paused.')}
        onBatch={(id, products) => perform(() => submitBatch(selected!.id, { id, products }), 'Batch saved.')}
        onResolve={(id) => selected && void perform(() => resolveIssue(selected.id, id), 'Issue resolved.')} />
    </div>
  </main>
}

function Summary({ label, value }: { label: string; value: number }) {
  return <div className="summary-item"><span>{label}</span><strong>{value}</strong></div>
}
