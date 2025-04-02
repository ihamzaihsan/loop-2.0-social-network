'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { WebSocketClient } from '../webscoket/websocket'

export default function Logout() {
  const router = useRouter()

  useEffect(() => {
    const performLogout = async () => {
      try {
        const response = await fetch('http://localhost:8080/logout', {
          method: 'POST',
          credentials: 'include'
        })

        // Close WebSocket connection before logout
        const client = WebSocketClient.getInstance();
        if (client.socket) {
          client.socket.close(1000, "User logged out");
        }
        
        WebSocketClient.resetInstance();
        
        localStorage.removeItem('sessionToken');
        
        router.push('/')
      } catch (error) {
        console.error('Logout failed:', error)
        router.push('/')
      }
    }

    performLogout()
  }, [router])

  return <div>Logging out...</div>
}
