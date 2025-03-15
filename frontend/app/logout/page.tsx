'use client'

import { useRouter } from 'next/navigation'

export default function useLogout() {
  const router = useRouter()

  const logout = async () => {
    const response = await fetch('http://localhost:8080/logout', {
      method: 'POST',
      credentials: 'include'
    })

    if (response.ok) {
      router.push('/') 
    } else {
      console.log('Error logging out')
    }
  }

  return logout
}
