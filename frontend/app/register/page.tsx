'use client'
import './register.css'
import Link from 'next/link'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { redirectBasedOnSession } from '../../utils/session'

export default function Register() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
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
    
    try {
        const response = await fetch('http://localhost:8080/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData),
            credentials: 'include'
        })

        if (response.ok) {
            router.push('/home')
        } else {
            const errorText = await response.text()
            console.log('Server response:', errorText)
        }
    } catch (error) {
        console.log('Fetch error:', error)
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
              <input type="text" required className="form-input" placeholder="Hussain" value={formData.firstName} onChange={(e) => setFormData({...formData, firstName: e.target.value})} />
            </div>
            <div className="form-group">
              <label className="form-label">Last Name *</label>
              <input type="text" required className="form-input" placeholder="Ali" value={formData.lastName} onChange={(e) => setFormData({...formData, lastName: e.target.value})} />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Email *</label>
            <input type="email" required className="form-input" placeholder="Hussain@example.com" value={formData.email} onChange={(e) => setFormData({...formData, email: e.target.value})} />
          </div>

          <div className="form-group">
            <label className="form-label">Password *</label>
            <input type="password" required className="form-input" placeholder="••••••••" value={formData.password} onChange={(e) => setFormData({...formData, password: e.target.value})} />
          </div>

          <div className="form-group">
            <label className="form-label">Date of Birth *</label>
            <input type="date" required className="form-input" value={formData.dob} onChange={(e) => setFormData({...formData, dob: e.target.value})} />
          </div>

          <div className="form-group">
            <label className="form-label">Nickname</label>
            <input type="text" className="form-input" placeholder="Your nickname (optional)" value={formData.nickname} onChange={(e) => setFormData({...formData, nickname: e.target.value})} />
          </div>

          <div className="form-group">
            <label className="form-label">About Me</label>
            <textarea className="form-textarea" rows={3} placeholder="Tell us about yourself (optional)" value={formData.aboutMe} onChange={(e) => setFormData({...formData, aboutMe: e.target.value})} />
          </div>

          <div className="form-group">
            <label className="form-label">Avatar</label>
            <input type="file" accept="image/*" className="file-input" onChange={(e) => setFormData({...formData, avatar: e.target.files ? e.target.files[0] : null})} />
          </div>

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