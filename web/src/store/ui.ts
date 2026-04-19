import { create } from 'zustand'

interface UIState {
  queueDrawerOpen: boolean
  addToPlaylistTrackId: number | null
  toggleQueueDrawer: () => void
  openAddToPlaylist: (trackId: number) => void
  closeAddToPlaylist: () => void
}

export const useUI = create<UIState>((set) => ({
  queueDrawerOpen: false,
  addToPlaylistTrackId: null,
  toggleQueueDrawer: () => set(s => ({ queueDrawerOpen: !s.queueDrawerOpen })),
  openAddToPlaylist: (trackId) => set({ addToPlaylistTrackId: trackId }),
  closeAddToPlaylist: () => set({ addToPlaylistTrackId: null }),
}))
