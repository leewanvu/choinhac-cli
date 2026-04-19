import { useState, useEffect, useCallback } from 'react'
import { api, Playlist } from '../api/client'

export function usePlaylists() {
  const [playlists, setPlaylists] = useState<Playlist[]>([])

  const refresh = useCallback(() => {
    api.playlists().then(r => setPlaylists(r.playlists))
  }, [])

  useEffect(() => { refresh() }, [refresh])

  async function create(name: string) {
    await api.createPlaylist(name)
    refresh()
  }

  async function remove(id: number) {
    await api.deletePlaylist(id)
    refresh()
  }

  async function rename(id: number, name: string) {
    await api.renamePlaylist(id, name)
    refresh()
  }

  return { playlists, refresh, create, remove, rename }
}
