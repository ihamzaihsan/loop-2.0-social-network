'use client'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import styles from './discover.module.css'
import '../groupsSidebar.css'
import { Group } from '../types'


export default function DiscoverPage() {
  const [allGroups, setAllGroups] = useState<Group[]>([])
  const [userGroups, setUserGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(true)
  const [requestingJoin, setRequestingJoin] = useState<{ [key: number]: boolean }>({})
  const [error, setError] = useState('')
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
      <div className={styles.discoverMainContainer}>
          <div className="groups-page">
              <Sidebar activePage="groups" />

              <div className="groups-container">
                  <div className="groups-sidebar">
                      <div className="groups-sidebar-header">
                          <h2>Discover Groups</h2>
                          <button
                              className="new-group-button"
                              onClick={() => router.push('/groups/create')}
                          >
                              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                  <line x1="12" y1="5" x2="12" y2="19"></line>
                                  <line x1="5" y1="12" x2="19" y2="12"></line>
                              </svg>
                              Create
                          </button>
                      </div>

                      <div className="groups-list-container">
                          <div className="groups-nav">
                              <button
                                  className="groups-nav-button active"
                                  onClick={() => router.push('/groups/discover')}
                              >
                                  Discover
                              </button>
                              <button
                                  className="groups-nav-button"
                                  onClick={() => router.push('/groups')}
                              >
                                  My Groups
                              </button>
                          </div>

                          {loading ? (
                              <div className="loading-message">Loading groups...</div>
                          ) : allGroups.length === 0 ? (
                              <div className="empty-list-message">
                                  <p>No groups available to join.</p>
                                  <button
                                      className="create-first-group"
                                      onClick={() => router.push('/groups/create')}
                                  >
                                      Create a New Group
                                  </button>
                              </div>
                          ) : (
                              <ul className="groups-list">
                                  {Array.isArray(allGroups) && allGroups.map((group) => (
                                      <li key={group.id} className="group-item">
                                          <div className="group-avatar">
                                              <div className="avatar-placeholder">
                                                  {group.title.charAt(0)}
                                              </div>
                                          </div>
                                          <div className="group-info">
                                              <span className="group-name">{group.title}</span>
                                              <p className="group-description">{group.description}</p>
                                              <div className="group-meta">
                                                  <span className="member-count">{group.member_count || 0} members</span>
                                              </div>
                                          </div>
                                          <div className="group-action">
                                              {isUserMember(group.id) ? (
                                                  <span className="member-badge">Member</span>
                                              ) : hasRequestedToJoin(group) ? (
                                                  <span className="requested-badge">Requested</span>
                                              ) : success[group.id] ? (
                                                  <span className="requested-badge">Requested</span>
                                              ) : (
                                                  <button
                                                      className="join-button"
                                                      onClick={(e) => {
                                                          e.stopPropagation();
                                                          handleRequestJoin(group.id);
                                                      }}
                                                      disabled={requestingJoin[group.id]}
                                                  >
                                                      {requestingJoin[group.id] ? 'Requesting...' : 'Join'}
                                                  </button>
                                              )}
                                          </div>
                                      </li>
                                  ))}
                              </ul>
                          )}
                      </div>
                  </div>

              </div>
          </div>
      </div>
  )
}
