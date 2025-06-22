'use client'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import './groupsSidebar.css'
import { Group } from './types'

export default function GroupsPage() {
    const [groups, setGroups] = useState<Group[]>([])
    const [loading, setLoading] = useState(true)
    const router = useRouter()

    useEffect(() => {
        const fetchGroups = async () => {
            try {
                const response = await fetch('http://localhost:8080/groups/user', {
                    method: 'GET',
                    credentials: 'include'
                })

                if (response.ok) {
                    const data = await response.json()
                    if (Array.isArray(data)) {
                        setGroups(data)
                    } else if (data && typeof data === 'object') {
                        if (Array.isArray(data.groups)) {
                            setGroups(data.groups)
                        } else {
                            console.warn('API returned object instead of array, attempting to convert')
                            const groupsArray = Object.values(data).filter(item =>
                                item && typeof item === 'object' && 'id' in item
                            ) as Group[]
                            setGroups(groupsArray)
                        }
                    } else {
                        console.error('Unexpected data format:', data)
                        setGroups([])
                    }
                } else {
                    console.error('Failed to fetch groups')
                }
            } catch (error) {
                console.error('Error fetching groups:', error)
            } finally {
                setLoading(false)
            }
        }

        fetchGroups()
    }, [])

    const handleGroupClick = (groupId: number) => {
        console.log('Group clicked:', groupId) 
        router.push(`/groups/${groupId}`)
    }

    const handleCreateGroup = () => {
        router.push('/groups/create')
    }

    return (
        <div className="groups-page">
            <Sidebar activePage="groups" />

            <div className="groups-container">
                <div className="groups-sidebar">
                    <div className="groups-sidebar-header">
                        <h2>Groups</h2>
                        <div className="header-buttons">
                            <button 
                                className="invitations-button" 
                                onClick={() => router.push('/groups/invitations')}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                                </svg>
                                Invitations
                            </button>
                            <button 
                                className="discover-button" 
                                onClick={() => router.push('/groups/discover')}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <circle cx="11" cy="11" r="8"></circle>
                                    <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                                </svg>
                                Discover
                            </button>
                            <button className="new-group-button" onClick={handleCreateGroup}>
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <line x1="12" y1="5" x2="12" y2="19"></line>
                                    <line x1="5" y1="12" x2="19" y2="12"></line>
                                </svg>
                                Create
                            </button>
                        </div>
                    </div>

                    <div className="groups-list-container">
                        {loading ? (
                            <div className="loading-message">Loading groups...</div>
                        ) : groups.length === 0 ? (
                            <div className="empty-list-message">
                                <p>You haven't joined any groups yet.</p>
                                <button className="create-first-group" onClick={handleCreateGroup}>
                                    Create Your First Group
                                </button>
                            </div>
                        ) : (
                            <ul className="groups-list">
                                {Array.isArray(groups) && groups.map((group) => (
                                    <li
                                        key={group.id}
                                        className="group-item"
                                        onClick={() => handleGroupClick(group.id)}
                                        style={{ cursor: 'pointer' }} 
                                    >
                                        <div className="group-avatar">
                                            <div className="avatar-placeholder">
                                                {group.title.charAt(0)}
                                            </div>
                                        </div>
                                        <div className="group-info">
                                            <span className="group-name">{group.title}</span>
                                            <p className="group-description">{group.description}</p>
                                        </div>
                                        <div className="group-meta">
                                            <span className="member-count">{group.member_count || 0} members</span>
                                        </div>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                </div>

                <div className="groups-main">
                    <div className="empty-group-selection">
                        <div className="empty-group-message">
                            <h3>Select a group or create a new one</h3>
                            <p>Choose a group from the sidebar or create a new group to get started.</p>
                            <button className="new-group-button-large" onClick={handleCreateGroup}>
                                Create New Group
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}