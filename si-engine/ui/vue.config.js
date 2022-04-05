module.exports = {
  devServer: {
    disableHostCheck: true,
    proxy: {
      '^/api': {
        target: 'http://localhost:7499/',
        ws: false,
        changeOrigin: true,
      },
      '^/auth': {
        target: 'http://localhost:7499/',
        ws: false,
        changeOrigin: true,
      },
      '^/static': {
        target: 'http://localhost:7499/',
        ws: false,
        changeOrigin: true,
      },
      '^/ws': {
        target: 'ws://localhost:7499/ws',
        ws: true,
        changeOrigin: true,
      },
    },
  },

  transpileDependencies: ['vuetify'],

  publicPath: '/ui',
}
