'use client'
import { API } from "../utils/api";

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { WebSocketClient } from '@/app/webscoket/websocket'
import styles from './RealtimeAlerts.module.css'

interface Alert {
  id: string
  title: string
  preview: string
  href: string
  expiresAt: number
  notificationId?: number
}

interface IncomingMessage {
  id: number
  sender_id: number
  receiver_id: number
  sender?: string
  content: string
  type?: string
}

interface IncomingNotification {
  id: number
  type: string
  content?: string
  from_user_id?: number
  related_id?: number
  group_id?: number
}

export default function RealtimeAlerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [paused, setPaused] = useState(false)

  useEffect(() => {
    // Some notification producers emit the same saved notification twice.
    const seen = new Set<string>()
    const show = (alert: Omit<Alert, 'expiresAt'>) => {
      if (seen.has(alert.id)) return
      seen.add(alert.id)
      if (seen.size > 100) seen.delete(seen.values().next().value!)
      setAlerts(previous => [...previous, { ...alert, expiresAt: Date.now() + 6000 }].slice(-3))
    }
    const message = (event: Event) => {
      const content = (event as CustomEvent<IncomingMessage>).detail
      const client = WebSocketClient.getInstance()
      if (content.receiver_id !== client.currentUserId || content.sender_id === client.currentUserId) return
      if (client.currentChatUser === content.sender_id && document.visibilityState === 'visible' && document.hasFocus()) return
      show({
        id: `message:${content.id}`,
        title: `Message from ${content.sender || 'someone in your circle'}`,
        preview: content.type === 'image' || /\.(jpe?g|png|gif|webp)$/i.test(content.content) ? 'Sent an image' : content.content,
        href: `/chat?user=${content.sender_id}`,
      })
    }
    const notification = (event: Event) => {
      const content = (event as CustomEvent<IncomingNotification>).detail
      const groupId = content.group_id || content.related_id
      let href = '/find-friends?tab=requests'
      if (content.type === 'group_invitation' || content.type === 'group_join_request') href = '/groups/invitations'
      else if (content.type === 'group_event' && groupId) href = `/groups/${groupId}`
      else if (content.from_user_id && content.type !== 'follow_request') href = `/profile/${content.from_user_id}`
      show({ id: `notification:${content.id}`, notificationId: content.id, title: 'New notification', preview: content.content || 'You have a new update from your circle.', href })
    }
    const clear = () => { seen.clear(); setAlerts([]); setPaused(false) }
    window.addEventListener('private_message', message)
    window.addEventListener('notification', notification)
    window.addEventListener('session_ended', clear)
    return () => {
      window.removeEventListener('private_message', message)
      window.removeEventListener('notification', notification)
      window.removeEventListener('session_ended', clear)
    }
  }, [])

  useEffect(() => {
    if (!alerts.length || paused) return
    const timer = setTimeout(() => setAlerts(previous => previous.filter(alert => alert.expiresAt > Date.now())), Math.max(0, Math.min(...alerts.map(alert => alert.expiresAt)) - Date.now()))
    return () => clearTimeout(timer)
  }, [alerts, paused])

  const dismiss = (id: string) => setAlerts(previous => previous.filter(alert => alert.id !== id))
  const open = (alert: Alert) => {
    dismiss(alert.id)
    if (alert.notificationId) {
      fetch(`${API}/notifications/read?id=${alert.notificationId}`, { method: 'POST', credentials: 'include' })
        .catch(error => console.error('Failed to mark notification read:', error))
    }
  }

  return (
    <section className={styles.alerts} aria-label="Incoming alerts" aria-live="polite" aria-relevant="additions" onMouseEnter={() => setPaused(true)} onMouseLeave={() => setPaused(false)} onFocusCapture={() => setPaused(true)} onBlurCapture={event => { if (!event.currentTarget.contains(event.relatedTarget)) setPaused(false) }}>
      {alerts.map(alert => (
        <div key={alert.id} className={styles.toast}>
          <Link href={alert.href} className={styles.content} onClick={() => open(alert)}>
            <strong>{alert.title}</strong>
            <span>{alert.preview}</span>
          </Link>
          <button type="button" className={styles.dismiss} aria-label="Dismiss notification" onClick={() => dismiss(alert.id)}>×</button>
        </div>
      ))}
    </section>
  )
}
