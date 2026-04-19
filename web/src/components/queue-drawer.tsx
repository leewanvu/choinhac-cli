import { usePlayer } from '../store/player'
import { useUI } from '../store/ui'

function fmtDuration(s: number): string {
  const m = Math.floor(s / 60)
  const sec = String(s % 60).padStart(2, '0')
  return `${m}:${sec}`
}

export function QueueDrawer() {
  const { queueDrawerOpen, toggleQueueDrawer } = useUI()
  const { queue, currentIndex, playTrack } = usePlayer()

  if (!queueDrawerOpen) return null

  const nowPlaying = queue[currentIndex] ?? null
  const upcoming = queue.slice(currentIndex + 1)

  return (
    <>
      <div onClick={toggleQueueDrawer} style={{ position: 'fixed', inset: 0, zIndex: 150 }} />
      <div style={{
        position: 'fixed', top: 0, right: 0, bottom: '80px',
        width: '320px', background: '#121212', borderLeft: '1px solid #282828',
        zIndex: 151, display: 'flex', flexDirection: 'column', overflowY: 'auto',
      }}>
        <div style={{ padding: '1rem', borderBottom: '1px solid #282828', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h3 style={{ fontWeight: 700, fontSize: '0.95rem' }}>Queue</h3>
          <button onClick={toggleQueueDrawer} style={{ background: 'none', border: 'none', color: '#999', cursor: 'pointer', fontSize: '1.1rem' }}>✕</button>
        </div>

        {nowPlaying && (
          <div style={{ padding: '0.75rem 1rem' }}>
            <div style={{ fontSize: '0.7rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.5rem' }}>Now playing</div>
            <div style={{ padding: '0.5rem 0.75rem', borderRadius: '4px', background: 'rgba(29,185,84,0.1)', borderLeft: '2px solid #1db954' }}>
              <div style={{ fontWeight: 600, fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: '#1db954' }}>{nowPlaying.title}</div>
              <div style={{ fontSize: '0.8rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{nowPlaying.artist}</div>
            </div>
          </div>
        )}

        {upcoming.length > 0 && (
          <div style={{ padding: '0 1rem 1rem' }}>
            <div style={{ fontSize: '0.7rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.5rem' }}>Next up</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.15rem' }}>
              {upcoming.map((t, i) => (
                <div
                  key={`${t.id}-${i}`}
                  onDoubleClick={() => playTrack(t, queue)}
                  style={{ padding: '0.5rem 0.75rem', borderRadius: '4px', cursor: 'pointer', display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '0.5rem' }}
                  onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.06)'}
                  onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'transparent'}
                >
                  <div style={{ overflow: 'hidden' }}>
                    <div style={{ fontWeight: 500, fontSize: '0.85rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.title}</div>
                    <div style={{ fontSize: '0.75rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.artist}</div>
                  </div>
                  <span style={{ fontSize: '0.75rem', color: '#999', flexShrink: 0 }}>{fmtDuration(t.duration)}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {upcoming.length === 0 && !nowPlaying && (
          <p style={{ padding: '1rem', color: '#999', fontSize: '0.9rem' }}>Queue is empty.</p>
        )}
      </div>
    </>
  )
}
