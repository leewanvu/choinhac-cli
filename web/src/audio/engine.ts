type EngineEvent = 'timeupdate' | 'ended' | 'play' | 'pause' | 'error'
type Handler = (data?: number) => void

class AudioEngine {
  private el: HTMLAudioElement
  private listeners: Map<EngineEvent, Set<Handler>> = new Map()

  constructor() {
    this.el = new Audio()
    this.el.preload = 'metadata'

    this.el.addEventListener('timeupdate', () => {
      this.emit('timeupdate', this.el.currentTime)
    })
    this.el.addEventListener('ended', () => this.emit('ended'))
    this.el.addEventListener('play', () => this.emit('play'))
    this.el.addEventListener('pause', () => this.emit('pause'))
    this.el.addEventListener('error', () => this.emit('error'))
  }

  play(streamUrl: string) {
    if (this.el.src !== window.location.origin + streamUrl) {
      this.el.src = streamUrl
    }
    this.el.play()
  }

  pause() { this.el.pause() }

  toggle() {
    if (this.el.paused) {
      this.el.play()
    } else {
      this.el.pause()
    }
  }

  seek(seconds: number) { this.el.currentTime = seconds }

  setVolume(v: number) { this.el.volume = Math.max(0, Math.min(1, v)) }

  get currentTime() { return this.el.currentTime }
  get duration() { return this.el.duration || 0 }
  get paused() { return this.el.paused }
  get volume() { return this.el.volume }

  on(event: EngineEvent, handler: Handler) {
    if (!this.listeners.has(event)) this.listeners.set(event, new Set())
    this.listeners.get(event)!.add(handler)
    return () => this.listeners.get(event)!.delete(handler)
  }

  private emit(event: EngineEvent, data?: number) {
    this.listeners.get(event)?.forEach(h => h(data))
  }
}

export const engine = new AudioEngine()
