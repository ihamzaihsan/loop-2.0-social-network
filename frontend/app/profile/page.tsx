'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import '../home/home.css'
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

  useEffect(() => {
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

    fetchProfile()
  }, [router])

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
              <h2 className="card-title">
                {profile.user.firstName} {profile.user.lastName}
              </h2>
              <p className="profile-nickname">@{profile.user.nickname || profile.user.firstName.toLowerCase()}</p>
              
              {profile.user.aboutMe && (
                <div className="profile-about">
                  <h3>About Me</h3>
                  <p>{profile.user.aboutMe}</p>
                </div>
              )}
              {profile.user.isprivate ? (
                <span className="privacy-badge">Private Account</span>
              ) : (
              <span className="privacy-badge">Public Account</span>
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
              {profile.posts.length > 0 ? (
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
