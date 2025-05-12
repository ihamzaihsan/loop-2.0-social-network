'use client'
import './register.css'
import Link from 'next/link'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { redirectBasedOnSession } from '../../utils/session'

export default function Register() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [formData, setFormData] = useState<{
    email: string;
    password: string;
    firstName: string;
    lastName: string;
    dob: string;
    nickname: string;
    aboutMe: string;
    avatar: File | null;
    isPrivate: boolean;
  }>({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    dob: '',
    nickname: '',
    aboutMe: '',
    avatar: null,
    isPrivate: false
  });
  
  // Check session on page load
  useEffect(() => {
    redirectBasedOnSession(router, false).finally(() => {
      setLoading(false)
    })
  }, [router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    // Clear any previous error messages
    setErrorMessage(null);
    
    try {
      // Check if we're sending a file
      if (formData.avatar) {
        // Use FormData for file uploads
        const form = new FormData()
        form.append('email', formData.email)
        form.append('password', formData.password)
        form.append('firstName', formData.firstName)
        form.append('lastName', formData.lastName)
        form.append('dob', formData.dob)
        
        if (formData.nickname) {
          form.append('nickname', formData.nickname)
        }
        
        if (formData.aboutMe) {
          form.append('aboutMe', formData.aboutMe)
        }
        
        // Add the avatar file
        form.append('avatar', formData.avatar)
        
        // Add isPrivate as a string
        form.append('isPrivate', formData.isPrivate ? 'true' : 'false')
        
        const response = await fetch('http://localhost:8080/register', {
          method: 'POST',
          // Don't set Content-Type - browser will set it with boundary for FormData
          body: form,
          credentials: 'include'
        })
  
        if (response.ok) {
          router.push('/home')
        } else {
          const errorText = await response.text()
          console.log('Server response:', errorText)
          
          // Check if it's a duplicate email error
          if (response.status === 409) {
            setErrorMessage('This email is already registered. Please use a different email or login.')
          } else {
            setErrorMessage(`Registration failed: ${errorText}`)
          }
        }
      } else {
        // No file, use JSON
        const response = await fetch('http://localhost:8080/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            email: formData.email,
            password: formData.password,
            firstName: formData.firstName,
            lastName: formData.lastName,
            dob: formData.dob,
            nickname: formData.nickname || null,
            aboutMe: formData.aboutMe || null,
            isPrivate: formData.isPrivate
          }),
          credentials: 'include'
        })
  
        if (response.ok) {
          router.push('/home')
        } else {
          const errorText = await response.text()
          console.log('Server response:', errorText)
          
          // Check if it's a duplicate email error
          if (response.status === 409) {
            setErrorMessage('This email is already registered. Please use a different email or login.')
          } else {
            setErrorMessage(`Registration failed: ${errorText}`)
          }
        }
      }
    } catch (error) {
      console.log('Fetch error:', error)
      setErrorMessage('An unexpected error occurred. Please try again.')
    }
  }
  
  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="register-container">
      <div className="register-form-wrapper">
        <div className="nav-buttons">
          <Link href="/">
            <button className="back-button">Back to Home</button>
          </Link>
        </div>

        <h2 className="form-title">Create Account</h2>
        
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label className="form-label">First Name *</label>
              <input 
                type="text" 
                required 
                className="form-input" 
                placeholder="First Name" 
                value={formData.firstName} 
                onChange={(e) => setFormData({...formData, firstName: e.target.value})} 
              />
            </div>
            <div className="form-group">
              <label className="form-label">Last Name *</label>
              <input 
                type="text" 
                required 
                className="form-input" 
                placeholder="Last Name" 
                value={formData.lastName} 
                onChange={(e) => setFormData({...formData, lastName: e.target.value})} 
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Email *</label>
            <input 
              type="email" 
              required 
              className="form-input" 
              placeholder="email@example.com" 
              value={formData.email} 
              onChange={(e) => setFormData({...formData, email: e.target.value})} 
            />
          </div>

          <div className="form-group">
            <label className="form-label">Password *</label>
            <input 
              type="password" 
              required 
              className="form-input" 
              placeholder="••••••••" 
              value={formData.password} 
              onChange={(e) => setFormData({...formData, password: e.target.value})} 
            />
          </div>

          <div className="form-group">
            <label className="form-label">Date of Birth *</label>
            <input 
              type="date" 
              required 
              className="form-input" 
              value={formData.dob} 
              onChange={(e) => setFormData({...formData, dob: e.target.value})} 
            />
          </div>

          <div className="form-group">
            <label className="form-label">Nickname</label>
            <input 
              type="text" 
              className="form-input" 
              placeholder="Your nickname (optional)" 
              value={formData.nickname} 
              onChange={(e) => setFormData({...formData, nickname: e.target.value})} 
            />
          </div>

          <div className="form-group">
            <label className="form-label">About Me</label>
            <textarea 
              className="form-textarea" 
              rows={3} 
              placeholder="Tell us about yourself (optional)" 
              value={formData.aboutMe} 
              onChange={(e) => setFormData({...formData, aboutMe: e.target.value})} 
            />
          </div>

          <div className="form-group">
            <label className="form-label">Avatar</label>
            <input 
              type="file" 
              accept="image/*" 
              className="file-input" 
              onChange={(e) => setFormData({...formData, avatar: e.target.files ? e.target.files[0] : null})} 
            />
          </div>

          <div className="form-group">
            <label className="form-checkbox-label">
              <input 
                type="checkbox" 
                className="form-checkbox" 
                checked={formData.isPrivate} 
                onChange={(e) => setFormData({...formData, isPrivate: e.target.checked})} 
              />
              Make my account private
            </label>
            <p className="form-help-text">Private accounts limit who can see your posts and profile information</p>
          </div>

          {errorMessage && (
            <div className="error-message">
              {errorMessage}
            </div>
          )}

          <button type="submit" className="submit-button">
            Create An Account
          </button>

          <div className="login-link">
            Already have an account? <Link href="/login" className="text-link">Login</Link>
          </div>
        </form>
      </div>
    </div>
  );
}