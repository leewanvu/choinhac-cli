import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { Playlist } from '../api/client'

interface Props {
  playlists: Playlist[]
  onCreate: (name: string) => void
  onDelete: (id: number) => void
}

export function PlaylistSidebar({ playlists, onCreate, onDelete }: Props) {
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')

  function submit() {
    const name = newName.trim()
    if (name) onCreate(name)
    setNewName('')
    setCreating(false)
  }

  return (
    <div style={{ marginTop: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem', padding: '0 0.75rem' }}>
        <span style={{ fontSize: '0.7rem', color: '#999', textTransform: 'uppercase', letterSpacing: '0.08em' }}>Playlists</span>
        <button
          onClick={() => setCreating(true)}
          title="New playlist"
          style={{ background: 'none', border: 'none', color: '#999', cursor: 'pointer', fontSize: '1.1rem', lineHeight: 1, padding: 0 }}
        >+</button>
      </div>

      {creating && (
        <div style={{ padding: '0.25rem 0.75rem', marginBottom: '0.25rem' }}>
          <input
            autoFocus
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') submit(); if (e.key === 'Escape') setCreating(false) }}
            onBlur={() => { if (!newName.trim()) setCreating(false) }}
            placeholder="Playlist name"
            style={{ width: '100%', padding: '0.3rem 0.5rem', borderRadius: '4px', border: '1px solid #444', background: '#333', color: '#fff', fontSize: '0.85rem', outline: 'none', boxSizing: 'border-box' }}
          />
        </div>
      )}

      {playlists.map(p => (
        <div
          key={p.id}
          style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}
          onMouseEnter={e => (e.currentTarget.querySelector('.del-btn') as HTMLElement | null)?.style.setProperty('opacity', '1')}
          onMouseLeave={e => (e.currentTarget.querySelector('.del-btn') as HTMLElement | null)?.style.setProperty('opacity', '0')}
        >
          <NavLink
            to={`/playlist/${p.id}`}
            style={({ isActive }) => ({
              flex: 1, display: 'block', padding: '0.5rem 0.75rem', borderRadius: '4px',
              textDecoration: 'none', fontWeight: 500, fontSize: '0.9rem',
              color: isActive ? '#fff' : '#b3b3b3',
              background: isActive ? 'rgba(255,255,255,0.1)' : 'transparent',
              overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            })}
          >{p.name}</NavLink>
          <button
            className="del-btn"
            onClick={() => onDelete(p.id)}
            title="Delete playlist"
            style={{ opacity: 0, background: 'none', border: 'none', color: '#666', cursor: 'pointer', fontSize: '0.85rem', padding: '0 0.25rem', flexShrink: 0 }}
          >✕</button>
        </div>
      ))}
    </div>
  )
}
