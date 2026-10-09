'use client'
import { API } from "../../utils/api";
import Link from 'next/link'
import './login.css'
import { Suspense, useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { redirectBasedOnSession } from '../../utils/session'
import { WebSocketClient } from '../webscoket/websocket'
import AuthShowcase from '../../components/AuthShowcase'
import BrandMark from '../../components/BrandMark'
import GoogleSignIn from '../../components/GoogleSignIn'

function LoginContent() {
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
      const response = await fetch(`${API}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData),
        credentials: 'include'
      })

      const data = await response.json();
      
      if (response.ok) {
        localStorage.removeItem('sessionToken');
        
        // Initialize WebSocket connection immediately after successful login
        console.log('Login successful, initializing WebSocket connection...');
        const wsClient = WebSocketClient.getInstance();
        
        // Connect after the browser has accepted the HttpOnly cookie
        setTimeout(() => {
          wsClient.connect();
          console.log('WebSocket connection initiated after login');
        }, 100);
        
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
      <AuthShowcase mode="login" />
      <div className="login-form-wrapper">
        <div className="auth-mobile-brand"><BrandMark /></div>
        <div className="auth-form-heading">
          <span className="eyebrow">Welcome back</span>
          <h2 className="form-title">Sign in to your circle</h2>
          <p>Pick up where the conversation left off.</p>
        </div>
        
        {/* Display error message if there is one */}
        {error && <div className="error-message">{error}</div>}
        
        <GoogleSignIn />
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
            Continue to Loop
          </button>

          <div className="register-link">
            New around here? <Link href="/register" className="text-link">Create an account</Link>
          </div>
        </form><a href="/forgot-password">Forgot password?</a>
      </div>
    </div>
  );
}

export default function Login() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <LoginContent />
    </Suspense>
  )
}
