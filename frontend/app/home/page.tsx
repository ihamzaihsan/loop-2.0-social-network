'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import './home.css'

export default function Home() {
  const router = useRouter()
  const [username, setUsername] = useState('User')

  useEffect(() => {
   
    const fetchUserData = async () => {
      try {
     
      } catch (error) {
        console.error('Error fetching user data:', error)
      }
    }

    fetchUserData()
  }, [])

  const handleLogout = async () => {
    try {
      const response = await fetch('http://localhost:3000/logout', {
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
          
          <div className="card profile-card">
            <h2 className="card-title">Your Profile</h2>
            <div className="profile-info">
              <div className="avatar"></div>
              <h3>{username}</h3>
              <p>Member since: {new Date().toLocaleDateString()}</p>
            </div>
            <button className="primary-button edit-profile-btn">Edit Profile</button>
          </div>
          
          <div className="card friends-card">
            <h2 className="card-title">Friends</h2>
            <p className="empty-friends">Connect with new friends to see them here.</p>
            <button className="secondary-button find-friends-btn">Find Friends</button>
          </div>
        </div>
      </main>
    </div>
  )
}