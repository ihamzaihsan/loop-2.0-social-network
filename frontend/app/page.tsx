import Link from 'next/link'

export default function Home() {
  return (
    <div className="home-container">
      <div className="content-wrapper">
        <h1 className="main-title">Welcome to Social Network</h1>
        <p className="subtitle">Connect with friends and share your moments</p>
        
        <div className="button-group">
          <Link href="/login">
            <button className="primary-button">Login</button>
          </Link>
          <Link href="/register">
            <button className="secondary-button">Register</button>
          </Link>
        </div>
      </div>
    </div>
  );
}
