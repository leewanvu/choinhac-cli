import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import {
  DndContext, closestCenter, KeyboardSensor, PointerSensor,
  useSensor, useSensors, DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy,
  useSortable, arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { api, Track } from '../api/client'
import { usePlayer } from '../store/player'
import { useUI } from '../store/ui'

function fmtDuration(s: number): string {
  const m = Math.floor(s / 60)
  const sec = String(s % 60).padStart(2, '0')
  return `${m}:${sec}`
}

function SortableTrack({ track, index, queue, onRemove }: { track: Track; index: number; queue: Track[]; onRemove: () => void }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: track.id })
  const { currentTrack, isPlaying, playTrack, togglePlay } = usePlayer()
  const { openAddToPlaylist } = useUI()
  const isActive = currentTrack?.id === track.id

  function handleClick() {
    if (isActive) togglePlay(); else playTrack(track, queue)
  }

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.4 : 1,
        display: 'grid',
        gridTemplateColumns: '1.5rem 2rem 1fr 1fr auto auto',
        gap: '0.5rem',
        alignItems: 'center',
        padding: '0.5rem 1rem',
        borderRadius: '4px',
        background: isActive ? 'rgba(255,255,255,0.08)' : 'transparent',
        color: isActive ? '#1db954' : '#fff',
        cursor: 'pointer',
      }}
      onMouseEnter={e => { if (!isActive) (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.04)' }}
      onMouseLeave={e => { if (!isActive) (e.currentTarget as HTMLDivElement).style.background = 'transparent' }}
    >
      <span {...attributes} {...listeners} style={{ cursor: 'grab', color: '#555', fontSize: '0.9rem', textAlign: 'center' }}>⠿</span>
      <span onClick={handleClick} style={{ textAlign: 'right', color: '#999', fontSize: '0.85rem' }}>
        {isActive && isPlaying ? '▶' : index + 1}
      </span>
      <div onClick={handleClick} style={{ overflow: 'hidden' }}>
        <div style={{ fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{track.title}</div>
        <div style={{ fontSize: '0.8rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{track.artist}</div>
      </div>
      <div onClick={handleClick} style={{ fontSize: '0.85rem', color: '#999', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{track.album}</div>
      <div onClick={handleClick} style={{ fontSize: '0.85rem', color: '#999' }}>{fmtDuration(track.duration)}</div>
      <div style={{ display: 'flex', gap: '0.25rem' }}>
        <button onClick={() => openAddToPlaylist(track.id)} title="Add to playlist" style={iconBtn}>+</button>
        <button onClick={onRemove} title="Remove" style={iconBtn}>✕</button>
      </div>
    </div>
  )
}

const iconBtn: React.CSSProperties = {
  background: 'none', border: 'none', color: '#666', cursor: 'pointer',
  fontSize: '0.85rem', padding: '0.2rem 0.3rem', borderRadius: '3px',
}

export function PlaylistDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [name, setName] = useState('')
  const [tracks, setTracks] = useState<Track[]>([])
  const [loading, setLoading] = useState(true)

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  const load = useCallback(() => {
    if (!id) return
    const numId = Number(id)
    Promise.all([api.playlists(), api.playlistTracks(numId)])
      .then(([pl, tr]) => {
        setName(pl.playlists.find(p => p.id === numId)?.name ?? '')
        setTracks(tr.tracks)
      })
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => { load() }, [load])

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIdx = tracks.findIndex(t => t.id === active.id)
    const newIdx = tracks.findIndex(t => t.id === over.id)
    const reordered = arrayMove(tracks, oldIdx, newIdx)
    setTracks(reordered)
    await api.reorderPlaylist(Number(id), reordered.map(t => t.id))
  }

  async function removeTrack(trackId: number) {
    await api.removeFromPlaylist(Number(id), trackId)
    setTracks(prev => prev.filter(t => t.id !== trackId))
  }

  if (loading) return <div style={{ padding: '2rem', color: '#999' }}>Loading…</div>

  return (
    <div style={{ padding: '1.5rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <div style={{ fontSize: '0.75rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.25rem' }}>Playlist</div>
        <h1 style={{ fontSize: '2rem', fontWeight: 700 }}>{name}</h1>
        <div style={{ color: '#999', fontSize: '0.9rem', marginTop: '0.25rem' }}>{tracks.length} track{tracks.length !== 1 ? 's' : ''}</div>
      </div>

      {tracks.length === 0 ? (
        <p style={{ color: '#999' }}>No tracks yet. Add some from the library.</p>
      ) : (
        <>
          <div style={{
            display: 'grid', gridTemplateColumns: '1.5rem 2rem 1fr 1fr auto auto',
            gap: '0.5rem', padding: '0.25rem 1rem',
            color: '#999', fontSize: '0.8rem', borderBottom: '1px solid #333',
            marginBottom: '0.25rem', textTransform: 'uppercase', letterSpacing: '0.05em',
          }}>
            <span />
            <span>#</span><span>Title / Artist</span><span>Album</span><span>Duration</span><span />
          </div>
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={tracks.map(t => t.id)} strategy={verticalListSortingStrategy}>
              {tracks.map((t, i) => (
                <SortableTrack key={t.id} track={t} index={i} queue={tracks} onRemove={() => removeTrack(t.id)} />
              ))}
            </SortableContext>
          </DndContext>
        </>
      )}
    </div>
  )
}
