<template>
  <v-card>
    <v-toolbar flat>
      <v-tabs
        v-model="tab"
        align-with-title
      >
        <v-tabs-slider color="yellow" />

        <v-tab key="GREY">
          GREY
        </v-tab>
        <v-tab key="WHITE">
          WHITE
        </v-tab>
        <v-tab key="BLACK">
          BLACK
        </v-tab>
        <v-spacer />
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search"
          single-line
          hide-details
          clearable
          class="my-auto mr-4"
        />
        <v-btn
          color="primary"
          dark
          class="my-auto mr-4"
          @click="pruneEntries"
        >
          Prune entries
        </v-btn>
      </v-tabs>
    </v-toolbar>
    <v-tabs-items v-model="tab">
      <v-tab-item key="GREY">
        <v-data-table
          :headers="headers"
          :items="greyitems"
          :search="search"
          :items-per-page="15"
          class="elevation-1"
        >
          <template v-slot:item.actions="{ item }">
            <v-icon
              class="mr-2"
              @click="makeWhite(item)"
            >
              mdi-shield-check
            </v-icon>
            <v-icon
              class="mr-2"
              @click="makeBlack(item)"
            >
              mdi-shield-off-outline
            </v-icon>
            <v-icon
              @click="deleteGrey(item)"
            >
              mdi-trash-can
            </v-icon>
          </template>
        </v-data-table>
      </v-tab-item>
      <v-tab-item key="WHITE">
        <v-data-table
          :headers="whiteheaders"
          :items="whiteitems"
          :search="search"
          :items-per-page="15"
          class="elevation-1"
        >
          <template v-slot:item.allowed="{ item }">
            <v-simple-checkbox
              v-model="item.allowed"
              @click="saveItem(item)"
            />
          </template>
          <template v-slot:item.noscan="{ item }">
            <v-simple-checkbox
              v-model="item.noscan"
              @click="saveItem(item)"
            />
          </template>
          <template v-slot:item.actions="{ item }">
            <v-icon
              class="mr-2"
              @click="editItem(item)"
            >
              mdi-pencil
            </v-icon>
            <v-icon
              @click="makeGrey(item)"
            >
              mdi-undo-variant
            </v-icon>
          </template>
        </v-data-table>
      </v-tab-item>
      <v-tab-item key="BLACK">
        <v-data-table
          :headers="headers"
          :items="blackitems"
          :search="search"
          :items-per-page="15"
          class="elevation-1"
        >
          <template v-slot:item.actions="{ item }">
            <v-icon
              class="mr-2"
              @click="editItem(item)"
            >
              mdi-pencil
            </v-icon>
            <v-icon
              @click="makeGrey(item)"
            >
              mdi-undo-variant
            </v-icon>
          </template>
        </v-data-table>
      </v-tab-item>
    </v-tabs-items>
    <v-dialog
      v-model="dialog"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">DNS item</span>
        </v-card-title>

        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.url"
                  label="Query"
                  outlined
                  hide-details
                />
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.matchurl"
                  label="Pattern"
                  outlined
                  hide-details
                />
              </v-col>
              <v-col
                v-if="editedItem.class === &quot;white&quot;"
                cols="12"
              >
                <v-checkbox
                  v-model="editedItem.allowed"
                  label="Allow through gateway"
                  hide-details
                  class="mt-n3"
                  :value="editedItem ? editedItem.allowed : true"
                />
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn
            color="blue darken-1"
            text
            @click="close"
          >
            Cancel
          </v-btn>
          <v-btn
            color="blue darken-1"
            text
            @click="save"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>
  import ApiService from '@/services/api.service'
  export default {
    name: 'BaselineHTTP',

    data: () => ({
      tab: null,
      dialog: false,
      dialogDelete: false,
      search: '',
      loading: false,
      headers: [
        { text: 'Time', align: 'start', filterable: true, value: 'timemod', width: '150' },
        { text: 'IP', value: 'fromip', width: '10%' },
        { text: 'Method', value: 'method', width: '70' },
        { text: 'URL', value: 'url', width: '50%' },
        { text: 'Count', value: 'count', width: '10%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      whiteheaders: [
        { text: 'Time', align: 'start', filterable: true, value: 'timemod', width: '150' },
        { text: 'IP', value: 'fromip', width: '10%' },
        { text: 'Method', value: 'method', width: '100' },
        { text: 'URL', value: 'url', width: '40%' },
        { text: 'Count', value: 'count', width: '100' },
        { text: 'Allowed', value: 'allowed', width: '120', sortable: false },
        { text: 'Exclude from scan', value: 'noscan', width: '100', sortable: false },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      greyitems: [],
      whiteitems: [],
      blackitems: [],
      editedIndex: -1,
      editedItem: {
        fullname: '',
        email: '',
      },
      defaultItem: {
        fullname: '',
        email: '',
      },
    }),

    created () {
      this.update()
    },

    methods: {
      initialize () {},

      update () {
        const t = this
        t.greyitems = []
        t.whiteitems = []
        t.blackitems = []

        this.loading = true
        ApiService.get('http')
          .then(response => {
            t.items = JSON.parse(response.data.items.payload)
            t.items.forEach((i) => {
              const dt = new Date(i.lastseen)
              i.lasttime = dt.toLocaleString() + '.' + dt.getMilliseconds()
              i.timemod = i.CreatedAt.replace('T', ' ').replace('Z', '').substring(0, 19)
              if (i.class === '' || i.class === 'grey') {
                t.greyitems.push(i)
              } else if (i.class === 'white') {
                t.whiteitems.push(i)
                if (i.allowed) i.allow = '<i class="nb-checkmark"></i>'
              } else {
                t.blackitems.push(i)
              }
            })
            t.loading = false
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      makeWhite (item) {
        var t = this
        item.class = 'white'
        item.scan = true
        ApiService.put('http/' + item.ID, item)
          .then(response => {
            t.update()
            this.$notification.success('HTTP(S) entry moved to white list!')
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      makeBlack (item) {
        var t = this
        item.class = 'black'
        ApiService.put('http/' + item.ID, item)
          .then(response => {
            t.update()
            this.$notification.success('HTTP(S) entry moved to black list!')
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      makeGrey (item) {
        var t = this
        item.class = 'grey'
        ApiService.put('http/' + item.ID, item)
          .then(response => {
            t.update()
            this.$notification.success('HTTP(S) entry moved to grey list!')
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      deleteGrey (item) {
        var t = this
        ApiService.delete('http/' + item.ID)
          .then(response => {
            console.log('response: ' + JSON.stringify(response))
            t.update()
            this.$notification.success('HTTP(S) entry deleted from grey list!')
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      pruneEntries () {
        var t = this
        ApiService.get('http/prune')
          .then(response => {
            t.update()
            this.$notification.success('HTTP(S) entries pruned!')
          }).catch(response => {
            console.log('ERROR response: ' + response)
          })
      },

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        console.log('editedItem: ' + JSON.stringify(this.editedItem))
        this.dialog = true
      },

      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      save () {
        if (this.editedIndex > -1) {
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('http/' + this.editedIndex.ID, this.editedItem)
            .then(response => {
              this.$notification.success('HTTP entry updated!')
            }).catch(response => {
              this.$notification.error('Failed to update HTTP entry!')
            })
        }
        this.close()
      },

      saveItem (item) {
        ApiService.put('http/' + item.ID, item)
          .then(response => {
            this.$notification.success('HTTP entry updated!')
          }).catch(response => {
            this.$notification.error('Failed to update HTTP entry!')
          })
        this.close()
      },
    },
  }
</script>

<style lang="sass">
td
  white-space: nowrap !important
  .truncate
    white-space: nowrap
    overflow: hidden
    text-overflow: ellipsis
</style>
