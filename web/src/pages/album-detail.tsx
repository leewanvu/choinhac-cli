import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, Album, Track } from '../api/client'
import { TrackRow } from '../components/track-row'

export function AlbumDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [album, setAlbum] = useState<Album | null>(null)
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    api.albumDetail(Number(id))
      .then(r => { setAlbum(r.album); setTracks(r.tracks) })
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div style={{ padding: '2rem', color: '#999' }}>Loading…</div>
  if (!album) return <div style={{ padding: '2rem', color: '#999' }}>Album not found.</div>

  return (
    <div style={{ padding: '1.5rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'flex-end', marginBottom: '2rem' }}>
        <div style={{
          width: '160px', height: '160px', flexShrink: 0,
          background: '#333', borderRadius: '6px',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontSize: '3rem',
        }}>🎵</div>
        <div>
          <div style={{ fontSize: '0.75rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.25rem' }}>Album</div>
          <h1 style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '0.5rem' }}>{album.title}</h1>
          <div style={{ fontSize: '0.95rem', color: '#ccc' }}>
            <Link to={`/artist/${album.artist_id}`} style={{ color: '#fff', fontWeight: 600, textDecoration: 'none' }}
              onMouseEnter={e => (e.currentTarget as HTMLAnchorElement).style.textDecoration = 'underline'}
              onMouseLeave={e => (e.currentTarget as HTMLAnchorElement).style.textDecoration = 'none'}
            >{album.artist}</Link>
            {album.year ? <span style={{ color: '#999' }}> · {album.year}</span> : null}
            <span style={{ color: '#999' }}> · {tracks.length} track{tracks.length !== 1 ? 's' : ''}</span>
          </div>
        </div>
      </div>

      {/* Track list */}
      <div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: '2rem 1fr 1fr auto',
          gap: '0.5rem',
          padding: '0.25rem 1rem',
          color: '#999',
          fontSize: '0.8rem',
          borderBottom: '1px solid #333',
          marginBottom: '0.25rem',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
        }}>
          <span>#</span><span>Title / Artist</span><span>Album</span><span>Duration</span>
        </div>
        {tracks.map((t, i) => <TrackRow key={t.id} track={t} queue={tracks} index={i} />)}
      </div>
    </div>
  )
}
