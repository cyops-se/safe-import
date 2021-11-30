/**
 * Service to call HTTP request via Axios
 */
const WebsocketService = {
  baseURL: 'ws://localhost:7499/ws',
  connection: null,
  subscriptions: [],

  /**
   * Set the default HTTP request headers
   */
  connect: function () {
    this.connection = new WebSocket(this.baseURL)
    this.connection.onmessage = this.onmessage
    this.connection.onopen = this.onopen
    this.connection.onclose = this.onclose
    this.connection.subscriptions = this.subscriptions
  },

  onopen: function () {
    console.log('Websocket successfully connected')
  },

  onclose: function () {
    console.log('Websocket  closed: ')
  },

  onmessage: function (event) {
    if (!event || !event.data) return
    var data = JSON.parse(event.data)
    // console.log('DATA: ' + JSON.stringify(data))

    if (!data || !data.topic || !data.data) return
    // var message = JSON.parse(data.message)
    var message = data.data.message

    // if (data.topic !== 'system.heartbeat' && data.topic.startsWith('_INBOX') === false) {
    //   console.log(JSON.stringify(data.topic) + ': ' + JSON.stringify(message))
    // }

    if (this.subscriptions) {
      const subs = this.subscriptions[data.topic]
      if (subs) {
        for (var i = 0; i < subs.length; i++) {
          // console.log('found a subscriber for topic: ' + data.topic)
          subs[i](data.topic, message)
        }
      }
    }
  },

  topic: function (name, callback) {
    if (!this.subscriptions[name]) this.subscriptions[name] = []
    this.subscriptions[name].push(callback)
  },
}

export default WebsocketService
