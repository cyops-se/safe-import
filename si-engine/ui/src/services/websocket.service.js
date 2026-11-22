/**
 * Server push over websocket. The connection is authenticated with the JWT
 * (passed as a query parameter since browsers can not set websocket headers)
 * and is automatically re-established with back-off when it is lost.
 */
import JwtService from '@/services/jwt.service'

const MAX_RETRY_DELAY = 30000

const WebsocketService = {
  connection: null,
  subscriptions: {},
  retryDelay: 1000,
  retryTimer: null,
  stopped: false,

  url () {
    const scheme = window.location.protocol === 'https:' ? 'wss://' : 'ws://'
    const token = encodeURIComponent(JwtService.getToken() || '')
    return scheme + window.location.host + '/ws?token=' + token
  },

  connect () {
    if (!JwtService.getToken()) return
    if (this.connection && this.connection.readyState <= WebSocket.OPEN) return

    this.stopped = false
    this.connection = new WebSocket(this.url())
    this.connection.onmessage = event => this.onmessage(event)
    this.connection.onopen = () => {
      this.retryDelay = 1000
    }
    this.connection.onclose = () => this.scheduleReconnect()
  },

  disconnect () {
    this.stopped = true
    clearTimeout(this.retryTimer)
    if (this.connection) this.connection.close()
    this.connection = null
  },

  scheduleReconnect () {
    if (this.stopped) return
    clearTimeout(this.retryTimer)
    this.retryTimer = setTimeout(() => this.connect(), this.retryDelay)
    this.retryDelay = Math.min(this.retryDelay * 2, MAX_RETRY_DELAY)
  },

  onmessage (event) {
    if (!event || !event.data) return
    let data
    try {
      data = JSON.parse(event.data)
    } catch (e) {
      return
    }

    if (!data || !data.topic || !data.data) return
    const subs = this.subscriptions[data.topic]
    if (subs) {
      for (let i = 0; i < subs.length; i++) {
        subs[i](data.topic, data.data.message)
      }
    }
  },

  topic (name, callback) {
    if (!this.subscriptions[name]) this.subscriptions[name] = []
    this.subscriptions[name].push(callback)
  },
}

export default WebsocketService
