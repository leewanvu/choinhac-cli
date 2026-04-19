const KEYFRAMES = `@keyframes skeletonPulse { 0%,100% { opacity:0.35 } 50% { opacity:0.7 } }`

const shimmer: React.CSSProperties = {
  background: '#282828',
  borderRadius: '4px',
  animation: 'skeletonPulse 1.4s ease-in-out infinite',
}

interface Props {
  rows?: number
}

export function SkeletonTrackList({ rows = 8 }: Props) {
  return (
    <>
      <style>{KEYFRAMES}</style>
      {Array.from({ length: rows }, (_, i) => (
        <div
          key={i}
          style={{
            display: 'grid',
            gridTemplateColumns: '2rem 1fr 1fr auto',
            gap: '0.5rem',
            alignItems: 'center',
            padding: '0.75rem 1rem',
          }}
        >
          <div style={{ ...shimmer, width: '1.2rem', height: '0.85rem' }} />
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <div style={{ ...shimmer, width: `${55 + (i * 17) % 35}%`, height: '0.85rem' }} />
            <div style={{ ...shimmer, width: `${25 + (i * 11) % 25}%`, height: '0.7rem' }} />
          </div>
          <div style={{ ...shimmer, width: `${40 + (i * 13) % 30}%`, height: '0.85rem' }} />
          <div style={{ ...shimmer, width: '2.5rem', height: '0.85rem' }} />
        </div>
      ))}
    </>
  )
}
