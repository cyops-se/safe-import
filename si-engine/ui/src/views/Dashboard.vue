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
              :lastseen="server.lastseen"
              :version="server.version"
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
        '1.si-inner.repos': { status: 'no connection', title: 'Inner - Repositories', color: 'error', icon: 'mdi-weather-pouring', lastseen: '' },
        '1.si-inner.http': { status: 'no connection', title: 'Inner - HTTP/HTTPS', color: 'error', icon: 'mdi-weather-pouring', lastseen: '' },
        '1.si-inner.dns': { status: 'no connection', title: 'Inner - DNS', color: 'error', icon: 'mdi-weather-pouring', lastseen: '' },
        '1.si-gatekeeper.proxy': { status: 'no connection', title: 'Gatekeeper', color: 'error', icon: 'mdi-weather-pouring', lastseen: '' },
        '1.si-outer.jobs': { status: 'no connection', title: 'Outer - Jobs', color: 'error', icon: 'mdi-weather-pouring', lastseen: '' },
      },
    }),

    computed: {
    },

    created () {
      var t = this
      WebsocketService.topic('system.heartbeat', function (topic, json) {
        var svc = JSON.parse(json)
        var server = t.servers[svc.name]
        if (server) {
          var now = new Date()
          server.status = now.toISOString().substring(0, 19).replace('T', ' ')
          server.icon = 'mdi-white-balance-sunny'
          server.color = 'success'
          server.version = svc.gitversion
          server.lastseen = now.toISOString()
        }
      })

      WebsocketService.topic('system.service.stopped', function (topic, name) {
        var server = t.servers[name]
        if (server) {
          var now = new Date()
          server.status = now.toISOString().substring(0, 19).replace('T', ' ')
          server.icon = 'mdi-weather-pouring'
          server.color = 'error'
        }
      })
    },
  }
</script>
