import { usePlayer } from '../store/player'
import { useUI } from '../store/ui'
import { engine } from '../audio/engine'
import { useEffect, useState } from 'react'
import { CoverImage } from './cover-image'

function fmtTime(s: number): string {
  if (!isFinite(s)) return '0:00'
  const m = Math.floor(s / 60)
  const sec = String(Math.floor(s % 60)).padStart(2, '0')
  return `${m}:${sec}`
}

export function NowPlayingBar() {
  const { currentTrack, isPlaying, volume, progress, togglePlay, next, prev, seek, setVolume } = usePlayer()
  const { toggleQueueDrawer, queueDrawerOpen } = useUI()
  const [duration, setDuration] = useState(0)

  useEffect(() => {
    const off = engine.on('timeupdate', () => setDuration(engine.duration))
    return () => { off() }
  }, [])

  if (!currentTrack) return null

  return (
    <div style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      height: '80px',
      background: '#181818',
      borderTop: '1px solid #282828',
      display: 'flex',
      alignItems: 'center',
      padding: '0 1rem',
      gap: '1rem',
      zIndex: 100,
    }}>
      {/* Track info with cover */}
      <div style={{ flex: '0 0 220px', overflow: 'hidden', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <CoverImage albumId={currentTrack.album_id} size={48} />
        <div style={{ overflow: 'hidden' }}>
          <div style={{ fontWeight: 600, fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {currentTrack.title}
          </div>
          <div style={{ fontSize: '0.8rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {currentTrack.artist}
          </div>
        </div>
      </div>

      {/* Controls + seek */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.4rem' }}>
        <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
          <button onClick={prev} style={btnStyle}>⏮</button>
          <button onClick={togglePlay} style={{ ...btnStyle, fontSize: '1.4rem', width: '2rem', height: '2rem' }}>
            {isPlaying ? '⏸' : '▶'}
          </button>
          <button onClick={next} style={btnStyle}>⏭</button>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', width: '100%', maxWidth: '500px' }}>
          <span style={{ fontSize: '0.75rem', color: '#999', width: '2.5rem', textAlign: 'right' }}>{fmtTime(progress)}</span>
          <input
            type="range"
            min={0}
            max={duration || 0}
            step={1}
            value={progress}
            onChange={e => seek(Number(e.target.value))}
            style={{ flex: 1, accentColor: '#1db954' }}
          />
          <span style={{ fontSize: '0.75rem', color: '#999', width: '2.5rem' }}>{fmtTime(duration)}</span>
        </div>
      </div>

      {/* Volume + queue toggle */}
      <div style={{ flex: '0 0 200px', display: 'flex', alignItems: 'center', gap: '0.5rem', justifyContent: 'flex-end' }}>
        <span style={{ fontSize: '0.9rem' }}>🔉</span>
        <input
          type="range"
          min={0}
          max={1}
          step={0.02}
          value={volume}
          onChange={e => setVolume(Number(e.target.value))}
          style={{ width: '80px', accentColor: '#1db954' }}
        />
        <button
          onClick={toggleQueueDrawer}
          title="Queue"
          style={{ ...btnStyle, fontSize: '1rem', color: queueDrawerOpen ? '#1db954' : '#fff' }}
        >☰</button>
      </div>
    </div>
  )
}

const btnStyle: React.CSSProperties = {
  background: 'none',
  border: 'none',
  color: '#fff',
  fontSize: '1.1rem',
  cursor: 'pointer',
  padding: '0.25rem',
  lineHeight: 1,
}
