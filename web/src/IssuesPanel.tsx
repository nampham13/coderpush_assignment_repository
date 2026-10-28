import { statusLabel } from './format'
import type { Issue } from './types'

export function IssuesPanel({ issues, busy, onResolve }: {
  issues: Issue[]; busy: boolean; onResolve: (id: string) => void
}) {
  if (!issues.length) return <p className="empty">No issues for this update.</p>
  return <ul className="issue-list">{issues.map((issue) => <li key={issue.id}>
    <div><strong>{statusLabel(issue.code)}</strong><p>{issue.message}</p>
      {issue.product_id && <small>Product: {issue.product_id}</small>}</div>
    {issue.resolved ? <span className="badge">Resolved</span> :
      <button className="secondary" disabled={busy} onClick={() => onResolve(issue.id)}>Resolve</button>}
  </li>)}</ul>
}
