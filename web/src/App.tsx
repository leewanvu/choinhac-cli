import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { LibraryPage } from './pages/library'
import { SearchPage } from './pages/search'
import { AlbumDetailPage } from './pages/album-detail'
import { ArtistDetailPage } from './pages/artist-detail'
import { PlaylistDetailPage } from './pages/playlist-detail'
import { NowPlayingBar } from './components/now-playing-bar'
import { PlaylistSidebar } from './components/playlist-sidebar'
import { QueueDrawer } from './components/queue-drawer'
import { AddToPlaylistDialog } from './components/add-to-playlist-dialog'
import { usePlaylists } from './hooks/use-playlists'
import { useKeyboardShortcuts } from './hooks/use-keyboard-shortcuts'

const MOBILE_BP = 768

function useIsMobile() {
  const [isMobile, setIsMobile] = useState(() => window.innerWidth < MOBILE_BP)
  useEffect(() => {
    const mq = window.matchMedia(`(max-width: ${MOBILE_BP - 1}px)`)
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])
  return isMobile
}

function AppShell() {
  useKeyboardShortcuts()
  const { playlists, create, remove } = usePlaylists()
  const isMobile = useIsMobile()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const sidebarContent = (
    <>
      <div style={{ fontWeight: 700, fontSize: '1.1rem', marginBottom: '1.5rem', color: '#1db954', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        musicweb
        {isMobile && (
          <button onClick={() => setSidebarOpen(false)} style={closeBtnStyle}>✕</button>
        )}
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
        <NavLink to="/" end style={navStyle} onClick={() => isMobile && setSidebarOpen(false)}>Library</NavLink>
        <NavLink to="/search" style={navStyle} onClick={() => isMobile && setSidebarOpen(false)}>Search</NavLink>
      </div>
      <PlaylistSidebar playlists={playlists} onCreate={create} onDelete={remove} />
    </>
  )

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
      {/* Mobile overlay backdrop */}
      {isMobile && sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', zIndex: 199 }}
        />
      )}

      {/* Sidebar — fixed overlay on mobile, static on desktop */}
      {(!isMobile || sidebarOpen) && (
        <nav style={{
          width: '220px',
          flexShrink: 0,
          background: '#000',
          padding: '1.5rem 1rem',
          display: 'flex',
          flexDirection: 'column',
          borderRight: '1px solid #282828',
          overflowY: 'auto',
          ...(isMobile ? { position: 'fixed', top: 0, left: 0, bottom: 0, zIndex: 200 } : {}),
        }}>
          {sidebarContent}
        </nav>
      )}

      {/* Main content */}
      <main style={{ flex: 1, overflowY: 'auto', paddingBottom: '80px' }}>
        {isMobile && (
          <button
            onClick={() => setSidebarOpen(true)}
            aria-label="Open menu"
            style={{ margin: '0.75rem', padding: '0.4rem 0.75rem', background: 'none', border: '1px solid #333', borderRadius: '4px', color: '#fff', cursor: 'pointer', fontSize: '1rem' }}
          >☰</button>
        )}
        <Routes>
          <Route path="/" element={<LibraryPage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/album/:id" element={<AlbumDetailPage />} />
          <Route path="/artist/:id" element={<ArtistDetailPage />} />
          <Route path="/playlist/:id" element={<PlaylistDetailPage />} />
        </Routes>
      </main>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppShell />
      <NowPlayingBar />
      <QueueDrawer />
      <AddToPlaylistDialog />
    </BrowserRouter>
  )
}

function navStyle({ isActive }: { isActive: boolean }): React.CSSProperties {
  return {
    display: 'block',
    padding: '0.5rem 0.75rem',
    borderRadius: '4px',
    textDecoration: 'none',
    fontWeight: 500,
    color: isActive ? '#fff' : '#b3b3b3',
    background: isActive ? 'rgba(255,255,255,0.1)' : 'transparent',
  }
}

const closeBtnStyle: React.CSSProperties = {
  background: 'none',
  border: 'none',
  color: '#999',
  cursor: 'pointer',
  fontSize: '1rem',
  padding: '0.1rem',
}
