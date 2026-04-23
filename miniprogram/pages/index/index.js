const api = require('../../services/api')
const jobs = require('../../services/jobs')

Page({
  data: {
    loading: true,
    forms: [],
    pendingJobs: [], // [{ id, jobType, status, title, path, inFlight }]
  },

  onShow() {
    this.loadForms()
    this.loadPendingJobs()
  },

  onPullDownRefresh() {
    Promise.all([this.loadForms(), this.loadPendingJobs()]).then(() => {
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

  async loadPendingJobs() {
    const app = getApp()
    await app.globalData.loginReady

    const raw = await jobs.fetchVisible()
    const items = []
    for (let i = 0; i < raw.length; i++) {
      const j = raw[i]
      const d = jobs.describe(j)
      if (!d) continue
      items.push({
        id: j.id,
        jobType: j.job_type,
        status: j.status,
        title: d.title,
        path: d.path,
        inFlight: j.status === 0 || j.status === 1,
      })
    }
    this.setData({ pendingJobs: items })
  },

  onJobTap(e) {
    const idx = e.currentTarget.dataset.index
    const job = this.data.pendingJobs[idx]
    if (!job) return
    wx.navigateTo({ url: job.path })
  },

  onJobDismiss(e) {
    const idx = e.currentTarget.dataset.index
    const job = this.data.pendingJobs[idx]
    if (!job || job.inFlight) return
    jobs.ack(job.id)
    const next = this.data.pendingJobs.slice()
    next.splice(idx, 1)
    this.setData({ pendingJobs: next })
  },

  onFormTap(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/form-detail/form-detail?formId=' + id })
  },

  onCreateTap() {
    wx.showActionSheet({
      itemList: ['手动创建', 'AI 创建'],
      success: function (res) {
        if (res.tapIndex === 0) {
          wx.navigateTo({ url: '/pages/form-editor/form-editor' })
        } else if (res.tapIndex === 1) {
          wx.navigateTo({ url: '/pages/ai-generate/ai-generate' })
        }
      },
    })
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
