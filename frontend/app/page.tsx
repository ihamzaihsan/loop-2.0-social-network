'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { checkSession } from '../utils/session'

export default function RootPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const verifySession = async () => {
      try {
        const hasSession = await checkSession()

        if (hasSession) {
          router.push('/home')
        } else {
          router.push('/login')
        }
      } catch (error) {
        console.error('Session verification failed:', error)
        router.push('/login')
      } finally {
        setLoading(false)
      }
    }

    verifySession()
  }, [router])

  if (loading) {
    return <div>Loading...</div>
  }

}
