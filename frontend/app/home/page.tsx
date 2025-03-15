'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import './home.css'

interface User {
  id: number
  firstName: string
  lastName: string
  email: string
  nickname?: string
}

export default function Home() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        // Fetch user data from the backend
        const response = await fetch('http://localhost:8080/profile', {
          method: 'GET',
          credentials: 'include'
        })

        if (!response.ok) {
          if (response.status === 401) {
            // Unauthorized, redirect to login
            router.push('/')
            return
          }
          throw new Error('Failed to fetch user data')
        }

        const data = await response.json()
        
        // Set the username based on the user's first name or nickname
        if (data.user) {
          // Use nickname if available, otherwise use firstName
          setUsername(data.user.nickname || data.user.firstName)
        }
      } catch (error: any) {
        console.error('Error fetching user data:', error)
        setError(error.message)
      } finally {
        setLoading(false)
      }
    }

    fetchUserData()
  }, [router])

  const handleLogout = async () => {
    try {
      const response = await fetch('http://localhost:8080/logout', {
        method: 'POST',
        credentials: 'include'
      })

      if (response.ok) {
        router.push('/')
      } else {
        console.error('Error logging out')
      }
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  const navigateToProfile = () => {
    router.push('/profile')
  }

  if (loading) return <div className="home-page">Loading...</div>
  if (error) return <div className="home-page">Error: {error}</div>

  return (
    <div className="home-page">
      <header className="header">
        <div className="header-content">
          <h1 className="site-title">Social Network</h1>
          <div className="user-actions">
            <span className="welcome-message">Welcome, {username}!</span>
            <button onClick={handleLogout} className="logout-button">
              Logout
            </button>
          </div>
        </div>
      </header>

      <main className="main-content">
        <div className="dashboard">
          <div className="card feed-card">
            <h2 className="card-title">Your Feed</h2>
            <p className="empty-feed">No posts yet. Start connecting with friends!</p>
          </div>
          
          <div className="card friends-card">
            <h2 className="card-title">Friends</h2>
            <p className="empty-friends">Connect with new friends to see them here.</p>
            <button className="secondary-button find-friends-btn">Find Friends</button>
          </div>
        </div>
        
        {/* Circular profile button */}
        <button 
          className="profile-circle-button" 
          onClick={navigateToProfile}
          aria-label="Go to profile"
        >
          {username ? username.charAt(0).toUpperCase() : ''}
        </button>
      </main>
    </div>
  )
}