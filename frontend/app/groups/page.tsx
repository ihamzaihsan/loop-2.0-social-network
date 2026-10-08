'use client'

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import './groups.css'
import { Group } from './types'

export default function GroupsPage() {
    const [groups, setGroups] = useState<Group[]>([])
    const [loading, setLoading] = useState(true)
    const router = useRouter()

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

    useEffect(() => { fetchGroups() }, [])
    useRealtimeRefresh(['groups', 'profiles'], fetchGroups)


    const handleGroupClick = (groupId: number) => {
        console.log('Group clicked:', groupId)
        router.push(`/groups/${groupId}`)
    }

    const handleCreateGroup = () => {
        router.push('/groups/create')
    }

    return (
        <div className="home-page">
            <Sidebar activePage="groups" />

            <main className="main-content">
                <div className="dashboard">
                    <div className="card feed-card">
                        <div className="groups-hero-section">
                            <div className="hero-background">
                                <div className="hero-pattern"></div>
                                <div className="hero-gradient"></div>
                            </div>
                            <div className="hero-content">
                                <div className="hero-text">
                                    <h1 className="hero-title">
                                        <span className="title-icon">🌟</span>
                                        My Groups
                                    </h1>
                                    <p className="hero-description">
                                        Connect, collaborate, and build amazing communities together
                                    </p>
                                    <div className="hero-stats">
                                        <div className="stat-item">
                                            <span className="stat-number">{groups.length}</span>
                                            <span className="stat-label">Groups Joined</span>
                                        </div>
                                        <div className="stat-divider"></div>
                                        <div className="stat-item">
                                            <span className="stat-number">
                                                {groups.reduce((total, group) => total + (group.member_count || 0), 0)}
                                            </span>
                                            <span className="stat-label">Total Members</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="hero-actions">
                                    <button
                                        className="hero-primary-button"
                                        onClick={handleCreateGroup}
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
                                            onClick={() => router.push('/groups/discover')}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                <circle cx="11" cy="11" r="8"></circle>
                                                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                                            </svg>
                                            Discover Groups
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

                        <div className="groups-content-section">
                            {loading ? (
                                <div className="loading-state">
                                    <div className="loading-spinner">
                                        <div className="spinner"></div>
                                    </div>
                                    <h3>Loading your groups...</h3>
                                    <p>Gathering your communities</p>
                                </div>
                            ) : groups.length === 0 ? (
                                <div className="empty-state-enhanced">
                                    <div className="empty-illustration">
                                        <div className="empty-icon">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                                                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                                                <circle cx="9" cy="7" r="4"></circle>
                                                <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                                                <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                                            </svg>
                                        </div>
                                        <div className="empty-sparkles">
                                            <span className="sparkle sparkle-1">✨</span>
                                            <span className="sparkle sparkle-2">⭐</span>
                                            <span className="sparkle sparkle-3">💫</span>
                                        </div>
                                    </div>
                                    <div className="empty-content">
                                        <h3>Your Group Journey Starts Here!</h3>
                                        <p>Join communities, share ideas, and connect with like-minded people. The perfect group is waiting for you!</p>
                                        <div className="empty-actions">
                                            <button className="empty-primary-button" onClick={handleCreateGroup}>
                                                <span className="button-icon">🚀</span>
                                                Create Your First Group
                                            </button>
                                            <button className="empty-secondary-button" onClick={() => router.push('/groups/discover')}>
                                                <span className="button-icon">🔍</span>
                                                Explore Existing Groups
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <div className="groups-showcase">
                                    <div className="showcase-header">
                                        <h2 className="showcase-title">Your Communities</h2>
                                        <div className="showcase-filters">
                                            <button className="filter-button active">All Groups</button>
                                        </div>
                                    </div>
                                    <div className="groups-grid-enhanced">
                                        {Array.isArray(groups) && groups.map((group, index) => (
                                            <div
                                                key={group.id}
                                                className="group-card-enhanced"
                                                onClick={() => handleGroupClick(group.id)}
                                                style={{
                                                    cursor: 'pointer',
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
                                                        <div className="activity-indicator">
                                                            <div className="activity-dot"></div>
                                                            <span>Active community</span>
                                                        </div>
                                                        <div className="card-arrow">
                                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                                <path d="M7 17L17 7M17 7H7M17 7V17"/>
                                                            </svg>
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