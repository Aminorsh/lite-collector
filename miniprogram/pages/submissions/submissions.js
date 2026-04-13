const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')

Page({
  data: {
    formId: null,
    loading: true,
    viewMode: 'list',
    submissions: [],
    anomalyCount: 0,
    // Overview data
    overviewColumns: [],
    overviewData: [],
  },

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
      this.loadList()
    }
  },

  onPullDownRefresh() {
    var p = this.data.viewMode === 'list' ? this.loadList() : this.loadOverview()
    p.then(() => { wx.stopPullDownRefresh() })
  },

  onSwitchView(e) {
    var mode = e.currentTarget.dataset.mode
    if (mode === this.data.viewMode) return
    this.setData({ viewMode: mode })
    if (mode === 'overview' && this.data.overviewData.length === 0) {
      this.loadOverview()
    }
  },

  async loadList() {
    var app = getApp()
    await app.globalData.loginReady
    this.setData({ loading: true })
    try {
      var res = await api.get('/forms/' + this.data.formId + '/submissions')
      var items = (res.submissions || []).map(function (s) {
        return Object.assign({}, s, {
          submittedAtText: formatTime(s.submitted_at),
        })
      })
      var anomalyCount = items.filter(function (s) { return s.status === 2 }).length
      this.setData({ submissions: items, anomalyCount: anomalyCount, loading: false })
    } catch (err) {
      console.error('[submissions] loadList error:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  async loadOverview() {
    var app = getApp()
    await app.globalData.loginReady
    this.setData({ loading: true })
    try {
      var res = await api.get('/forms/' + this.data.formId + '/submissions/overview')
      var fields = schemaToFields(res.schema)
      var columns = fields.map(function (f) { return { key: f.key, label: f.label } })

      var dataRows = (res.submissions || []).map(function (s) {
        // Flatten array values for display
        var displayValues = {}
        if (s.values) {
          Object.keys(s.values).forEach(function (k) {
            var v = s.values[k]
            displayValues[k] = Array.isArray(v) ? v.join('、') : v
          })
        }
        return {
          id: s.id,
          status: s.status,
          values: displayValues,
          anomaly_reasons: s.anomaly_reasons || [],
        }
      })

      // Also update the list-view submissions count
      var anomalyCount = dataRows.filter(function (s) { return s.status === 2 }).length
      this.setData({
        overviewColumns: columns,
        overviewData: dataRows,
        anomalyCount: anomalyCount,
        loading: false,
      })
    } catch (err) {
      console.error('[submissions] loadOverview error:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onSubmissionTap(e) {
    var id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: '/pages/submission-detail/submission-detail?formId=' + this.data.formId + '&submissionId=' + id,
    })
  },

  onOverviewRowTap(e) {
    var idx = e.currentTarget.dataset.index
    var row = this.data.overviewData[idx]
    if (row.status === 2 && row.anomaly_reasons.length > 0) {
      wx.showModal({
        title: '异常原因',
        content: row.anomaly_reasons.join('\n'),
        showCancel: false,
      })
    } else {
      wx.navigateTo({
        url: '/pages/submission-detail/submission-detail?formId=' + this.data.formId + '&submissionId=' + row.id,
      })
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
