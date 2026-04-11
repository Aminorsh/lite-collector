const auth = require('../../services/auth')
const storage = require('../../services/storage')

Page({
  data: {
    userInfo: null,
  },

  onShow() {
    const userInfo = storage.getUser()
    this.setData({ userInfo })
  },

  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          auth.logout()
          this.setData({ userInfo: null })
          // Re-login silently
          auth.silentLogin().then(() => {
            this.onShow()
          })
        }
      },
    })
  },
})
