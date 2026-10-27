import { BatchForm } from './BatchForm'
import { IssuesPanel } from './IssuesPanel'
import { OutageForm } from './OutageForm'
import type { Update } from './types'

type Props = {
  updates: Update[]; selected?: Update; busy: boolean; selectedID: string
  onSelect: (id: string) => void; onCreate: (id: string) => void
  onResume: () => void; onComplete: () => void; onOutage: (message: string) => void
  onBatch: (id: string, products: unknown[]) => Promise<void>
  onResolve: (id: string) => void
}

export function UpdatePanel(props: Props) {
  const active = props.selected?.status === 'in_progress'
  return <section className="panel" aria-labelledby="updates-heading">
    <div className="panel-heading"><div><p className="eyebrow">OPERATIONS</p><h2 id="updates-heading">Catalog updates</h2></div>
      <button className="secondary" disabled={props.busy} onClick={() => props.onCreate('')}>New update</button></div>
    <label>Select update<select value={props.selectedID} onChange={(event) => props.onSelect(event.target.value)}>
      <option value="">Choose an update</option>{props.updates.map((update) =>
        <option key={update.id} value={update.id}>{update.id} · {update.status}</option>)}
    </select></label>
    {props.selected ? <UpdateDetails {...props} active={active} /> : <p className="empty">Create an update or choose one to manage batches and issues.</p>}
  </section>
}

function UpdateDetails(props: Props & { active: boolean }) {
  const update = props.selected!
  const batches = Object.values(update.batches || {})
  return <div className="update-details">
    <div className="update-meta"><div><small>Update ID</small><strong>{update.id}</strong></div>
      <span className="badge">{update.status.split('_').join(' ')}</span></div>
    <div className="metrics"><Metric label="Batches" value={batches.length} />
      <Metric label="Applied" value={batches.reduce((sum, batch) => sum + batch.applied, 0)} />
      <Metric label="Issues" value={update.issues?.filter((issue) => !issue.resolved).length || 0} /></div>
    <div className="actions">{update.status === 'interrupted' && <button disabled={props.busy} onClick={props.onResume}>Resume update</button>}
      {props.active && <button className="secondary" disabled={props.busy} onClick={props.onComplete}>Complete update</button>}</div>
    {props.active && <OutageForm disabled={props.busy} onSubmit={props.onOutage} />}
    {props.active && <BatchForm disabled={props.busy} onSubmit={props.onBatch} />}
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
    <li key={`${change.product_id}-${index}`}><strong>{change.product_id}</strong><span>{change.kind}</span><span>{change.after.name || 'Withdrawn'}</span></li>)}</ul></>
}
