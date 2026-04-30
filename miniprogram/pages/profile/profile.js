const auth = require('../../services/auth')
const storage = require('../../services/storage')
const api = require('../../services/api')

Page({
  data: {
    userInfo: null,
    editing: false,
    editNickname: '',
    saving: false,
    uploading: false,
    todayQuote: "",
    quoteAuthor: "",
  },

  onLoad() {
    this.generateDailyQuote();
  },

  generateDailyQuote() {
    // 纯文本语录库
    const quoteList = [
      "不要来，慢慢急",
      "出淤泥而抹全身",
      "慢工出烂活，欲速则一坨",
      "话到嘴边又咽了下去 每天以此获得饱腹感",
      "花香蕉的钱就只能请到我这样的猴子",
      "退一万步来讲的话 根本就听不到你讲什么",
      "干我们这一行最忌讳的就是干我们这一行",
      "我收到牛国英津大学的邀请了，再见了我的朋友们",
      "我不惹事，我也怕事",
      "以柔克刚 以巧克力 以德克士",
      "马上离开 不要回来",
      "每天一个仰卧起坐，已经坚持了二十多年了",
    ];
    
    // 随机选择一条
    const randomIndex = Math.floor(Math.random() * quoteList.length);
    const quote = quoteList[randomIndex];
    
    this.setData({
      todayQuote: quote
    });
  },

  onOpenAIChat() {
    wx.navigateTo({
      url: '/pages/ai-chat/ai-chat'
    })
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

  onChooseAvatar(e) {
    var tempPath = e.detail.avatarUrl
    if (!tempPath) return

    this.setData({ uploading: true })
    var token = storage.getToken()

    var that = this
    wx.uploadFile({
      url: api.BASE_URL + '/user/avatar',
      filePath: tempPath,
      name: 'file',
      header: { Authorization: 'Bearer ' + token },
      success(res) {
        var data
        try { data = JSON.parse(res.data) } catch (e) { data = null }
        if (res.statusCode >= 200 && res.statusCode < 300 && data && data.avatar_url) {
          var user = storage.getUser() || {}
          user.avatar_url = data.avatar_url
          storage.setUser(user)
          var app = getApp()
          app.globalData.userInfo = user
          that.setData({ uploading: false, userInfo: user })
          wx.showToast({ title: '头像已更新', icon: 'success' })
        } else {
          var msg = (data && data.error && data.error.message) || '上传失败'
          that.setData({ uploading: false })
          wx.showToast({ title: msg, icon: 'none' })
        }
      },
      fail() {
        that.setData({ uploading: false })
        wx.showToast({ title: '网络错误', icon: 'none' })
      },
    })
  },

  // onLogout() {
  //   wx.showModal({
  //     title: '提示',
  //     content: '确定退出登录吗？',
  //     success: (res) => {
  //       if (res.confirm) {
  //         auth.logout()
  //         this.setData({ userInfo: null })
  //         auth.silentLogin().then(() => {
  //           this.onShow()
  //         })
  //       }
  //     },
  //   })
  // },
})
