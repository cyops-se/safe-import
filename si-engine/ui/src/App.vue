<template>
  <v-fade-transition mode="out-in">
    <router-view />
  </v-fade-transition>
</template>

<script>
  // Styles
  import '@/styles/overrides.sass'
  import { sync } from 'vuex-pathify'
  import WebsocketService from '@/services/websocket.service'
  import ApiService from '@/services/api.service'

  export default {
    name: 'App',
    metaInfo: {
      title: 'si-engine',
      titleTemplate: '%s | cyops-se admin',
      htmlAttrs: { lang: 'en' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
    },

    computed: {
      sysinfo: sync('app/sysinfo'),
    },

    created () {
      WebsocketService.connect()

      ApiService.get('system/info')
        .then(response => {
          this.sysinfo = response.data.sysinfo
        }).catch(response => {
          console.log('ERROR response: ' + response.message)
        })
    },
  }
</script>
