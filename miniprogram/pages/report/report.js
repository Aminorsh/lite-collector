const api = require('../../services/api')
const storage = require('../../services/storage')
const markdown = require('../../utils/markdown')

const MAX_POLLS = 60 // 2 minutes at 2s interval

Page({
  data: {
    formId: null,
    jobId: null,
    statusText: '',
    statusHint: '',
    loading: true,
    completed: false,
    failed: false,
    failedMsg: '',
    reportOutput: '',
    reportNodes: [],
    finishedAtText: '',
    hasPrevious: false,
    downloading: false,
  },

  _pollTimer: null,
  _pollCount: 0,

  onLoad(options) {
    if (!options.formId) return
    this.setData({ formId: options.formId })
    this.loadLatest()
  },

  onUnload() {
    this.clearPoll()
  },

  async loadLatest() {
    const app = getApp()
    await app.globalData.loginReady

    this.setData({ loading: true, completed: false, failed: false })
    try {
      const res = await api.request({
        url: '/forms/' + this.data.formId + '/report/latest',
        method: 'GET',
      })
      // 204 → api wrapper resolves with undefined; check.
      if (res && res.output) {
        this.setReport(res.output, res.job_id, res.finished_at)
      } else {
        this.setData({ loading: false, hasPrevious: false })
      }
    } catch (err) {
      if (err.status === 204) {
        this.setData({ loading: false, hasPrevious: false })
      } else {
        this.setData({ loading: false })
        wx.showToast({ title: err.message || '加载历史报告失败', icon: 'none' })
      }
    }
  },

  setReport(output, jobId, finishedAt) {
    const nodes = markdown.parse(output)
    this.setData({
      loading: false,
      completed: true,
      failed: false,
      reportOutput: output,
      reportNodes: nodes,
      jobId: jobId,
      finishedAtText: formatFinishedAt(finishedAt),
      hasPrevious: true,
    })
  },

  async onGenerate() {
    const app = getApp()
    await app.globalData.loginReady

    this.setData({
      loading: false,
      completed: false,
      failed: false,
      failedMsg: '',
      reportOutput: '',
      reportNodes: [],
      statusText: '正在排队...',
      statusHint: 'AI 正在准备分析数据',
    })

    try {
      const res = await api.post('/forms/' + this.data.formId + '/report')
      this.setData({ jobId: res.job_id })
      this._pollCount = 0
      this.pollJob()
    } catch (err) {
      this.setData({ failed: true, failedMsg: err.message || '无法创建报告任务' })
    }
  },

  pollJob() {
    this.clearPoll()
    this._pollTimer = setTimeout(() => this.checkJob(), 2000)
  },

  async checkJob() {
    this._pollCount++
    if (this._pollCount > MAX_POLLS) {
      this.setData({ failed: true, failedMsg: '生成超时，请稍后重试' })
      return
    }

    try {
      const res = await api.get('/jobs/' + this.data.jobId)
      if (res.status === 0) {
        this.setData({ statusText: '排队中...', statusHint: 'AI 正在准备分析数据' })
        this.pollJob()
      } else if (res.status === 1) {
        this.setData({ statusText: 'AI 正在分析数据...', statusHint: '即将完成，请耐心等待' })
        this.pollJob()
      } else if (res.status === 2) {
        let output = res.output || ''
        try {
          const parsed = JSON.parse(output)
          output = parsed.report || parsed.content || parsed.text || parsed.summary || output
        } catch (e) {
          // plain markdown — use as-is
        }
        this.setReport(output, this.data.jobId, res.finished_at)
      } else {
        this.setData({ failed: true, failedMsg: res.output || '报告生成失败' })
      }
    } catch (err) {
      this.setData({ failed: true, failedMsg: err.message || '查询任务状态失败' })
    }
  },

  clearPoll() {
    if (this._pollTimer) {
      clearTimeout(this._pollTimer)
      this._pollTimer = null
    }
  },

  onRegenerate() {
    this.onGenerate()
  },

  onRetry() {
    this.onGenerate()
  },

  async onDownloadPDF() {
    if (!this.data.jobId || this.data.downloading) return
    this.setData({ downloading: true })

    const token = storage.getToken()
    if (!token) {
      this.setData({ downloading: false })
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }

    const url = api.BASE_URL + '/jobs/' + this.data.jobId + '/pdf'
    wx.downloadFile({
      url: url,
      header: { Authorization: 'Bearer ' + token },
      success: (res) => {
        this.setData({ downloading: false })
        if (res.statusCode !== 200) {
          wx.showToast({ title: '下载失败 (' + res.statusCode + ')', icon: 'none' })
          return
        }
        wx.openDocument({
          filePath: res.tempFilePath,
          fileType: 'pdf',
          showMenu: true,
          fail: () => wx.showToast({ title: '打开 PDF 失败', icon: 'none' }),
        })
      },
      fail: () => {
        this.setData({ downloading: false })
        wx.showToast({ title: '网络错误', icon: 'none' })
      },
    })
  },
})

function formatFinishedAt(s) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  const pad = (n) => (n < 10 ? '0' + n : '' + n)
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
    ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
}
