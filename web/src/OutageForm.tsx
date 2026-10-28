import { useState } from 'react'
import { ConfirmButton } from './ConfirmButton'

export function OutageForm({ disabled, onSubmit }: { disabled: boolean; onSubmit: (message: string) => void }) {
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  function hasMessage() {
    if (message.trim()) { setError(''); return true }
    setError('Describe the outage before reporting it.')
    return false
  }
  return <div className="outage-form">
    <label htmlFor="outage-message">Report supplier outage</label>
    <div className="inline-form"><input id="outage-message" value={message}
      onChange={(event) => setMessage(event.target.value)} placeholder="Briefly describe the issue" />
      <ConfirmButton label="Report outage" question="Pause this update and record an outage issue?"
        confirmLabel="Report outage" danger disabled={disabled} canAsk={hasMessage}
        onConfirm={() => { onSubmit(message.trim()); setMessage('') }} /></div>
    {error && <p className="inline-error" role="alert">{error}</p>}
  </div>
}
