<template>
  <v-container
    fluid
    tag="section"
  >
    <v-row>
      <v-col cols="12">
        <v-row>
          <v-col
            v-for="(item, i) in items"
            :key="`item-${i}`"
            cols="12"
          >
            <simple-info-card
              icon="mdi-file"
              :color="item.state.includes('fail') ? 'error' : 'success'"
              :value="item.state"
              :title="item.url"
            />
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
  import Vue from 'vue'
  import WebsocketService from '@/services/websocket.service'
  export default {
    name: 'Jobs',

    data: () => ({
      items: [],
    }),

    created () {
      var t = this
      WebsocketService.topic('file.download.start', function (topic, url) {
        console.log('jobs: ' + url)
        t.setState(url, 'downloading')
      })
      WebsocketService.topic('file.download.fail', function (topic, url) {
        t.setState(url, 'download failed')
      })
      WebsocketService.topic('file.download.success', function (topic, url) {
        t.setState(url, 'download complete')
      })
      WebsocketService.topic('file.download.scan.start', function (topic, url) {
        t.setState(url, 'scanning')
      })
      WebsocketService.topic('file.download.scan.fail', function (topic, url) {
        t.setState(url, 'scan failed')
      })
      WebsocketService.topic('file.download.scan.success', function (topic, url) {
        t.setState(url, 'scan complete')
      })
      WebsocketService.topic('file.download.link.start', function (topic, url) {
        t.setState(url, 'link start')
      })
      WebsocketService.topic('file.download.link.fail', function (topic, url) {
        t.setState(url, 'link failed')
      })
      WebsocketService.topic('file.download.link.success', function (topic, url) {
        t.setState(url, 'link complete')
        setTimeout(function () {
          var item = t.items.find(i => i.url === url)
          var idx = t.items.indexOf(item)
          Vue.delete(t.items, idx)
        }, 1000)
      })
    },

    methods: {
      initialize () {},
      setState (url, state) {
        var item = this.items.find(i => i.url === url)
        if (!item) {
          item = { url: url }
          this.items.push(item)
        }

        var idx = this.items.indexOf(item)

        item.state = state
        item.time = new Date()
        // Object.assign(this.items[idx], item)
        Vue.set(this.items, idx, item)

        console.log('items: ' + JSON.stringify(this.items))
      },
    },
  }
</script>
