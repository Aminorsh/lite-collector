const api = require('../../services/api')

Page({
  data: {
    loading: true,
    forms: [],
  },

  onShow() {
    this.loadForms()
  },

  onPullDownRefresh() {
    this.loadForms().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadForms() {
    const app = getApp()
    await app.globalData.loginReady

    this.setData({ loading: true })
    try {
      const res = await api.get('/forms')
      const forms = (res.forms || []).map((f) => {
        return Object.assign({}, f, {
          updatedAtText: formatTime(f.updated_at || f.created_at),
        })
      })
      this.setData({ forms: forms, loading: false })
    } catch (err) {
      console.error('[index] loadForms error:', err)
      this.setData({ loading: false })
    }
  },

  onFormTap(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/form-detail/form-detail?formId=' + id })
  },

  onCreateTap() {
    wx.navigateTo({ url: '/pages/form-editor/form-editor' })
  },
})

function formatTime(isoStr) {
  if (!isoStr) return ''
  var d = new Date(isoStr)
  var month = d.getMonth() + 1
  var day = d.getDate()
  var hour = d.getHours()
  var minute = d.getMinutes()
  return month + '月' + day + '日 ' +
    (hour < 10 ? '0' : '') + hour + ':' +
    (minute < 10 ? '0' : '') + minute
}
