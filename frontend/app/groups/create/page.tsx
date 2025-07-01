'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Sidebar from '@/components/Sidebar'
import './create.css'

export default function CreateGroupPage() {
    const [title, setTitle] = useState('')
    const [description, setDescription] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')
    const router = useRouter()

    interface CreateGroupResponse {
        group_id: string
    }

    const handleSubmit = async (e: { preventDefault: () => void }) => {
        e.preventDefault()

        if (!title.trim()) {
            setError('Group title is required')
            return
        }

        // Validate field lengths
        if (title.length > 100) {
            setError('Group title exceeds maximum length of 100 characters')
            return
        }

        if (description.length > 100) {
            setError('Group description exceeds maximum length of 100 characters')
            return
        }

        setLoading(true)
        setError('')

        try {
            const response = await fetch('http://localhost:8080/groups/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    title,
                    description
                }),
            })
            
            const responseText = await response.text()
            console.log('Raw server response:', responseText)
            
            if (response.ok) {
                try {
                    const data = JSON.parse(responseText) as CreateGroupResponse
                    router.push(`/groups/${data.group_id}`)
                } catch (parseError) {
                    console.error('Error parsing JSON response:', parseError)
                    setError('Received invalid response from server')
                }
            } else {
                setError(`Failed to create group: ${responseText}`)
            }
        } catch (err) {
            setError('An error occurred. Please try again.')
            console.error('Error creating group:', err)
        } finally {
            setLoading(false)
        }
    }
    return (
        <div className="home-page">
            <Sidebar activePage="groups" />

            <main className="main-content">
                <div className="dashboard">
                    <div className="card feed-card">
                        <div className="create-hero-section">
                            <div className="hero-background">
                                <div className="hero-pattern"></div>
                                <div className="hero-gradient"></div>
                            </div>
                            <div className="hero-content">
                                <div className="hero-text">
                                    <h1 className="hero-title">
                                        <span className="title-icon">🚀</span>
                                        Create Your Community
                                    </h1>
                                    <p className="hero-description">
                                        Build something amazing! Start a new group and bring together people who share your passion and interests.
                                    </p>
                                    <div className="hero-features">
                                        <div className="feature-item">
                                            <span className="feature-icon">👥</span>
                                            <span className="feature-text">Connect People</span>
                                        </div>
                                        <div className="feature-item">
                                            <span className="feature-icon">💬</span>
                                            <span className="feature-text">Share Ideas</span>
                                        </div>
                                        <div className="feature-item">
                                            <span className="feature-icon">🎯</span>
                                            <span className="feature-text">Achieve Goals</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="hero-actions">
                                    <button
                                        type="button"
                                        className="hero-secondary-button"
                                        onClick={() => router.back()}
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                            <path d="M19 12H5M12 19l-7-7 7-7" />
                                        </svg>
                                        Go Back
                                    </button>
                                </div>
                            </div>
                        </div>

                        <div className="create-form-section">
                            <div className="form-header">
                                <h2 className="form-title">Group Details</h2>
                                <p className="form-subtitle">Tell us about your community</p>
                            </div>

                            {error && (
                                <div className="error-message-enhanced">
                                    <div className="error-icon">⚠️</div>
                                    <div className="error-content">
                                        <h4>Oops! Something went wrong</h4>
                                        <p>{error}</p>
                                    </div>
                                </div>
                            )}

                            <form onSubmit={handleSubmit} className="create-form-enhanced">
                                <div className="form-group-enhanced">
                                    <label htmlFor="title" className="form-label-enhanced">
                                        <span className="label-icon">🏷️</span>
                                        <span className="label-text">Group Name*</span>
                                    </label>
                                    <div className="input-container">
                                        <input
                                            type="text"
                                            id="title"
                                            className="form-input-enhanced"
                                            value={title}
                                            onChange={(e) => setTitle(e.target.value)}
                                            placeholder="Enter a catchy name for your group"
                                            required
                                            maxLength={100}
                                        />
                                        <div className="input-border"></div>
                                    </div>
                                    <div className="input-help">
                                        Choose a name that reflects your group's purpose and attracts the right members
                                    </div>
                                </div>

                                <div className="form-group-enhanced">
                                    <label htmlFor="description" className="form-label-enhanced">
                                        <span className="label-icon">📝</span>
                                        <span className="label-text">Description</span>
                                    </label>
                                    <div className="textarea-container">
                                        <textarea
                                            id="description"
                                            className="form-textarea-enhanced"
                                            value={description}
                                            onChange={(e) => setDescription(e.target.value)}
                                            placeholder="What makes your group special? Describe its purpose, goals, and what members can expect to gain from joining..."
                                            rows={6}
                                            maxLength={100}
                                        />
                                        <div className="textarea-border"></div>
                                        <div className="character-count">
                                            {description.length}/100 characters
                                        </div>
                                    </div>
                                    <div className="input-help">
                                        A compelling description helps potential members understand what your group is about
                                    </div>
                                </div>

                                <div className="form-actions-enhanced">
                                    <button
                                        type="button"
                                        className="cancel-button-enhanced"
                                        onClick={() => router.back()}
                                        disabled={loading}
                                    >
                                        <span className="button-icon">↩️</span>
                                        <span>Cancel</span>
                                    </button>
                                    <button
                                        type="submit"
                                        className="submit-button-enhanced"
                                        disabled={loading || !title.trim()}
                                    >
                                        {loading ? (
                                            <>
                                                <div className="button-spinner"></div>
                                                <span>Creating Your Group...</span>
                                            </>
                                        ) : (
                                            <>
                                                <span className="button-icon">🎉</span>
                                                <span>Create Group</span>
                                            </>
                                        )}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
