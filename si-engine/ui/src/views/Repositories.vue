<template>
  <v-data-table
    :headers="headers"
    :items="items"
    class="elevation-1"
  >
    <template v-slot:top>
      <v-toolbar
        flat
      >
        <v-toolbar-title>Approved repositories</v-toolbar-title>
        <v-divider
          class="mx-4"
          inset
          vertical
        />
        <v-spacer />
        <v-dialog
          v-model="dialog"
          max-width="500px"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              color="primary"
              dark
              class="mb-2"
              v-bind="attrs"
              v-on="on"
            >
              New repository
            </v-btn>
          </template>
          <v-card>
            <v-card-title>
              <span class="text-h5">Repository</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col cols="12">
                    <v-text-field
                      v-model="editedItem.url"
                      label="URL"
                      outlined
                      hide-details
                    />
                  </v-col>
                  <v-col cols="12">
                    <v-checkbox
                      v-model="editedItem.recursive"
                      label="Recursive download"
                      hide-details
                      class="mt-n3"
                      :value="editedItem ? editedItem.recursive : true"
                    />
                  </v-col>
                  <v-col cols="12">
                    <v-text-field
                      v-model="editedItem.username"
                      label="User name"
                      outlined
                      hide-details
                    />
                  </v-col>
                  <v-col cols="12">
                    <v-text-field
                      v-model="editedItem.password"
                      label="Password"
                      outlined
                      hide-details
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
      </v-toolbar>
    </template>
    <template v-slot:item.actions="{ item }">
      <v-icon
        class="mr-1"
        @click="playItem(item)"
      >
        mdi-play
      </v-icon>
      <v-icon
        class="mr-1"
        @click="editItem(item)"
      >
        mdi-pencil
      </v-icon>
      <v-icon
        @click="deleteItem(item)"
      >
        mdi-delete
      </v-icon>
    </template>
  </v-data-table>
</template>

<script>
  import ApiService from '@/services/api.service'

  export default {
    name: 'Repositories',

    data: () => ({
      dialog: false,
      dialogDelete: false,
      search: '',
      loading: false,
      headers: [
        {
          text: 'ID',
          align: 'start',
          filterable: false,
          value: 'ID',
          width: 75,
        },
        { text: 'URL', value: 'url', width: '50%' },
        { text: 'User', value: 'username', width: '10%' },
        { text: 'Last Success', value: 'lastsuccessmod', width: '10%' },
        { text: 'Last Failure', value: 'lastfailuremod', width: '10%' },
        { text: 'Available', value: 'available', width: '10%' },
        { text: 'Actions', value: 'actions', width: '100px', sortable: false },
      ],
      items: [],
      editedIndex: -1,
      editedItem: {},
      defaultItem: {
      },
    }),

    created () {
      this.loading = true
      this.editedItem = Object.assign({}, this.defaultItem)
      this.editedIndex = -1
      ApiService.get('repo')
        .then(response => {
          var items = JSON.parse(response.data.items.payload)
          items.forEach((item) => {
            item.lastsuccessmod = item.lastsuccess.replace('T', ' ').substring(0, 19)
            item.lastfailuremod = item.lastfailure.replace('T', ' ').substring(0, 19)
          })
          this.items = items
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + response.message)
        })
    },

    methods: {
      initialize () {},

      playItem (item) {
        var t = this
        ApiService.get('repo/download/' + item.ID)
          .then(response => {
            console.log('Download request response: ' + response.data.data.payload)
            t.$notification.success('Repository download started!')
            this.loading = false
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
          })
      },

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.dialog = true
      },

      deleteItem (item) {
        var t = this
        ApiService.delete('repo/' + item.ID)
          .then(response => {
            for (var i = 0; i < this.items.length; i++) {
              if (this.items[i].ID === item.ID) this.items.splice(i, 1)
            }
            t.$notification.success('Repository deleted!')
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
            t.$notification.error('Failed to delete respoitory!' + response.message)
          })
      },

      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      save () {
        console.log('edit item' + JSON.stringify(this.editedItem))
        var t = this
        if (this.editedIndex > -1) {
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('repo/' + this.editedItem.ID, this.editedItem)
            .then(response => {
              t.$notification.success('Repository updated!')
            }).catch(function (response) {
              console.log('Failed to update respoitory! ' + JSON.stringify(response))
              t.$notification.error('Failed to update respoitory!' + response.message)
            })
        } else {
          ApiService.post('repo', this.editedItem)
            .then(response => {
              console.log('Respoitory created! ' + response.data.data.payload)
              var item = JSON.parse(response.data.data.payload)
              this.$notification.success('Repository created!')
              this.items.push(item)
            }).catch(function (response) {
              console.log('Failed to create respoitory! ' + response.message)
              this.$notification.error('Failed to create respoitory!' + response)
            })
        }
        this.editedItem = Object.assign({}, this.defaultItem)
        this.editedIndex = -1
        this.close()
      },
    },
  }
</script>

<style lang="sass">
td
  white-space: nowrap !important
</style>
