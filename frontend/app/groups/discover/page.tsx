'use client'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import './discover.css'
import { Group } from '../types'


export default function DiscoverPage() {
  const [allGroups, setAllGroups] = useState<Group[]>([])
  const [userGroups, setUserGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(true)
  const [requestingJoin, setRequestingJoin] = useState<{ [key: number]: boolean }>({})
  const [, setError] = useState('')
  const [success, setSuccess] = useState<{ [key: number]: boolean }>({})
  const router = useRouter()

  useEffect(() => {
      const fetchGroups = async () => {
          try {
              // Fetch all available groups
              const allGroupsResponse = await fetch('http://localhost:8080/groups/all', {
                  method: 'GET',
                  credentials: 'include'
              })

              // Fetch user's groups to check membership
              const userGroupsResponse = await fetch('http://localhost:8080/groups/user', {
                  method: 'GET',
                  credentials: 'include'
              })

              if (allGroupsResponse.ok && userGroupsResponse.ok) {
                  const allGroupsData = await allGroupsResponse.json()
                  const userGroupsData = await userGroupsResponse.json()

                  // Process all groups
                  let allGroupsList: Group[] = []
                  if (Array.isArray(allGroupsData)) {
                      allGroupsList = allGroupsData
                  } else if (allGroupsData && typeof allGroupsData === 'object') {
                      if (Array.isArray(allGroupsData.groups)) {
                          allGroupsList = allGroupsData.groups
                      } else {
                          const groupsArray = Object.values(allGroupsData).filter(item =>
                              item && typeof item === 'object' && 'id' in item
                          ) as Group[]
                          allGroupsList = groupsArray
                      }
                  }

                  // Process user groups
                  let userGroupsList: Group[] = []
                  if (Array.isArray(userGroupsData)) {
                      userGroupsList = userGroupsData
                  } else if (userGroupsData && typeof userGroupsData === 'object') {
                      if (Array.isArray(userGroupsData.groups)) {
                          userGroupsList = userGroupsData.groups
                      } else {
                          const groupsArray = Object.values(userGroupsData).filter(item =>
                              item && typeof item === 'object' && 'id' in item
                          ) as Group[]
                          userGroupsList = groupsArray
                      }
                  }

                  setAllGroups(allGroupsList)
                  setUserGroups(userGroupsList)
              } else {
                  console.error('Failed to fetch groups')
                  setError('Failed to load groups. Please try again.')
              }
          } catch (error) {
              console.error('Error fetching groups:', error)
              setError('An error occurred while loading groups.')
          } finally {
              setLoading(false)
          }
      }

      fetchGroups()
  }, [])

  const handleRequestJoin = async (groupId: number) => {
      setRequestingJoin(prev => ({ ...prev, [groupId]: true }))
      setError('')

      try {
          const response = await fetch('http://localhost:8080/groups/join/request', {
              method: 'POST',
              headers: {
                  'Content-Type': 'application/json',
              },
              credentials: 'include',
              body: JSON.stringify({ group_id: groupId }),
          })

          if (response.ok) {
              setSuccess(prev => ({ ...prev, [groupId]: true }))
              setAllGroups(prev =>
                  prev.map(group =>
                      group.id === groupId
                          ? { ...group, status: 'requested' }
                          : group
                  )
              )
          } else {
              try {
                  // Check if the response is JSON
                  const contentType = response.headers.get("content-type");
                  if (contentType && contentType.includes("application/json")) {
                      const errorData = await response.json();
                      setError(errorData.error || 'Failed to request joining the group');
                  } else {
                      // Handle plain text or other non-JSON responses
                      const errorText = await response.text();
                      setError(errorText || 'Failed to request joining the group');
                  }
              } catch (parseError) {
                  console.error('Error parsing error response:', parseError);
                  setError('Failed to request joining the group');
              }
          }
      } catch (err) {
          console.error('Error requesting to join group:', err)
          setError('An error occurred. Please try again.')
      } finally {
          setRequestingJoin(prev => ({ ...prev, [groupId]: false }))
      }
  }

  const isUserMember = (groupId: number) => {
      return userGroups.some(group => group.id === groupId)
  }

  const hasRequestedToJoin = (group: Group) => {
      return group.status === 'requested' || group.status === 'invited'
  }

  return (
      <div className="home-page">
              <Sidebar activePage="groups" />

            <main className="main-content">
                <div className="dashboard">
                    <div className="card feed-card">
                        <div className="discover-hero-section">
                            <div className="hero-background">
                                <div className="hero-pattern"></div>
                                <div className="hero-gradient"></div>
                            </div>
                            <div className="hero-content">
                                <div className="hero-text">
                                    <h1 className="hero-title">
                                        <span className="title-icon">🔍</span>
                                        Discover Amazing Groups
                                    </h1>
                                    <p className="hero-description">
                                        Explore vibrant communities, connect with like-minded people, and join conversations that matter to you
                                    </p>
                                    <div className="hero-stats">
                                        <div className="stat-item">
                                            <span className="stat-number">{allGroups.length}</span>
                                            <span className="stat-label">Groups Available</span>
                                        </div>
                                        <div className="stat-divider"></div>
                                        <div className="stat-item">
                                            <span className="stat-number">
                                                {allGroups.filter(group => !isUserMember(group.id)).length}
                                            </span>
                                            <span className="stat-label">Ready to Join</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="hero-actions">
                                    <button
                                        className="hero-primary-button"
                                        onClick={() => router.push('/groups/create')}
                                    >
                                        <div className="button-icon">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                <line x1="12" y1="5" x2="12" y2="19"></line>
                                                <line x1="5" y1="12" x2="19" y2="12"></line>
                                            </svg>
                                        </div>
                                        <span>Create New Group</span>
                                    </button>
                                    <div className="hero-secondary-actions">
                                        <button
                                            className="hero-secondary-button"
                                            onClick={() => router.push('/groups')}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                <path d="M19 12H5M12 19l-7-7 7-7" />
                                            </svg>
                                            My Groups
                                        </button>
                                        <button
                                            className="hero-secondary-button"
                                            onClick={() => router.push('/groups/invitations')}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                                            </svg>
                                            View Invitations
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="discover-content-section">
                            {loading ? (
                                <div className="loading-state">
                                    <div className="loading-spinner">
                                        <div className="spinner"></div>
                                    </div>
                                    <h3>Discovering amazing groups...</h3>
                                    <p>Finding the perfect communities for you</p>
                                </div>
                            ) : allGroups.length === 0 ? (
                                <div className="empty-state-enhanced">
                                    <div className="empty-illustration">
                                        <div className="empty-icon">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                                                <circle cx="11" cy="11" r="8"></circle>
                                                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                                                <circle cx="11" cy="11" r="3"></circle>
                                            </svg>
                                        </div>
                                        <div className="empty-sparkles">
                                            <span className="sparkle sparkle-1">🌟</span>
                                            <span className="sparkle sparkle-2">✨</span>
                                            <span className="sparkle sparkle-3">💫</span>
                                        </div>
                                    </div>
                                    <div className="empty-content">
                                        <h3>No Groups to Discover Yet!</h3>
                                        <p>Be a pioneer! Create the first group and start building an amazing community that others will want to join.</p>
                                        <div className="empty-actions">
                                            <button className="empty-primary-button" onClick={() => router.push('/groups/create')}>
                                                <span className="button-icon">🚀</span>
                                                Create the First Group
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <div className="discover-showcase">
                                    <div className="showcase-header">
                                        <h2 className="showcase-title">Available Communities</h2>
                                        <div className="showcase-filters">
                                            <button className="filter-button active">All Groups</button>
                                        </div>
                                    </div>
                                    <div className="discover-grid-enhanced">
                                        {Array.isArray(allGroups) && allGroups.map((group, index) => (
                                            <div
                                                key={group.id}
                                                className="discover-card-enhanced"
                                                style={{
                                                    animationDelay: `${index * 0.1}s`
                                                }}
                                            >
                                                <div className="card-background">
                                                    <div className="card-pattern"></div>
                                                    <div className="card-glow"></div>
                                                </div>
                                                <div className="card-content">
                                                    <div className="group-header-enhanced">
                                                        <div className="group-avatar-enhanced">
                                                            <div className="avatar-background"></div>
                                                            <div className="avatar-letter">
                                                                {group.title.charAt(0)}
                                                            </div>
                                                            <div className="avatar-ring"></div>
                                                        </div>
                                                        <div className="group-badge">
                                                            <span className="badge-icon">👥</span>
                                                            <span className="badge-text">{group.member_count || 0}</span>
                                                        </div>
                                                    </div>
                                                    <div className="group-details">
                                                        <h3 className="group-title-enhanced">{group.title}</h3>
                                                        {group.description && (
                                                            <p className="group-description-enhanced">{group.description}</p>
                                                        )}
                                                    </div>
                                                    <div className="group-footer">
                                                        <div className="join-status">
                                                            {isUserMember(group.id) ? (
                                                                <span className="status-badge member-status">
                                                                    <span className="status-icon">✓</span>
                                                                    Member
                                                                </span>
                                                            ) : hasRequestedToJoin(group) ? (
                                                                <span className="status-badge pending-status">
                                                                    <span className="status-icon">⏳</span>
                                                                    Pending
                                                                </span>
                                                            ) : success[group.id] ? (
                                                                <span className="status-badge sent-status">
                                                                    <span className="status-icon">📤</span>
                                                                    Sent
                                                                </span>
                                                            ) : (
                                                                <button
                                                                    className="join-button-enhanced"
                                                                    onClick={(e) => {
                                                                        e.stopPropagation();
                                                                        handleRequestJoin(group.id);
                                                                    }}
                                                                    disabled={requestingJoin[group.id]}
                                                                >
                                                                    {requestingJoin[group.id] ? (
                                                                        <>
                                                                            <div className="button-spinner"></div>
                                                                            <span>Requesting...</span>
                                                                        </>
                                                                    ) : (
                                                                        <>
                                                                            <span className="button-icon">🚀</span>
                                                                            <span>Join Group</span>
                                                                        </>
                                                                    )}
                                                                </button>
                                                            )}
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
