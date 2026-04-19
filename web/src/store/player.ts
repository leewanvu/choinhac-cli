import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { Track } from '../api/client'
import { engine } from '../audio/engine'

interface PlayerState {
  currentTrack: Track | null
  queue: Track[]
  currentIndex: number
  isPlaying: boolean
  volume: number
  progress: number
  duration: number

  playTrack: (track: Track, queue?: Track[]) => void
  togglePlay: () => void
  next: () => void
  prev: () => void
  seek: (seconds: number) => void
  setVolume: (v: number) => void
  setProgress: (p: number) => void
  setDuration: (d: number) => void
}

export const usePlayer = create<PlayerState>()(
  persist(
    (set, get) => ({
      currentTrack: null,
      queue: [],
      currentIndex: -1,
      isPlaying: false,
      volume: 1,
      progress: 0,
      duration: 0,

      playTrack: (track, queue) => {
        const q = queue ?? [track]
        const idx = q.findIndex(t => t.id === track.id)
        set({ currentTrack: track, queue: q, currentIndex: idx, isPlaying: true, progress: 0 })
        engine.play(`/stream/${track.id}`)
        engine.setVolume(get().volume)
      },

      togglePlay: () => {
        engine.toggle()
        set({ isPlaying: !engine.paused })
      },

      next: () => {
        const { queue, currentIndex } = get()
        const nextIdx = currentIndex + 1
        if (nextIdx < queue.length) {
          get().playTrack(queue[nextIdx], queue)
        }
      },

      prev: () => {
        const { queue, currentIndex, progress } = get()
        if (progress > 3) {
          engine.seek(0)
          return
        }
        const prevIdx = currentIndex - 1
        if (prevIdx >= 0) {
          get().playTrack(queue[prevIdx], queue)
        }
      },

      seek: (seconds) => {
        engine.seek(seconds)
        set({ progress: seconds })
      },

      setVolume: (v) => {
        engine.setVolume(v)
        set({ volume: v })
      },

      setProgress: (p) => set({ progress: p }),
      setDuration: (d) => set({ duration: d }),
    }),
    {
      name: 'musicweb-player',
      partialize: (s) => ({ queue: s.queue, currentIndex: s.currentIndex, volume: s.volume }),
    }
  )
)

// Wire engine events to store
engine.on('timeupdate', (t) => usePlayer.getState().setProgress(t ?? 0))
engine.on('ended', () => usePlayer.getState().next())
engine.on('play', () => usePlayer.setState({ isPlaying: true }))
engine.on('pause', () => usePlayer.setState({ isPlaying: false }))
