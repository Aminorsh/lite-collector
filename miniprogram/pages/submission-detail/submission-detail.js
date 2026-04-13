const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')

Page({
  data: {
    formId: null,
    submissionId: null,
    loading: true,
    submission: null,
    fields: [],
    submittedAtText: '',
  },

  onLoad(options) {
    this.setData({
      formId: options.formId,
      submissionId: options.submissionId,
    })
    this.loadData()
  },

  async loadData() {
    var app = getApp()
    await app.globalData.loginReady

    try {
      // Load form schema and submission in parallel
      var [form, submission] = await Promise.all([
        api.get('/forms/' + this.data.formId),
        api.get('/forms/' + this.data.formId + '/submissions/' + this.data.submissionId),
      ])

      var fields = schemaToFields(form.schema)
      this.setData({
        fields: fields,
        submission: submission,
        submittedAtText: formatTime(submission.submitted_at),
        loading: false,
      })
    } catch (err) {
      console.error('[submission-detail] load error:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
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
