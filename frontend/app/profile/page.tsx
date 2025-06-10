'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import '../home/home.css'
import './profile.css'
import Sidebar from '../../components/Sidebar'

interface User {
  id: number
  firstName?: string
  lastName?: string
  nickname?: string
  aboutMe?: string
  email?: string
  avatar?: string
  isprivate?: boolean 
  createdAt?: string
}

interface Post {
  id: number
  userId: number
  content: string
  image: string
  privacy: string
  createdAt: string
  author: {
    firstName: string
    lastName: string
    nickname?: string
    avatar?: string
  }
  likeCount: number
  isLiked: boolean
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
  const [activeModal, setActiveModal] = useState<'followers' | 'following' | null>(null)
  const [userCache, setUserCache] = useState<Map<number, User>>(new Map())
  const [loadingUserDetails, setLoadingUserDetails] = useState(false)

  useEffect(() => {
    fetchProfile()
  }, [router])

  // When a modal is opened, fetch user details for that list
  useEffect(() => {
    if (profile && activeModal) {
      const userList = activeModal === 'followers' ? profile.followers : profile.following
      fetchUserDetails(userList)
    }
  }, [profile, activeModal])

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
      console.log('Profile data:', data)
      setProfile(data)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  // Fetch user details for a list of users
  const fetchUserDetails = async (users: User[]) => {
    if (!users || users.length === 0) return
    
    setLoadingUserDetails(true)
    
    try {
      // Create a new cache with existing entries
      const newCache = new Map(userCache)
      
      // For each user in the list
      for (const user of users) {
        // Skip if we already have this user in cache
        if (newCache.has(user.id)) continue
        
        try {
          // Fetch user profile from the profile endpoint
          const response = await fetch(`http://localhost:8080/profile/${user.id}`, {
            method: 'GET',
            credentials: 'include'
          })
          
          if (response.ok) {
            const userData = await response.json()
            if (userData && userData.user) {
              newCache.set(user.id, userData.user)
            }
          } else {
            // If we can't get the profile, at least store what we have
            newCache.set(user.id, { id: user.id, avatar: user.avatar })
          }
        } catch (error) {
          console.error(`Error fetching details for user ${user.id}:`, error)
        }
      }
      
      // Update the cache with new entries
      setUserCache(newCache)
    } catch (err) {
      console.error('Error fetching user details:', err)
    } finally {
      setLoadingUserDetails(false)
    }
  }

  // Get user display name from cache
  const getUserDisplayName = (userId: number) => {
    const user = userCache.get(userId)
    
    if (user) {
      if (user.nickname) return user.nickname
      if (user.firstName && user.lastName) return `${user.firstName} ${user.lastName}`
      if (user.firstName) return user.firstName
      if (user.email) return user.email
    }
    
    return `User ${userId}`
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

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  const handleDeletePost = async (postId: number) => {
    if (confirm('Are you sure you want to delete this post?')) {
      try {
        const response = await fetch(`http://localhost:8080/posts?id=${postId}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        
        if (!response.ok) {
          throw new Error('Failed to delete post')
        }
        
        // Remove the deleted post from state
        setProfile(prevProfile => {
          if (!prevProfile) return null
          return {
            ...prevProfile,
            posts: prevProfile.posts.filter(post => post.id !== postId),
            postsCount: prevProfile.postsCount - 1
          }
        })
      } catch (err: any) {
        console.error('Error deleting post:', err)
      }
    }
  }

  const navigateToCreatePost = () => {
    router.push('/create-post')
  }

  const openFollowersModal = () => {
    setActiveModal('followers')
  }

  const openFollowingModal = () => {
    setActiveModal('following')
  }

  const closeModal = () => {
    setActiveModal(null)
  }

  const navigateToProfile = (userId: number) => {
    router.push(`/profile/${userId}`)
    closeModal()
  }

  const fetchFollowers = async (userId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/user/${userId}/connections/followers`, {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        throw new Error('Failed to fetch followers')
      }

      const data = await response.json()
      return data.connections
    } catch (err: any) {
      console.error('Error fetching followers:', err)
      return []
    }
  }

  const fetchFollowing = async (userId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/user/${userId}/connections/following`, {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        throw new Error('Failed to fetch following')
      }

      const data = await response.json()
      return data.connections
    } catch (err: any) {
      console.error('Error fetching following:', err)
      return []
    }
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
                      {profile.user.firstName?.charAt(0) || '?'}
                    </div>
                  )}
                </div>
                
                <div className="profile-info">
                  <h2 className="profile-name">
                    {profile.user.firstName} {profile.user.lastName}
                  </h2>
                  <p className="profile-nickname">@{profile.user.nickname || profile.user.firstName?.toLowerCase()}</p>
                  
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
                <div className="stat clickable" onClick={openFollowersModal}>
                  <span className="stat-count">{profile.followersCount}</span>
                  <span className="stat-label">Followers</span>
                </div>
                <div className="stat clickable" onClick={openFollowingModal}>
                  <span className="stat-count">{profile.followingCount}</span>
                  <span className="stat-label">Following</span>
                </div>
              </div>
            </div>
            
            <div className="card posts-card">
              <h2 className="card-title">My Posts</h2>
              
              <button 
                className="primary-button create-post-btn" 
                onClick={navigateToCreatePost}
              >
                Create New Post
              </button>
              
              {profile.posts && profile.posts.length > 0 ? (
                <div className="posts-container">
                  {profile.posts.map(post => (
                    <div key={post.id} className="post-card">
                      <div className="post-header">
                      <div className="post-author" onClick={() => router.push(`/profile/${post.userId}`)}>
                    <div className="author-avatar">
                      {post.author.avatar ? (
                        <img src={post.author.avatar} alt={`${post.author.firstName}'s avatar`} />
                      ) : (
                        <div className="avatar-placeholder">
                          {post.author.firstName.charAt(0)}
                        </div>
                      )}
                    </div>
                    <div className="author-info">
                      <h3 className="author-name">
                        {post.author.nickname || `${post.author.firstName} ${post.author.lastName}`}
                      </h3>
                      <span className="post-date">{formatDate(post.createdAt)}</span>
                    </div>
                  </div>
                        
                        <div className="post-actions">
                          <button 
                            onClick={() => router.push(`/edit-post?id=${post.id}`)}
                            className="edit-post-btn"
                          >
                            Edit
                          </button>
                          <button 
                            onClick={() => handleDeletePost(post.id)}
                            className="delete-post-btn"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                      
                      <div className="post-content">
                        {post.content && <p className="post-text">{post.content}</p>}
                        {post.image && (
                          <div className="post-image-container">
                            <img 
                              src={post.image.startsWith('http') ? post.image : `http://localhost:8080/${post.image}`} 
                              alt="Post image" 
                              className="post-image" 
                            />
                          </div>
                        )}
                      </div>
                      
                    
                    </div>
                  ))}
                </div>
              ) : (
                <p className="empty-posts">No posts yet. Create your first post!</p>
              )}
            </div>
          </div>
        )}

        {/* Followers Modal */}
        {activeModal === 'followers' && profile && (
          <div className="modal-overlay" onClick={closeModal}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
              <div className="modal-header">
                <h3>Followers</h3>
                <button className="modal-close" onClick={closeModal}>×</button>
              </div>
              
              <div className="modal-body">
                {loadingUserDetails ? (
                  <p className="loading-text">Loading followers...</p>
                ) : profile.followers.length === 0 ? (
                  <p className="empty-list">No followers yet.</p>
                ) : (
                  <ul className="follow-list">
                    {profile.followers.map((follower, index) => (
                      <li key={follower.id || index} className="follow-item" onClick={() => navigateToProfile(follower.id)}>
                        <div className="follow-avatar">
                          {follower.avatar ? (
                            <img 
                              src={follower.avatar.startsWith('http') ? follower.avatar : `http://localhost:8080${follower.avatar}`} 
                              alt="Follower avatar" 
                            />
                          ) : (
                            <div className="avatar-placeholder">
                              {getUserDisplayName(follower.id).charAt(0).toUpperCase()}
                            </div>
                          )}
                        </div>
                        <div className="follow-info">
                          <p className="follow-name">{getUserDisplayName(follower.id)}</p>
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </div>
        )}

{activeModal === 'following' && profile && (
  <div className="modal-overlay" onClick={closeModal}>
    <div className="modal-content" onClick={e => e.stopPropagation()}>
      <div className="modal-header">
        <h3>Following</h3>
        <button className="modal-close" onClick={closeModal}>×</button>
      </div>

      <div className="modal-body">
        {loadingUserDetails ? (
          <p className="loading-text">Loading following...</p>
        ) : profile.following.length === 0 ? (
          <p className="empty-list">You're not following anyone yet.</p>
        ) : (
          <ul className="follow-list">
            {profile.following.map((followedUser, index) => (
              <li key={followedUser.id || index} className="follow-item" onClick={() => navigateToProfile(followedUser.id)}>
                <div className="follow-avatar">
                  {followedUser.avatar ? (
                    <img
                      src={followedUser.avatar.startsWith('http') ? followedUser.avatar : `http://localhost:8080${followedUser.avatar}`}
                      alt="User avatar"
                    />
                  ) : (
                    <div className="avatar-placeholder">
                      {getUserDisplayName(followedUser.id).charAt(0).toUpperCase()}
                    </div>
                  )}
                </div>
                <div className="follow-info">
                  <p className="follow-name">{getUserDisplayName(followedUser.id)}</p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  </div>
)}

      </main>
    </div>
  )      
}