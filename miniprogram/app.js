const auth = require('./services/auth')

App({
  globalData: {
    token: '',
    userInfo: null,
    // loginReady resolves once the initial silent login completes.
    // Pages can: await getApp().loginReady
    loginReady: null,
  },

  onLaunch() {
    this.globalData.loginReady = auth.silentLogin()
      .then((user) => {
        console.log('[App] login success:', user.nickname)
      })
      .catch((err) => {
        console.error('[App] login failed:', err)
      })
  },
})
