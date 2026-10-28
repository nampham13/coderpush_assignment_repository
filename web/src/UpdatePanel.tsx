import { useState, type FormEvent } from 'react'
import { BatchForm } from './BatchForm'
import { ConfirmButton } from './ConfirmButton'
import { formatDate, statusLabel } from './format'
import { IssuesPanel } from './IssuesPanel'
import { OutageForm } from './OutageForm'
import type { IncomingProduct, Update } from './types'

type Props = {
  updates: Update[]; selected?: Update; busy: boolean; selectedID: string
  onSelect: (id: string) => void; onCreate: (id: string) => Promise<boolean>
  onResume: () => void; onComplete: () => void; onOutage: (message: string) => void
  onBatch: (id: string, products: IncomingProduct[]) => Promise<boolean>
  onResolve: (id: string) => void
}

export function UpdatePanel(props: Props) {
  const [creating, setCreating] = useState(false)
  const unfinished = props.updates.find((update) => update.status !== 'completed')
  const newestFirst = [...props.updates].sort((a, b) => b.created_at.localeCompare(a.created_at))
  return <section className="panel" aria-labelledby="updates-heading">
    <div className="panel-heading"><div><p className="eyebrow">OPERATIONS</p><h2 id="updates-heading">Catalog updates</h2></div>
      {!creating && <button className="secondary" disabled={props.busy || Boolean(unfinished)}
        title={unfinished ? `Finish update ${unfinished.id} before starting another.` : undefined}
        onClick={() => setCreating(true)}>New update</button>}</div>
    {unfinished && !creating && <p className="hint">Only one update can be open at a time. Finish <strong>{unfinished.id}</strong> to start a new one.</p>}
    {creating && <NewUpdateForm updates={props.updates} busy={props.busy} onCancel={() => setCreating(false)}
      onCreate={async (id) => { if (await props.onCreate(id)) setCreating(false) }} />}
    <label>Select update<select value={props.selectedID} onChange={(event) => props.onSelect(event.target.value)}>
      <option value="">Choose an update</option>{newestFirst.map((update) =>
        <option key={update.id} value={update.id}>{update.id} · {statusLabel(update.status)} · {formatDate(update.created_at)}</option>)}
    </select></label>
    {props.selected ? <UpdateDetails {...props} /> : <p className="empty">Create an update or choose one to manage batches and issues.</p>}
  </section>
}

function suggestedName(updates: Update[]) {
  const base = `update-${new Date().toISOString().slice(0, 10)}`
  const taken = new Set(updates.map((update) => update.id))
  let name = base
  for (let n = 2; taken.has(name); n++) name = `${base}-${n}`
  return name
}

function NewUpdateForm({ updates, busy, onCreate, onCancel }: {
  updates: Update[]; busy: boolean; onCreate: (id: string) => Promise<void>; onCancel: () => void
}) {
  const [name, setName] = useState(() => suggestedName(updates))
  function submit(event: FormEvent) {
    event.preventDefault()
    void onCreate(name.trim())
  }
  return <form className="new-update" onSubmit={submit}>
    <label>Update name<input autoFocus value={name} onChange={(event) => setName(event.target.value)} placeholder="spring-2026" /></label>
    <p className="hint">Use something staff will recognise, such as the supplier and season. Leave blank to generate an ID.</p>
    <div className="actions"><button disabled={busy}>Create update</button>
      <button type="button" className="secondary" onClick={onCancel}>Cancel</button></div>
  </form>
}

function UpdateDetails(props: Props) {
  const update = props.selected!
  const active = update.status === 'in_progress'
  const batches = Object.values(update.batches || {})
  const openIssues = update.issues?.filter((issue) => !issue.resolved).length || 0
  return <div className="update-details">
    <div className="update-meta"><div><small>Update</small><strong>{update.id}</strong>
      <small>Created {formatDate(update.created_at)}</small></div>
      <span className={`badge status-${update.status}`}>{statusLabel(update.status)}</span></div>
    <div className="metrics"><Metric label="Batches" value={batches.length} />
      <Metric label="Applied" value={batches.reduce((sum, batch) => sum + batch.applied, 0)} />
      <Metric label="Open issues" value={openIssues} /></div>
    <div className="actions">{update.status === 'interrupted' && <button disabled={props.busy} onClick={props.onResume}>Resume update</button>}
      {active && <ConfirmButton label="Complete update" question="Has the supplier delivered every batch? Completing cannot be undone."
        confirmLabel="Complete update" disabled={props.busy || openIssues > 0} onConfirm={props.onComplete} />}</div>
    {active && openIssues > 0 && <p className="hint">Resolve {openIssues === 1 ? 'the open issue' : `all ${openIssues} open issues`} below before completing this update.</p>}
    {update.status === 'interrupted' && <p className="hint">This update is paused. Resume it to submit more batches.</p>}
    {active && <OutageForm disabled={props.busy} onSubmit={props.onOutage} />}
    {active && <BatchForm disabled={props.busy} onSubmit={props.onBatch} />}
    <h3>Issues</h3><IssuesPanel issues={update.issues || []} busy={props.busy} onResolve={props.onResolve} />
    <ChangeList update={update} />
  </div>
}

function Metric({ label, value }: { label: string; value: number }) {
  return <div className="metric"><span>{label}</span><strong>{value}</strong></div>
}

function ChangeList({ update }: { update: Update }) {
  if (!update.changes?.length) return <><h3>Recent changes</h3><p className="empty">No product changes yet.</p></>
  return <><h3>Recent changes</h3><ul className="change-list">{update.changes.map((change, index) =>
    <li key={`${change.product_id}-${index}`}><strong>{change.product_id}</strong><span>{statusLabel(change.kind)}</span><span>{change.after.name || 'Withdrawn'}</span></li>)}</ul></>
}
