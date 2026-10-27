import { useState, type FormEvent } from 'react'

export function OutageForm({ disabled, onSubmit }: { disabled: boolean; onSubmit: (message: string) => void }) {
  const [message, setMessage] = useState('')
  function submit(event: FormEvent) {
    event.preventDefault()
    if (!message.trim()) return
    onSubmit(message.trim())
    setMessage('')
  }
  return <form className="outage-form" onSubmit={submit}>
    <label htmlFor="outage-message">Report supplier outage</label>
    <div className="inline-form"><input id="outage-message" required value={message}
      onChange={(event) => setMessage(event.target.value)} placeholder="Briefly describe the issue" />
      <button className="secondary" disabled={disabled}>Report outage</button></div>
  </form>
}
