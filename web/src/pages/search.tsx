import { useState, useEffect, useRef } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { api, Track, Album, Artist } from '../api/client'
import { TrackRow } from '../components/track-row'

function AlbumCard({ album }: { album: Album }) {
  return (
    <Link to={`/album/${album.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div style={{
        padding: '0.75rem',
        borderRadius: '6px',
        background: 'rgba(255,255,255,0.04)',
        cursor: 'pointer',
        transition: 'background 0.15s',
      }}
        onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.08)'}
        onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.04)'}
      >
        <div style={{ width: '100%', aspectRatio: '1', background: '#333', borderRadius: '4px', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '2rem' }}>🎵</div>
        <div style={{ fontWeight: 600, fontSize: '0.9rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{album.title}</div>
        <div style={{ fontSize: '0.8rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{album.artist}{album.year ? ` · ${album.year}` : ''}</div>
      </div>
    </Link>
  )
}

function ArtistItem({ artist }: { artist: Artist }) {
  return (
    <Link to={`/artist/${artist.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div style={{
        padding: '0.5rem 0.75rem',
        borderRadius: '4px',
        background: 'rgba(255,255,255,0.04)',
        cursor: 'pointer',
        fontWeight: 500,
      }}
        onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.08)'}
        onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.04)'}
      >
        {artist.name}
      </div>
    </Link>
  )
}

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get('q') ?? '')
  const [tracks, setTracks] = useState<Track[]>([])
  const [albums, setAlbums] = useState<Album[]>([])
  const [artists, setArtists] = useState<Artist[]>([])
  const [loading, setLoading] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  function runSearch(q: string) {
    if (!q.trim()) {
      setTracks([]); setAlbums([]); setArtists([])
      return
    }
    setLoading(true)
    api.search(q)
      .then(r => { setTracks(r.tracks); setAlbums(r.albums); setArtists(r.artists) })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    const q = searchParams.get('q') ?? ''
    setQuery(q)
    runSearch(q)
  }, [searchParams])

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const v = e.target.value
    setQuery(v)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      setSearchParams(v ? { q: v } : {})
    }, 300)
  }

  const hasResults = tracks.length > 0 || albums.length > 0 || artists.length > 0

  return (
    <div style={{ padding: '1.5rem' }}>
      <input
        autoFocus
        type="search"
        placeholder="Search tracks, albums, artists…"
        value={query}
        onChange={handleChange}
        style={{
          width: '100%',
          maxWidth: '500px',
          padding: '0.6rem 1rem',
          borderRadius: '20px',
          border: '1px solid #444',
          background: '#222',
          color: '#fff',
          fontSize: '1rem',
          outline: 'none',
          marginBottom: '2rem',
        }}
      />

      {loading && <p style={{ color: '#999' }}>Searching…</p>}

      {!loading && query && !hasResults && (
        <p style={{ color: '#999' }}>No results for "{query}"</p>
      )}

      {artists.length > 0 && (
        <section style={{ marginBottom: '2rem' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Artists</h2>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
            {artists.map(a => <ArtistItem key={a.id} artist={a} />)}
          </div>
        </section>
      )}

      {albums.length > 0 && (
        <section style={{ marginBottom: '2rem' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Albums</h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '1rem' }}>
            {albums.map(a => <AlbumCard key={a.id} album={a} />)}
          </div>
        </section>
      )}

      {tracks.length > 0 && (
        <section>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Tracks</h2>
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
        </section>
      )}
    </div>
  )
}
