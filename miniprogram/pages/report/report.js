const api = require('../../services/api')

Page({
  data: {
    formId: null,
    jobId: null,
    statusText: '正在排队...',
    statusHint: '请稍候，AI 正在准备分析数据',
    completed: false,
    failed: false,
    failedMsg: '',
    reportOutput: '',
  },

  _pollTimer: null,
  _pollCount: 0,

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
      this.startReport()
    }
  },

  onUnload() {
    this.clearPoll()
  },

  async startReport() {
    var app = getApp()
    await app.globalData.loginReady

    this.setData({
      completed: false,
      failed: false,
      failedMsg: '',
      reportOutput: '',
      statusText: '正在排队...',
      statusHint: '请稍候，AI 正在准备分析数据',
    })

    try {
      var res = await api.post('/forms/' + this.data.formId + '/report')
      this.setData({ jobId: res.job_id })
      this._pollCount = 0
      this.pollJob()
    } catch (err) {
      this.setData({
        failed: true,
        failedMsg: err.message || '无法创建报告任务',
      })
    }
  },

  pollJob() {
    this.clearPoll()
    this._pollTimer = setTimeout(() => {
      this.checkJob()
    }, 2000)
  },

  async checkJob() {
    this._pollCount++

    // Cap at 60 polls (2 minutes)
    if (this._pollCount > 60) {
      this.setData({
        failed: true,
        failedMsg: '生成超时，请稍后重试',
      })
      return
    }

    try {
      var res = await api.get('/jobs/' + this.data.jobId)

      // Status: 0=queued, 1=processing, 2=completed, 3=failed
      if (res.status === 0) {
        this.setData({
          statusText: '排队中...',
          statusHint: '请稍候，AI 正在准备分析数据',
        })
        this.pollJob()
      } else if (res.status === 1) {
        this.setData({
          statusText: 'AI 正在分析数据...',
          statusHint: '即将完成，请耐心等待',
        })
        this.pollJob()
      } else if (res.status === 2) {
        // Completed
        var output = res.output || ''
        // Try to parse as JSON in case the output is wrapped
        try {
          var parsed = JSON.parse(output)
          output = parsed.report || parsed.content || parsed.text || output
        } catch (e) {
          // Not JSON, use as-is
        }
        this.setData({ completed: true, reportOutput: output })
      } else {
        // Failed
        this.setData({
          failed: true,
          failedMsg: res.output || '报告生成失败',
        })
      }
    } catch (err) {
      this.setData({
        failed: true,
        failedMsg: err.message || '查询任务状态失败',
      })
    }
  },

  clearPoll() {
    if (this._pollTimer) {
      clearTimeout(this._pollTimer)
      this._pollTimer = null
    }
  },

  onRetry() {
    this.startReport()
  },
})
