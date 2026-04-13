const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')

Page({
  data: {
    formId: null,
    form: null,
    fieldCount: 0,
    createdAtText: '',
    updatedAtText: '',
  },

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
    }
  },

  onShow() {
    if (this.data.formId) {
      this.loadForm(this.data.formId)
    }
  },

  async loadForm(formId) {
    var app = getApp()
    await app.globalData.loginReady

    try {
      var res = await api.get('/forms/' + formId)
      var fields = schemaToFields(res.schema)
      this.setData({
        form: res,
        fieldCount: fields.length,
        createdAtText: formatTime(res.created_at),
        updatedAtText: formatTime(res.updated_at),
      })
    } catch (err) {
      console.error('[form-detail] load error:', err)
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onEdit() {
    wx.navigateTo({
      url: '/pages/form-editor/form-editor?formId=' + this.data.formId,
    })
  },

  onPublish() {
    wx.showModal({
      title: '确认发布',
      content: '发布后将开放填写，确定发布吗？',
      success: (res) => {
        if (res.confirm) this.doPublish()
      },
    })
  },

  async doPublish() {
    wx.showLoading({ title: '发布中...' })
    try {
      await api.post('/forms/' + this.data.formId + '/publish')
      wx.showToast({ title: '发布成功', icon: 'success' })
      this.loadForm(this.data.formId)
    } catch (err) {
      wx.showToast({ title: err.message || '发布失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

  onArchive() {
    wx.showModal({
      title: '确认归档',
      content: '归档后将不再接受新提交，确定归档吗？',
      success: (res) => {
        if (res.confirm) this.doArchive()
      },
    })
  },

  async doArchive() {
    wx.showLoading({ title: '归档中...' })
    try {
      await api.post('/forms/' + this.data.formId + '/archive')
      wx.showToast({ title: '已归档', icon: 'success' })
      this.loadForm(this.data.formId)
    } catch (err) {
      wx.showToast({ title: err.message || '归档失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

  onViewSubmissions() {
    wx.navigateTo({
      url: '/pages/submissions/submissions?formId=' + this.data.formId,
    })
  },

  onBaseData() {
    wx.navigateTo({
      url: '/pages/base-data/base-data?formId=' + this.data.formId,
    })
  },

  onGenerateReport() {
    wx.navigateTo({
      url: '/pages/report/report?formId=' + this.data.formId,
    })
  },

  onShareAppMessage() {
    var form = this.data.form
    return {
      title: form ? form.title : '数据收集表单',
      path: '/pages/form-fill/form-fill?formId=' + this.data.formId,
    }
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
