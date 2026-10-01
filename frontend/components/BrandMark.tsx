type BrandMarkProps = {
  compact?: boolean
  inverse?: boolean
}

export default function BrandMark({ compact = false, inverse = false }: BrandMarkProps) {
  return (
    <div className={`brand-mark ${compact ? 'brand-mark--compact' : ''} ${inverse ? 'brand-mark--inverse' : ''}`}>
      <span className="brand-symbol" aria-hidden="true">
        <span></span>
        <span></span>
      </span>
      <span className="brand-word">loop</span>
      {!compact && <span className="brand-version">2.0</span>}
    </div>
  )
}
