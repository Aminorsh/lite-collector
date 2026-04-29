const api = require('../../services/api')
const jobs = require('../../services/jobs')

const FILTER_KEY = 'formListFilter'

const STATUS_TABS = [
  { label: '全部', value: '' },
  { label: '草稿', value: '0' },
  { label: '已发布', value: '1' },
  { label: '已归档', value: '2' },
]

const SORT_OPTIONS = [
  { label: '最近更新', sort: 'updated_at', order: 'desc' },
  { label: '最近创建', sort: 'created_at', order: 'desc' },
  { label: '标题 A→Z', sort: 'title', order: 'asc' },
]

Page({
  data: {
    loading: true,
    forms: [],
    query: '',
    status: '',
    sortIndex: 0,
    statusTabs: STATUS_TABS,
    sortOptions: SORT_OPTIONS,
    pendingJobs: [], // [{ id, jobType, status, title, path, inFlight }]
    showSplashModal: true
  },

  _debounceTimer: null,

  closeSplashModal() {
    this.setData({ showSplashModal: false })
  },



  onLoad() {
    const saved = wx.getStorageSync(FILTER_KEY) || {}
    this.setData({
      query: saved.query || '',
      status: saved.status || '',
      sortIndex: typeof saved.sortIndex === 'number' ? saved.sortIndex : 0,
    })
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

  persistFilter() {
    wx.setStorageSync(FILTER_KEY, {
      query: this.data.query,
      status: this.data.status,
      sortIndex: this.data.sortIndex,
    })
  },

  async loadForms() {
    const app = getApp()
    await app.globalData.loginReady

    this.setData({ loading: true })
    try {
      const sort = SORT_OPTIONS[this.data.sortIndex]
      const params = { sort: sort.sort, order: sort.order }
      if (this.data.query) params.q = this.data.query
      if (this.data.status) params.status = this.data.status

      const res = await api.get('/forms', params)
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

  onQueryInput(e) {
    const val = e.detail.value
    this.setData({ query: val })
    clearTimeout(this._debounceTimer)
    this._debounceTimer = setTimeout(() => {
      this.persistFilter()
      this.loadForms()
    }, 300)
  },

  onStatusTap(e) {
    const value = e.currentTarget.dataset.value
    if (value === this.data.status) return
    this.setData({ status: value })
    this.persistFilter()
    this.loadForms()
  },

  onSortChange(e) {
    const idx = Number(e.detail.value)
    if (idx === this.data.sortIndex) return
    this.setData({ sortIndex: idx })
    this.persistFilter()
    this.loadForms()
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
