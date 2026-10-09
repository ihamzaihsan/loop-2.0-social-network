'use client'

import { useEffect, useRef } from 'react'

// Coalesce bursts, serialize reloads, and recover missed events on reconnect or
// returning to the tab. The latest callback preserves the page's current state.
export function useRealtimeRefresh(resources: string[], refresh: () => void | Promise<void>, enabled = true) {
  const callback = useRef(refresh)
  useEffect(() => { callback.current = refresh })
  const key = resources.join(',')

  useEffect(() => {
    if (!enabled) return
    const watched = new Set(key.split(','))
    let timer: ReturnType<typeof setTimeout> | undefined
    let running = false
    let pending = false
    let disposed = false
    const run = async () => {
      timer = undefined
      if (disposed) return
      if (running) { pending = true; return }
      running = true
      try { await callback.current() }
      catch (error) { console.error('Realtime refresh failed:', error) }
      finally {
        running = false
        if (pending && !disposed) { pending = false; schedule() }
      }
    }
    const schedule = () => {
      if (!disposed && !timer) timer = setTimeout(run, 100)
    }
    const changed = (event: Event) => {
      const resource = (event as CustomEvent<{resource: string}>).detail.resource
      if (resource === 'all' || watched.has(resource)) schedule()
    }
    const visible = () => { if (document.visibilityState === 'visible') schedule() }
    window.addEventListener('realtime_changed', changed)
    window.addEventListener('realtime_reconnected', schedule)
    window.addEventListener('focus', visible)
    document.addEventListener('visibilitychange', visible)
    return () => {
      disposed = true
      if (timer) clearTimeout(timer)
      window.removeEventListener('realtime_changed', changed)
      window.removeEventListener('realtime_reconnected', schedule)
      window.removeEventListener('focus', visible)
      document.removeEventListener('visibilitychange', visible)
    }
  }, [key, enabled])
}
