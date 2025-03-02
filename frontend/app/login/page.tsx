import Link from 'next/link'
import './login.css'

export default function Login() {
  return (
    <div className="login-container">
      <div className="login-form-wrapper">
        <div className="nav-buttons">
          <Link href="/">
            <button className="back-button">Back to Home</button>
          </Link>
        </div>

        <h2 className="form-title">Login</h2>
        
        <form>
          <div className="form-group">
            <label className="form-label">Email</label>
            <input 
              type="email" 
              required 
              className="form-input" 
              placeholder="Enter your email"
            />
          </div>

          <div className="form-group">
            <label className="form-label">Password</label>
            <input 
              type="password" 
              required 
              className="form-input" 
              placeholder="Enter your password"
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
