import { Track } from '../api/client'
import { usePlayer } from '../store/player'
import { useUI } from '../store/ui'

interface Props {
  track: Track
  queue: Track[]
  index: number
}

function fmtDuration(s: number): string {
  const m = Math.floor(s / 60)
  const sec = String(s % 60).padStart(2, '0')
  return `${m}:${sec}`
}

export function TrackRow({ track, queue, index }: Props) {
  const { currentTrack, isPlaying, playTrack, togglePlay } = usePlayer()
  const { openAddToPlaylist } = useUI()
  const isActive = currentTrack?.id === track.id

  function handleClick() {
    if (isActive) togglePlay(); else playTrack(track, queue)
  }

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '2rem 1fr 1fr auto auto',
        gap: '0.5rem',
        alignItems: 'center',
        padding: '0.5rem 1rem',
        borderRadius: '4px',
        background: isActive ? 'rgba(255,255,255,0.08)' : 'transparent',
        color: isActive ? '#1db954' : '#fff',
      }}
      onMouseEnter={e => { if (!isActive) (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.04)' }}
      onMouseLeave={e => { if (!isActive) (e.currentTarget as HTMLDivElement).style.background = 'transparent' }}
    >
      <span onClick={handleClick} style={{ textAlign: 'right', color: '#999', fontSize: '0.85rem', cursor: 'pointer' }}>
        {isActive && isPlaying ? '▶' : index + 1}
      </span>
      <div onClick={handleClick} style={{ overflow: 'hidden', cursor: 'pointer' }}>
        <div style={{ fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{track.title}</div>
        <div style={{ fontSize: '0.8rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{track.artist}</div>
      </div>
      <div onClick={handleClick} style={{ fontSize: '0.85rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', cursor: 'pointer' }}>{track.album}</div>
      <div onClick={handleClick} style={{ fontSize: '0.85rem', color: '#999', cursor: 'pointer' }}>{fmtDuration(track.duration)}</div>
      <button
        onClick={() => openAddToPlaylist(track.id)}
        title="Add to playlist"
        style={{ background: 'none', border: 'none', color: '#666', cursor: 'pointer', fontSize: '1rem', padding: '0.1rem 0.3rem', borderRadius: '3px', lineHeight: 1 }}
        onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.color = '#fff'}
        onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.color = '#666'}
      >+</button>
    </div>
  )
}
