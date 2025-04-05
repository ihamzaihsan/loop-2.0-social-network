'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import '../home/home.css'
import './profile.css'
import Sidebar from '../../components/Sidebar'

interface User {
  id: number
  firstName: string
  lastName: string
  nickname: string
  aboutMe: string
  email: string
  avatar: string
  isprivate: boolean 
  createdAt: string
}

interface Post {
  id: number
  content: string
  createdAt: string
}

interface ProfileData {
  user: User
  posts: Post[]
  postsCount: number
  followersCount: number
  followingCount: number
  followers: User[]
  following: User[]
}

export default function Profile() {
  const router = useRouter()
  const [profile, setProfile] = useState<ProfileData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isUpdating, setIsUpdating] = useState(false)

  useEffect(() => {
    fetchProfile()
  }, [router])

  const fetchProfile = async () => {
    try {
      const response = await fetch('http://localhost:8080/profile', {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/')
          return
        }
        throw new Error('Failed to fetch profile data')
      }

      const data = await response.json()
      setProfile(data)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const togglePrivacy = async () => {
    if (!profile) return
    
    setIsUpdating(true)
    
    try {
      const response = await fetch('http://localhost:8080/profile/privacy', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          isPrivate: !profile.user.isprivate
        }),
        credentials: 'include'
      })
      
      if (!response.ok) {
        throw new Error('Failed to update privacy setting')
      }
      
      // Update the local state to reflect the change
      setProfile(prevProfile => {
        if (!prevProfile) return null
        
        return {
          ...prevProfile,
          user: {
            ...prevProfile.user,
            isprivate: !prevProfile.user.isprivate
          }
        }
      })
      
    } catch (err: any) {
      console.error('Error updating privacy:', err)
      setError(err.message)
    } finally {
      setIsUpdating(false)
    }
  }

  const handleBack = () => {
    router.push('/home')
  }

  if (loading) return <div className="profile-page">Loading profile...</div>
  if (error) return <div className="profile-page">Error: {error}</div>

  return (
    <div className="profile-page">
      <Sidebar activePage="profile" />
      
      <main className="main-content">
        {profile && (
          <div className="profile-container">
            <div className="card profile-card">
              <div className="profile-header">
                <div className="profile-avatar">
                  {profile.user.avatar ? (
                    <img 
                      src={`http://localhost:8080${profile.user.avatar}`} 
                      alt={`${profile.user.firstName}'s avatar`} 
                      className="avatar-image"
                    />
                  ) : (
                    <div className="avatar-placeholder">
                      {profile.user.firstName.charAt(0)}
                    </div>
                  )}
                </div>
                
                <div className="profile-info">
                  <h2 className="profile-name">
                    {profile.user.firstName} {profile.user.lastName}
                  </h2>
                  <p className="profile-nickname">@{profile.user.nickname || profile.user.firstName.toLowerCase()}</p>
                  
                  <div className="privacy-controls">
                    {profile.user.isprivate ? (
                      <span className="privacy-badge private">Private Account</span>
                    ) : (
                      <span className="privacy-badge public">Public Account</span>
                    )}
                    
                    <button 
                      className={`privacy-toggle-btn ${isUpdating ? 'updating' : ''}`}
                      onClick={togglePrivacy}
                      disabled={isUpdating}
                    >
                      {isUpdating ? 'Updating...' : profile.user.isprivate ? 'Make Public' : 'Make Private'}
                    </button>
                  </div>
                </div>
              </div>
              
              {profile.user.aboutMe && (
                <div className="profile-about">
                  <h3>About Me</h3>
                  <p>{profile.user.aboutMe}</p>
                </div>
              )}
              
              <div className="profile-stats">
                <div className="stat">
                  <span className="stat-count">{profile.postsCount}</span>
                  <span className="stat-label">Posts</span>
                </div>
                <div className="stat">
                  <span className="stat-count">{profile.followersCount}</span>
                  <span className="stat-label">Followers</span>
                </div>
                <div className="stat">
                  <span className="stat-count">{profile.followingCount}</span>
                  <span className="stat-label">Following</span>
                </div>
              </div>
            </div>
            
            <div className="card posts-card">
              <h2 className="card-title">Posts</h2>
              {profile.posts && profile.posts.length > 0 ? (
                <div className="posts-grid">
                  {profile.posts.map(post => (
                    <div key={post.id} className="post-item">
                      {/* Display post content */}
                      <p>{post.content}</p>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="empty-posts">No posts yet.</p>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
