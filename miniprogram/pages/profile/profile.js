const auth = require('../../services/auth')
const storage = require('../../services/storage')
const api = require('../../services/api')

Page({
  data: {
    userInfo: null,
    editing: false,
    editNickname: '',
    saving: false,
  },

  onShow() {
    const userInfo = storage.getUser()
    this.setData({ userInfo })
  },

  onEditNickname() {
    this.setData({
      editing: true,
      editNickname: this.data.userInfo ? this.data.userInfo.nickname : '',
    })
  },

  onNicknameInput(e) {
    this.setData({ editNickname: e.detail.value })
  },

  onCancelEdit() {
    this.setData({ editing: false, editNickname: '' })
  },

  async onSaveNickname() {
    var nickname = this.data.editNickname.trim()
    if (!nickname) {
      wx.showToast({ title: '昵称不能为空', icon: 'none' })
      return
    }

    this.setData({ saving: true })
    try {
      var res = await api.put('/user/profile', { nickname: nickname, avatar_url: '' })
      // Update local storage
      var user = storage.getUser()
      user.nickname = res.nickname || nickname
      storage.setUser(user)
      var app = getApp()
      app.globalData.userInfo = user

      this.setData({
        saving: false,
        editing: false,
        userInfo: user,
      })
      wx.showToast({ title: '修改成功', icon: 'success' })
    } catch (err) {
      this.setData({ saving: false })
      wx.showToast({ title: err.message || '修改失败', icon: 'none' })
    }
  },

  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          auth.logout()
          this.setData({ userInfo: null })
          auth.silentLogin().then(() => {
            this.onShow()
          })
        }
      },
    })
  },
})
