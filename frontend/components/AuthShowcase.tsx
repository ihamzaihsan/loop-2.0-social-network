import BrandMark from './BrandMark'

type AuthShowcaseProps = {
  mode: 'login' | 'register'
}

export default function AuthShowcase({ mode }: AuthShowcaseProps) {
  const isLogin = mode === 'login'

  return (
    <section className="auth-showcase" aria-label="About Loop">
      <div className="auth-showcase__noise" aria-hidden="true"></div>
      <BrandMark inverse />

      <div className="auth-showcase__copy">
        <span className="eyebrow">Your people. Your pace.</span>
        <h1>{isLogin ? 'Come back to what matters.' : 'Make the internet feel personal again.'}</h1>
        <p>
          A thoughtful place to share the everyday, find your people, and keep
          conversations moving without the noise.
        </p>
      </div>

      <div className="auth-showcase__preview" aria-hidden="true">
        <div className="preview-orbit preview-orbit--one"></div>
        <div className="preview-orbit preview-orbit--two"></div>
        <div className="preview-card preview-card--primary">
          <div className="preview-avatar">A</div>
          <div>
            <strong>Small moments, shared.</strong>
            <span>Keep up with the people you care about.</span>
          </div>
        </div>
        <div className="preview-card preview-card--float">
          <span className="preview-pulse"></span>
          <span>12 friends online</span>
        </div>
      </div>

      <div className="auth-showcase__footer">
        <span>Private by design</span>
        <span>Real-time conversations</span>
        <span>Built for community</span>
      </div>
    </section>
  )
}
