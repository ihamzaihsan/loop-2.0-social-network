'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Sidebar from '../../components/Sidebar'
import styles from './groups.module.css'

interface Group {
  id: number
  title: string
  description: string
  creator_id: number
  created_at: string
  member_count: number
  is_creator: boolean
  has_pending_requests: boolean
}

interface GroupInvitation {
  id: number
  group_id: number
  group_title: string
  inviter_id: number
  inviter_name: string
  created_at: string
  status: string
}

export default function Groups() {
  const [groups, setGroups] = useState<Group[]>([])
  const [invitations, setInvitations] = useState<GroupInvitation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newGroup, setNewGroup] = useState({
    title: '',
    description: ''
  })
  
  const router = useRouter()

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        // Fetch user's groups
        const groupsResponse = await fetch('http://localhost:8080/groups/user', {
          credentials: 'include'
        })

        if (!groupsResponse.ok) {
          throw new Error('Failed to fetch groups')
        }

        const groupsData = await groupsResponse.json()
        if (groupsData.success) {
          setGroups(groupsData.groups || [])
        }

        // Fetch group invitations
        const invitationsResponse = await fetch('http://localhost:8080/groups/invitations', {
          credentials: 'include'
        })

        if (invitationsResponse.ok) {
          const invitationsData = await invitationsResponse.json()
          if (invitationsData.success) {
            setInvitations(invitationsData.invitations || [])
          }
        }
      } catch (err) {
        setError('Failed to load groups')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchGroups()
  }, [])

  const handleCreateGroup = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!newGroup.title.trim()) {
      return
    }

    try {
      const response = await fetch('http://localhost:8080/groups/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({
          title: newGroup.title,
          description: newGroup.description
        })
      })

      if (!response.ok) {
        throw new Error('Failed to create group')
      }

      const data = await response.json()
      if (data.success) {
        // Redirect to the new group page
        router.push(`/groups/${data.group_id}`)
      }
    } catch (err) {
      setError('Failed to create group')
      console.error(err)
    }
  }

  const handleInvitationResponse = async (groupId: number, invitationId: number, action: string) => {
    try {
      const response = await fetch('http://localhost:8080/groups/membership/handle', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({
          group_id: groupId,
          user_id: 0, // Not needed for invitations as the current user is the invitee
          action: action,
          request_type: 'invitation'
        })
      })

      if (!response.ok) {
        throw new Error(`Failed to ${action} invitation`)
      }

      // Update invitations list
      setInvitations(invitations.filter(inv => inv.id !== invitationId))

      // If accepted, refresh groups list
      if (action === 'accept') {
        const groupsResponse = await fetch('http://localhost:8080/groups/user', {
          credentials: 'include'
        })
        
        if (groupsResponse.ok) {
          const groupsData = await groupsResponse.json()
          if (groupsData.success) {
            setGroups(groupsData.groups || [])
          }
        }
      }
    } catch (err) {
      setError(`Failed to ${action} invitation`)
      console.error(err)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString()
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <Sidebar activePage="groups" />
        <main className={styles.content}>
          <div className={styles.loading}>Loading groups...</div>
        </main>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Sidebar activePage="groups" />
      
      <main className={styles.content}>
        <h1 className={styles.pageTitle}>My Groups</h1>
        
        {error && <div className={styles.error}>{error}</div>}
        
        <div className={styles.actionsBar}>
          <button 
            className={styles.createButton}
            onClick={() => setShowCreateModal(true)}
          >
            Create New Group
          </button>
        </div>
        
        {invitations.length > 0 && (
          <div className={styles.invitationsSection}>
            <h2>Group Invitations</h2>
            
            <div className={styles.invitationsList}>
              {invitations.map(invitation => (
                <div key={invitation.id} className={styles.invitationCard}>
                  <div className={styles.invitationInfo}>
                    <h3>{invitation.group_title}</h3>
                    <p>Invited by: {invitation.inviter_name}</p>
                    <p className={styles.invitationDate}>Received: {formatDate(invitation.created_at)}</p>
                  </div>
                  
                  <div className={styles.invitationActions}>
                    <button 
                      className={styles.acceptButton}
                      onClick={() => handleInvitationResponse(invitation.group_id, invitation.id, 'accept')}
                    >
                      Accept
                    </button>
                    <button 
                      className={styles.rejectButton}
                      onClick={() => handleInvitationResponse(invitation.group_id, invitation.id, 'reject')}
                    >
                      Decline
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
        
        {groups.length > 0 ? (
          <div className={styles.groupsList}>
            {groups.map(group => (
              <div key={group.id} className={styles.groupCard}>
                <div className={styles.groupInfo}>
                  <h2 className={styles.groupTitle}>{group.title}</h2>
                  {group.description && (
                    <p className={styles.groupDescription}>{group.description}</p>
                  )}
                  <div className={styles.groupMeta}>
                    <span>{group.member_count} members</span>
                    <span>Created: {formatDate(group.created_at)}</span>
                    {group.is_creator && <span className={styles.creatorBadge}>Creator</span>}
                  </div>
                </div>
                
                <div className={styles.groupActions}>
                  {group.has_pending_requests && group.is_creator && (
                    <div className={styles.pendingBadge}>Pending Requests</div>
                  )}
                  <button 
                    className={styles.viewButton}
                    onClick={() => router.push(`/groups/${group.id}`)}
                  >
                    View Group
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className={styles.emptyState}>
            <p>You're not a member of any groups yet.</p>
            <p>Create a new group or accept invitations to get started!</p>
          </div>
        )}
      </main>
      
      {showCreateModal && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <h2>Create New Group</h2>
            
            <form onSubmit={handleCreateGroup}>
              <div className={styles.formGroup}>
                <label htmlFor="groupTitle">Group Title *</label>
                <input
                  type="text"
                  id="groupTitle"
                  value={newGroup.title}
                  onChange={(e) => setNewGroup({...newGroup, title: e.target.value})}
                  required
                />
              </div>
              
              <div className={styles.formGroup}>
                <label htmlFor="groupDescription">Description</label>
                <textarea
                  id="groupDescription"
                  value={newGroup.description}
                  onChange={(e) => setNewGroup({...newGroup, description: e.target.value})}
                  rows={4}
                />
              </div>
              
              <div className={styles.modalActions}>
                <button 
                  type="button"
                  className={styles.cancelButton}
                  onClick={() => setShowCreateModal(false)}
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  className={styles.createButton}
                  disabled={!newGroup.title.trim()}
                >
                  Create Group
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
