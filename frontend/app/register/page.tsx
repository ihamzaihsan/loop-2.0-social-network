import './register.css'
import Link from 'next/link'

export default function Register() {
  return (
    <div className="register-container">
      <div className="register-form-wrapper">
        <div className="nav-buttons">
          <Link href="/">
            <button className="back-button">Back to Home</button>
          </Link>
        </div>

        <h2 className="form-title">Create Account</h2>
        
        <form>
          <div className="form-grid">
            <div className="form-group">
              <label className="form-label">First Name *</label>
              <input type="text" required className="form-input" placeholder="Hussain" />
            </div>
            <div className="form-group">
              <label className="form-label">Last Name *</label>
              <input type="text" required className="form-input" placeholder="Ali" />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Email *</label>
            <input type="email" required className="form-input" placeholder="Hussain@example.com" />
          </div>

          <div className="form-group">
            <label className="form-label">Password *</label>
            <input type="password" required className="form-input" placeholder="••••••••" />
          </div>

          <div className="form-group">
            <label className="form-label">Date of Birth *</label>
            <input type="date" required className="form-input" />
          </div>

          <div className="form-group">
            <label className="form-label">Nickname</label>
            <input type="text" className="form-input" placeholder="Your nickname (optional)" />
          </div>

          <div className="form-group">
            <label className="form-label">About Me</label>
            <textarea className="form-textarea" rows={3} placeholder="Tell us about yourself (optional)" />
          </div>

          <div className="form-group">
            <label className="form-label">Avatar</label>
            <input type="file" accept="image/*" className="file-input" />
          </div>

          <button type="submit" className="submit-button">
            Create Account
          </button>

          <div className="login-link">
            Already have an account? <Link href="/login" className="text-link">Login</Link>
          </div>
        </form>
      </div>
    </div>
  );
}