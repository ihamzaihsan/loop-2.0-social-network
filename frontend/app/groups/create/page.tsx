'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Sidebar from '@/components/Sidebar'
import styles from './createGroup.module.css'

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
        <div className={styles.container}>
            <Sidebar activePage="groups" />
            <main className={styles.main}>
                <div className={styles.formContainer}>
                    <h1>Create a New Group</h1>

                    {error && <div className={styles.error}>{error}</div>}

                    <form onSubmit={handleSubmit} className={styles.form}>
                        <div className={styles.formGroup}>
                            <label htmlFor="title">Group Name*</label>
                            <input
                                type="text"
                                id="title"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                placeholder="Enter a name for your group"
                                required
                            />
                        </div>

                        <div className={styles.formGroup}>
                            <label htmlFor="description">Description</label>
                            <textarea
                                id="description"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="What is this group about?"
                                rows={5}
                            />
                        </div>

                        <div className={styles.buttonGroup}>
                            <button
                                type="button"
                                className={styles.cancelButton}
                                onClick={() => router.back()}
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                className={styles.createButton}
                                disabled={loading}
                            >
                                {loading ? 'Creating...' : 'Create Group'}
                            </button>
                        </div>
                    </form>
                </div>
            </main>
        </div>
    )
}
