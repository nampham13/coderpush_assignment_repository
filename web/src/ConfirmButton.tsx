import { useState } from 'react'

// A button that asks for an inline confirmation before running a consequential action.
export function ConfirmButton({ label, question, confirmLabel, disabled, danger, canAsk = () => true, onConfirm }: {
  label: string; question: string; confirmLabel: string; disabled?: boolean; danger?: boolean
  canAsk?: () => boolean; onConfirm: () => void
}) {
  const [asking, setAsking] = useState(false)
  if (!asking) return <button type="button" className="secondary" disabled={disabled}
    onClick={() => canAsk() && setAsking(true)}>{label}</button>
  return <div className="confirm" role="group" aria-label={question}>
    <span>{question}</span>
    <div className="confirm-actions">
      <button type="button" className={danger ? 'danger' : ''} disabled={disabled}
        onClick={() => { setAsking(false); onConfirm() }}>{confirmLabel}</button>
      <button type="button" className="secondary" onClick={() => setAsking(false)}>Cancel</button>
    </div>
  </div>
}
