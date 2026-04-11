Page({
  data: {
    loading: true,
  },

  onLoad() {
    this.init()
  },

  async init() {
    const app = getApp()
    await app.globalData.loginReady
    this.setData({ loading: false })
  },
})
