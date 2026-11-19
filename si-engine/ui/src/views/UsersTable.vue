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
        <v-toolbar-title>Users</v-toolbar-title>
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
              New User
            </v-btn>
          </template>
          <v-card>
            <v-card-title>
              <span class="text-h5">{{ formTitle }}</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col
                    cols="12"
                    md="12"
                  >
                    <v-text-field
                      v-model="editedItem.fullname"
                      label="Full name"
                    />
                  </v-col>
                  <v-col
                    cols="12"
                    md="12"
                  >
                    <v-text-field
                      v-model="editedItem.email"
                      label="E-mail"
                    />
                  </v-col>
                  <v-col
                    cols="12"
                    md="12"
                  >
                    <v-select
                      v-model="editedItem.role"
                      :items="roles"
                      label="Role"
                    />
                  </v-col>
                  <v-col
                    cols="12"
                    md="12"
                  >
                    <v-text-field
                      v-model="editedItem.password"
                      type="password"
                      autocomplete="new-password"
                      :label="editedIndex === -1 ? 'Password (min 10 characters)' : 'New password (leave empty to keep)'"
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
        <v-dialog
          v-model="dialogDelete"
          max-width="500px"
        >
          <v-card>
            <v-card-title class="text-h5">
              Are you sure you want to delete user {{ editedItem.email }}?
            </v-card-title>
            <v-card-actions>
              <v-spacer />
              <v-btn
                color="blue darken-1"
                text
                @click="closeDelete"
              >
                Cancel
              </v-btn>
              <v-btn
                color="blue darken-1"
                text
                @click="deleteItemConfirm"
              >
                OK
              </v-btn>
              <v-spacer />
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>
    </template>
    <template v-slot:item.actions="{ item }">
      <v-icon
        small
        class="mr-2"
        @click="editItem(item)"
      >
        mdi-pencil
      </v-icon>
      <v-icon
        small
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
    name: 'UsersTablesView',

    data: () => ({
      snackbar: false,
      snackbarMessage: '',
      snackbarType: 'warning',
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
          width: 20,
        },
        { text: 'Full name', value: 'fullname', width: 150 },
        { text: 'E-mail', value: 'email' },
        { text: 'Role', value: 'role', width: 100 },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      roles: [
        { text: 'Administrator (read/write)', value: 'admin' },
        { text: 'Viewer (read only)', value: 'viewer' },
      ],
      items: [],
      editedIndex: -1,
      editedItem: {
        fullname: '',
        email: '',
        role: 'viewer',
        password: '',
      },
      defaultItem: {
        fullname: '',
        email: '',
        role: 'viewer',
        password: '',
      },
    }),

    computed: {
      formTitle () {
        return this.editedIndex === -1 ? 'New Item' : 'Edit Item'
      },
    },

    created () {
      this.initialize()
    },

    methods: {
      initialize () {
        this.loading = true
        ApiService.get('user')
          .then(response => {
            this.items = response.data.users
            this.loading = false
          }).catch(() => {
            this.loading = false
            this.$notification.error('Failed to load users')
          })
      },

      errorText (error) {
        return (error && error.response && error.response.data && error.response.data.error) || 'unknown error'
      },

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item, { password: '' })
        this.dialog = true
      },

      deleteItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.dialogDelete = true
      },

      deleteItemConfirm () {
        const id = this.editedItem.ID
        this.closeDelete()

        ApiService.delete('user/' + id)
          .then(() => {
            this.$notification.success('User deleted!')
            this.initialize()
          }).catch(error => {
            this.$notification.error('Failed to delete user: ' + this.errorText(error))
          })
      },

      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      closeDelete () {
        this.dialogDelete = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      save () {
        const item = Object.assign({}, this.editedItem)
        const request = this.editedIndex > -1
          ? ApiService.put('user/' + item.ID, item)
          : ApiService.post('user', item)

        request
          .then(response => {
            this.$notification.success('User ' + response.data.data.fullname + ' saved!')
            this.initialize()
          }).catch(error => {
            this.$notification.error('Failed to save user: ' + this.errorText(error))
          })
        this.close()
      },
    },
  }
</script>
