import { useState } from 'react'

interface Props {
  albumId?: number
  size?: number
  style?: React.CSSProperties
}

export function CoverImage({ albumId, size = 48, style }: Props) {
  const [failed, setFailed] = useState(false)

  const base: React.CSSProperties = {
    width: size,
    height: size,
    flexShrink: 0,
    borderRadius: '4px',
    background: '#282828',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: '#555',
    fontSize: size * 0.45,
    ...style,
  }

  if (!albumId || failed) {
    return <div style={base}>♪</div>
  }

  return (
    <img
      src={`/cover/${albumId}`}
      loading="lazy"
      width={size}
      height={size}
      style={{ objectFit: 'cover', borderRadius: '4px', flexShrink: 0, ...style }}
      onError={() => setFailed(true)}
    />
  )
}
