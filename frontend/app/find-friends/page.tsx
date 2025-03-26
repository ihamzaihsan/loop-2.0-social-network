'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import './find-friends.css'
import Sidebar from '../../components/Sidebar'

interface User {
  id: number
  firstName: string
  lastName: string
  email: string
  nickname?: string
  following?: boolean
}

export default function FindFriends() {
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentUserId, setCurrentUserId] = useState<number | null>(null)

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        // First, get the current user profile to know who we are
        const profileResponse = await fetch('http://localhost:8080/profile', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          }
        })

        if (!profileResponse.ok) {
          if (profileResponse.status === 401) {
            router.push('/')
            return
          }
          throw new Error('Failed to fetch your profile')
        }

        const profileData = await profileResponse.json()
        setCurrentUserId(profileData.user.id)

        // Then fetch all users
        const usersResponse = await fetch('http://localhost:8080/users', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          }
        })

        if (!usersResponse.ok) {
          throw new Error('Failed to fetch users')
        }

        const responseData = await usersResponse.json()
        
        // Check if the response has the expected format
        if (!responseData.success || !responseData.users) {
          throw new Error('Unexpected response format from server')
        }
        
        const userData = responseData.users
        
        // Filter out the current user
        const filteredUsers = userData.filter((user: User) => user.id !== profileData.user.id)
        setUsers(filteredUsers.map((user: User) => ({
          ...user,
          following: user.following || false
        })))
      } catch (err: any) {
        console.error('Error fetching users:', err)
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    fetchUsers()
  }, [router])

  const handleFollowToggle = async (userId: number) => {
    try {
      const user = users.find(u => u.id === userId)
      const isFollowing = user?.following

      const endpoint = isFollowing 
        ? 'http://localhost:8080/unfollow' 
        : 'http://localhost:8080/follow'

      const response = await fetch(endpoint, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ followed_id: userId })
      })

      if (!response.ok) {
        throw new Error(`Failed to ${isFollowing ? 'unfollow' : 'follow'} user`)
      }

      // Update the users list to reflect the change
      setUsers(prevUsers => 
        prevUsers.map(user => 
          user.id === userId 
            ? { ...user, following: !isFollowing } 
            : user
        )
      )
    } catch (err: any) {
      console.error('Error toggling follow status:', err)
      setError(err.message)
    }
  }

  const navigateToProfile = (userId: number) => {
    router.push(`/profile/${userId}`)
  }

  if (loading) return <div className="find-friends-page">Loading users...</div>
  if (error) return <div className="find-friends-page">Error: {error}</div>

  return (
    <div className="find-friends-page">
      <Sidebar activePage="find-friends" />
      
      <main className="users-container">
        <h1 className="page-title">Find Friends</h1>
        
        {users.length === 0 ? (
          <p className="no-users-message">No other users found.</p>
        ) : (
          <ul className="users-list">
            {users.map(user => (
              <li key={user.id} className="user-card">
                <div 
                  className="user-info"
                  onClick={() => navigateToProfile(user.id)}
                >
                  <div className="user-avatar">
                    {(user.nickname || user.firstName).charAt(0).toUpperCase()}
                  </div>
                  <div className="user-details">
                    <h3 className="user-name">
                      {user.nickname || `${user.firstName} ${user.lastName}`}
                    </h3>
                  </div>
                </div>
                <button 
                  className={`follow-button ${user.following ? 'following' : ''}`}
                  onClick={() => handleFollowToggle(user.id)}
                >
                  {user.following ? 'Unfollow' : 'Follow'}
                </button>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  )
}