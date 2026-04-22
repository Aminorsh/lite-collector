const auth = require('./services/auth')

App({
  globalData: {
    token: '',
    userInfo: null,
    // loginReady resolves once the initial silent login completes.
    // Pages can: await getApp().loginReady
    loginReady: null,
    // If the app is launched from a share card, the target formId is stashed
    // here and redirected to after silent login completes.
    pendingShareFormId: null,
  },

  onLaunch(options) {
    var pendingFormId = this.extractShareFormId(options)
    if (pendingFormId) {
      this.globalData.pendingShareFormId = pendingFormId
    }

    this.globalData.loginReady = auth.silentLogin()
      .then((user) => {
        console.log('[App] login success:', user.nickname)
        this.redirectToShareTarget()
      })
      .catch((err) => {
        console.error('[App] login failed:', err)
      })
  },

  // onShow fires whenever the app is foregrounded — including when a
  // previously-opened app is re-entered via a share card. Handle that path
  // too; onLaunch only fires on a cold start.
  onShow(options) {
    var pendingFormId = this.extractShareFormId(options)
    if (!pendingFormId) return
    this.globalData.pendingShareFormId = pendingFormId
    // If login already resolved, redirect immediately; otherwise onLaunch
    // will handle it when login completes.
    if (this.globalData.token) {
      this.redirectToShareTarget()
    }
  },

  // Extracts a share target's formId from the launch/show options. WeChat
  // routes shared cards with the target page's query stashed in options.query.
  extractShareFormId(options) {
    if (!options) return null
    if (options.query && options.query.formId) return options.query.formId
    // Fallback: some scenes put it on options.path with a query string.
    if (options.path && options.path.indexOf('formId=') !== -1) {
      var m = options.path.match(/formId=([^&]+)/)
      if (m) return decodeURIComponent(m[1])
    }
    return null
  },

  redirectToShareTarget() {
    var formId = this.globalData.pendingShareFormId
    if (!formId) return
    this.globalData.pendingShareFormId = null
    wx.reLaunch({ url: '/pages/form-fill/form-fill?formId=' + formId })
  },
})
