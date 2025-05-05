'use client'
import Link from 'next/link'
import './login.css'
import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { redirectBasedOnSession } from '../../utils/session'

export default function Login() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    // Check if redirected due to session invalidation
    const reason = searchParams.get('reason')
    if (reason === 'session_expired') {
      setError('Your session was ended because you logged in on another device')
    }
    
    redirectBasedOnSession(router, false).finally(() => {
      setLoading(false)
    })
  }, [router, searchParams])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    
    try {
      const response = await fetch('http://localhost:8080/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData),
        credentials: 'include'
      })

      const data = await response.json();
      
      if (response.ok) {
        // Get the token from the response
        if (data.token) {
          localStorage.setItem('sessionToken', data.token);
        }
        router.push('/home')
      } else {
        // Display the error message from the server
        setError(data.error || 'Login failed. Please try again.');
      }
    } catch (error) {
      console.error('Login failed:', error)
      setError('An unexpected error occurred. Please try again.');
    }
  }

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="login-container">
      <div className="login-form-wrapper">
        <div className="nav-buttons">
          <Link href="/">
            <button className="back-button">Back to Home</button>
          </Link>
        </div>

        <h2 className="form-title">Login</h2>
        
        {/* Display error message if there is one */}
        {error && <div className="error-message">{error}</div>}
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Email</label>
            <input 
              type="email" 
              required 
              className="form-input" 
              placeholder="Enter your email"
              value={formData.email}
              onChange={(e) => setFormData({...formData, email: e.target.value})}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Password</label>
            <input 
              type="password" 
              required 
              className="form-input" 
              placeholder="Enter your password"
              value={formData.password}
              onChange={(e) => setFormData({...formData, password: e.target.value})}
            />
          </div>

          <button type="submit" className="submit-button">
            Login
          </button>

          <div className="register-link">
            Don't have an account? <Link href="/register" className="text-link">Register</Link>
          </div>
        </form>
      </div>
    </div>
  );
}
