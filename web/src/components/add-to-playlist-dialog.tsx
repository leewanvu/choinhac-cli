import { useEffect, useState } from 'react'
import { api, Playlist } from '../api/client'
import { useUI } from '../store/ui'

export function AddToPlaylistDialog() {
  const { addToPlaylistTrackId, closeAddToPlaylist } = useUI()
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [done, setDone] = useState<number | null>(null)

  useEffect(() => {
    if (addToPlaylistTrackId !== null) {
      api.playlists().then(r => setPlaylists(r.playlists))
      setDone(null)
    }
  }, [addToPlaylistTrackId])

  if (addToPlaylistTrackId === null) return null

  async function add(playlistId: number) {
    await api.addToPlaylist(playlistId, addToPlaylistTrackId!)
    setDone(playlistId)
    setTimeout(closeAddToPlaylist, 600)
  }

  return (
    <div
      onClick={closeAddToPlaylist}
      style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', zIndex: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
    >
      <div onClick={e => e.stopPropagation()} style={{ background: '#282828', borderRadius: '8px', padding: '1.5rem', minWidth: '280px', maxWidth: '360px', width: '90%' }}>
        <h3 style={{ fontWeight: 700, marginBottom: '1rem', fontSize: '1rem' }}>Add to playlist</h3>
        {playlists.length === 0 ? (
          <p style={{ color: '#999', fontSize: '0.9rem' }}>No playlists yet.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', maxHeight: '260px', overflowY: 'auto' }}>
            {playlists.map(p => (
              <button
                key={p.id}
                onClick={() => add(p.id)}
                style={{
                  textAlign: 'left', background: done === p.id ? 'rgba(29,185,84,0.2)' : 'transparent',
                  border: 'none', color: done === p.id ? '#1db954' : '#fff',
                  padding: '0.5rem 0.75rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.9rem',
                }}
                onMouseEnter={e => { if (done !== p.id) (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.08)' }}
                onMouseLeave={e => { if (done !== p.id) (e.currentTarget as HTMLButtonElement).style.background = 'transparent' }}
              >
                {done === p.id ? '✓ ' : ''}{p.name}
              </button>
            ))}
          </div>
        )}
        <button
          onClick={closeAddToPlaylist}
          style={{ marginTop: '1rem', width: '100%', padding: '0.4rem', background: 'transparent', border: '1px solid #555', color: '#999', borderRadius: '4px', cursor: 'pointer', fontSize: '0.85rem' }}
        >Cancel</button>
      </div>
    </div>
  )
}
