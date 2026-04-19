import { useEffect, useState } from 'react'
import { FixedSizeList as List } from 'react-window'
import { api, Track } from '../api/client'
import { TrackRow } from '../components/track-row'
import { SkeletonTrackList } from '../components/skeleton-track-list'
import { EmptyState } from '../components/empty-state'

const TRACK_ROW_HEIGHT = 56
const HEADER_OFFSET = 160 // approx header + padding above the list

function useListHeight() {
  const [height, setHeight] = useState(() => window.innerHeight - HEADER_OFFSET)
  useEffect(() => {
    const update = () => setHeight(window.innerHeight - HEADER_OFFSET)
    window.addEventListener('resize', update)
    return () => window.removeEventListener('resize', update)
  }, [])
  return height
}

export function LibraryPage() {
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(true)
  const [scanning, setScanning] = useState(false)
  const [scanMsg, setScanMsg] = useState('')
  const listHeight = useListHeight()

  useEffect(() => {
    api.tracks(5000, 0)
      .then(r => setTracks(r.tracks ?? []))
      .finally(() => setLoading(false))
  }, [])

  async function handleScan() {
    setScanning(true)
    setScanMsg('Starting scan...')
    await api.startScan()
    const poll = setInterval(async () => {
      const s = await api.scanStatus()
      setScanMsg(`Scanning: ${s.scanned}/${s.total}`)
      if (s.done) {
        clearInterval(poll)
        setScanning(false)
        setScanMsg(`Done — ${s.scanned} tracks`)
        const r = await api.tracks(5000, 0)
        setTracks(r.tracks ?? [])
      }
    }, 1000)
  }

  return (
    <div style={{ padding: '1.5rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1.5rem' }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Library</h1>
        <button
          onClick={handleScan}
          disabled={scanning}
          style={{
            padding: '0.4rem 1rem',
            borderRadius: '20px',
            border: 'none',
            background: '#1db954',
            color: '#000',
            fontWeight: 600,
            cursor: scanning ? 'default' : 'pointer',
            opacity: scanning ? 0.6 : 1,
          }}
        >
          {scanning ? 'Scanning...' : 'Scan'}
        </button>
        {scanMsg && <span style={{ color: '#999', fontSize: '0.85rem' }}>{scanMsg}</span>}
      </div>

      {loading ? (
        <SkeletonTrackList rows={12} />
      ) : tracks.length === 0 ? (
        <EmptyState
          icon="🎵"
          title="No tracks found"
          subtitle='Point --music-dir at your library and click Scan'
        />
      ) : (
        <>
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
            <span>#</span>
            <span>Title / Artist</span>
            <span>Album</span>
            <span>Duration</span>
          </div>
          <List
            height={listHeight}
            itemCount={tracks.length}
            itemSize={TRACK_ROW_HEIGHT}
            width="100%"
          >
            {({ index, style }) => (
              <div style={style}>
                <TrackRow track={tracks[index]} queue={tracks} index={index} />
              </div>
            )}
          </List>
        </>
      )}
    </div>
  )
}
