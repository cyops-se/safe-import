<template>
  <v-container
    id="dashboard-view"
    fluid
    tag="section"
  >
    <v-row>
      <v-col cols="12">
        <v-row>
          <v-col
            v-for="(server, i) in servers"
            :key="`server-${i}`"
            cols="12"
            md="4"
            lg="3"
          >
            <simple-info-card
              :icon="server.icon"
              :color="server.color"
              :value="server.status"
              :title="server.title"
            />
          </v-col>
        </v-row>
      </v-col>
      <error-logs-tables-view />
    </v-row>
  </v-container>
</template>

<script>
  // Utilities
  import ErrorLogsTablesView from './ErrorLogs'
  import WebsocketService from '@/services/websocket.service'

  export default {
    name: 'DashboardView',

    components: {
      ErrorLogsTablesView,
    },

    data: () => ({
      servers: {
        '1.si-inner.repos': { status: 'ok', title: 'Inner - Repositories', color: 'error', icon: 'mdi-weather-pouring', lastseen: new Date() },
        '1.si-inner.http': { status: 'ok', title: 'Inner - HTTP/HTTPS', color: 'error', icon: 'mdi-weather-pouring', lastseen: new Date() },
        '1.si-inner.dns': { status: 'ok', title: 'Inner - DNS', color: 'error', icon: 'mdi-weather-pouring', lastseen: new Date() },
        '1.si-gatekeeper.proxy': { status: 'ok', title: 'Gatekeeper', color: 'error', icon: 'mdi-weather-pouring', lastseen: new Date() },
        '1.si-outer.jobs': { status: 'ok', title: 'Outer - Jobs', color: 'error', icon: 'mdi-weather-pouring', lastseen: new Date() },
      },
    }),

    computed: {
    },

    watch: {
      $route (to, from) {
        console.log('route change: ', to, from)
      },
    },

    created () {
      var t = this
      WebsocketService.topic('system.heartbeat', function (topic, json) {
        console.log(topic + ': ' + json)
        var svc = JSON.parse(json)
        var server = t.servers[svc.name]
        if (server) {
          server.status = 'ok'
          server.icon = 'mdi-white-balance-sunny'
          server.color = 'success'
        }
      })
    },
  }
</script>
