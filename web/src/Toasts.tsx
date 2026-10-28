import { useCallback, useEffect, useRef, useState } from 'react'

export type ToastTone = 'success' | 'warning' | 'error'
export type Toast = { id: number; tone: ToastTone; text: string }

// Successes and warnings fade out on their own; errors stay until dismissed.
const autoDismissMs = 6000

export function useToasts() {
  const [toasts, setToasts] = useState<Toast[]>([])
  const nextID = useRef(1)
  const dismiss = useCallback((id: number) => setToasts((items) => items.filter((item) => item.id !== id)), [])
  const clear = useCallback(() => setToasts([]), [])
  const show = useCallback((tone: ToastTone, text: string) => {
    const id = nextID.current++
    setToasts((items) => [...items.slice(-2), { id, tone, text }])
  }, [])
  return { toasts, show, dismiss, clear }
}

export function ToastRegion({ toasts, onDismiss }: { toasts: Toast[]; onDismiss: (id: number) => void }) {
  return <div className="toast-region" aria-live="polite">
    {toasts.map((toast) => <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />)}
  </div>
}

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: (id: number) => void }) {
  useEffect(() => {
    if (toast.tone === 'error') return
    const timer = window.setTimeout(() => onDismiss(toast.id), autoDismissMs)
    return () => window.clearTimeout(timer)
  }, [toast, onDismiss])
  return <div className={`toast ${toast.tone}`} role={toast.tone === 'error' ? 'alert' : 'status'}>
    <span>{toast.text}</span>
    <button className="toast-close" aria-label="Dismiss message" onClick={() => onDismiss(toast.id)}>×</button>
  </div>
}
