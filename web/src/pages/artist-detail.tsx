import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, Artist, Album } from '../api/client'

function AlbumCard({ album }: { album: Album }) {
  return (
    <Link to={`/album/${album.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div
        style={{
          padding: '0.75rem',
          borderRadius: '6px',
          background: 'rgba(255,255,255,0.04)',
          cursor: 'pointer',
        }}
        onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.08)'}
        onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.04)'}
      >
        <div style={{
          width: '100%', aspectRatio: '1', background: '#333', borderRadius: '4px',
          marginBottom: '0.5rem', display: 'flex', alignItems: 'center',
          justifyContent: 'center', fontSize: '2rem',
        }}>🎵</div>
        <div style={{ fontWeight: 600, fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{album.title}</div>
        <div style={{ fontSize: '0.8rem', color: '#999' }}>{album.year || 'Album'}</div>
      </div>
    </Link>
  )
}

export function ArtistDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [artist, setArtist] = useState<Artist | null>(null)
  const [albums, setAlbums] = useState<Album[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    api.artistDetail(Number(id))
      .then(r => { setArtist(r.artist); setAlbums(r.albums) })
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div style={{ padding: '2rem', color: '#999' }}>Loading…</div>
  if (!artist) return <div style={{ padding: '2rem', color: '#999' }}>Artist not found.</div>

  return (
    <div style={{ padding: '1.5rem' }}>
      {/* Header */}
      <div style={{ marginBottom: '2rem' }}>
        <div style={{ fontSize: '0.75rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.25rem' }}>Artist</div>
        <h1 style={{ fontSize: '2.5rem', fontWeight: 700 }}>{artist.name}</h1>
        <div style={{ color: '#999', fontSize: '0.9rem', marginTop: '0.25rem' }}>
          {albums.length} album{albums.length !== 1 ? 's' : ''}
        </div>
      </div>

      {/* Albums grid */}
      {albums.length === 0 ? (
        <p style={{ color: '#999' }}>No albums found.</p>
      ) : (
        <>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '1rem' }}>Albums</h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: '1rem' }}>
            {albums.map(a => <AlbumCard key={a.id} album={a} />)}
          </div>
        </>
      )}
    </div>
  )
}
