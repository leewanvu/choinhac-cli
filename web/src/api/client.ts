export interface Track {
  id: number
  title: string
  album_id: number
  album: string
  artist_id: number
  artist: string
  track_num: number
  duration: number
  format: string
  file_path: string
}

export interface Album {
  id: number
  title: string
  artist_id: number
  artist: string
  year: number
  cover_path: string
}

export interface Artist {
  id: number
  name: string
}

export interface Playlist {
  id: number
  name: string
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, options)
  if (!res.ok) {
    const body = await res.text()
    throw new Error(`${res.status} ${body}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  tracks: (limit = 50, offset = 0) =>
    request<{ tracks: Track[] }>(`/api/library/tracks?limit=${limit}&offset=${offset}`),

  albums: () => request<{ albums: Album[] }>('/api/library/albums'),

  artists: () => request<{ artists: Artist[] }>('/api/library/artists'),

  search: (q: string) =>
    request<{ tracks: Track[]; albums: Album[]; artists: Artist[] }>(`/api/search?q=${encodeURIComponent(q)}`),

  albumDetail: (id: number) =>
    request<{ album: Album; tracks: Track[] }>(`/api/library/albums/${id}`),

  artistDetail: (id: number) =>
    request<{ artist: Artist; albums: Album[] }>(`/api/library/artists/${id}`),

  startScan: () =>
    request<{ started: boolean }>('/api/scan', { method: 'POST' }),

  scanStatus: () =>
    request<{ running: boolean; scanned: number; total: number; done: boolean }>('/api/scan/status'),

  playlists: () => request<{ playlists: Playlist[] }>('/api/playlists'),

  createPlaylist: (name: string) =>
    request<{ id: number; name: string }>('/api/playlists', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    }),

  deletePlaylist: (id: number) =>
    fetch(`/api/playlists/${id}`, { method: 'DELETE' }),

  playlistTracks: (id: number) =>
    request<{ tracks: Track[] }>(`/api/playlists/${id}/tracks`),

  addToPlaylist: (playlistId: number, trackId: number) =>
    request<{ ok: boolean }>(`/api/playlists/${playlistId}/tracks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ track_id: trackId }),
    }),

  renamePlaylist: (id: number, name: string) =>
    request<{ ok: boolean }>(`/api/playlists/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    }),

  removeFromPlaylist: (playlistId: number, trackId: number) =>
    fetch(`/api/playlists/${playlistId}/tracks/${trackId}`, { method: 'DELETE' }),

  reorderPlaylist: (playlistId: number, trackIds: number[]) =>
    request<{ ok: boolean }>(`/api/playlists/${playlistId}/reorder`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ track_ids: trackIds }),
    }),
}
