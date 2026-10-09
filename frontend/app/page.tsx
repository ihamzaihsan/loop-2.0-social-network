'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { checkSession } from '../utils/session'

export default function RootPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const verifySession = async () => {
      try {
        const hasSession = await checkSession()

        if (hasSession) {
          router.push('/home')
        } else if (hasSession === false) {
          router.push('/login')
        } else {
          setError('We could not check your connection. Please try again in a moment.')
        }
      } catch (error) {
        console.error('Session verification failed:', error)
        setError('We could not check your connection. Please try again in a moment.')
      } finally {
        setLoading(false)
      }
    }

    verifySession()
  }, [router])

  if (loading) {
    return <div>Loading...</div>
  }

  if (error) return <main className="feature-page"><p role="alert">{error}</p><button onClick={() => window.location.reload()}>Try again</button></main>
  return null
}
