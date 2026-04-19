import { useEffect, useRef } from 'react'
import { usePlayer } from '../store/player'

export function useKeyboardShortcuts() {
  const { togglePlay, next, prev, setVolume } = usePlayer()
  // Use ref to always read latest volume without re-registering the listener
  const volumeRef = useRef(usePlayer.getState().volume)

  useEffect(() => {
    return usePlayer.subscribe(s => { volumeRef.current = s.volume })
  }, [])

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      const tag = (document.activeElement?.tagName ?? '').toLowerCase()
      if (tag === 'input' || tag === 'textarea') return

      switch (e.code) {
        case 'Space':
          e.preventDefault()
          togglePlay()
          break
        case 'ArrowRight':
          e.preventDefault()
          next()
          break
        case 'ArrowLeft':
          e.preventDefault()
          prev()
          break
        case 'ArrowUp':
          e.preventDefault()
          setVolume(Math.min(1, volumeRef.current + 0.1))
          break
        case 'ArrowDown':
          e.preventDefault()
          setVolume(Math.max(0, volumeRef.current - 0.1))
          break
        case 'KeyM':
          e.preventDefault()
          setVolume(volumeRef.current > 0 ? 0 : 1)
          break
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [togglePlay, next, prev, setVolume])
}
