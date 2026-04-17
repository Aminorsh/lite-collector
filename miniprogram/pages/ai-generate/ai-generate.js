const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')
const { FIELD_TYPES } = require('../../utils/constants')

const STORAGE_KEY = 'aiGenerateJobId'
const MAX_POLLS = 60 // 2 minutes at 2s interval

const TYPE_LABEL_MAP = {}
FIELD_TYPES.forEach(function (t) { TYPE_LABEL_MAP[t.value] = t.label })

Page({
  data: {
    description: '',
    generating: false,
    statusText: '',
    result: null,
    previewFields: [],
  },

  _pollTimer: null,
  _pollCount: 0,

  onLoad() {
    const jobId = wx.getStorageSync(STORAGE_KEY)
    if (jobId) {
      this.setData({
        generating: true,
        statusText: '继续等待上次的生成任务...',
      })
      this._pollCount = 0
      this.pollJob(jobId)
    }
  },

  onUnload() {
    this.clearPoll()
  },

  onDescInput(e) {
    this.setData({ description: e.detail.value })
  },

  async onGenerate() {
    const desc = this.data.description.trim()
    if (!desc) {
      wx.showToast({ title: '请输入表单描述', icon: 'none' })
      return
    }

    this.setData({
      generating: true,
      result: null,
      previewFields: [],
      statusText: '正在排队...',
    })

    try {
      const res = await api.post('/forms/generate', { description: desc })
      wx.setStorageSync(STORAGE_KEY, res.job_id)
      this._pollCount = 0
      this.pollJob(res.job_id)
    } catch (err) {
      this.setData({ generating: false, statusText: '' })
      this.showError(err)
    }
  },

  pollJob(jobId) {
    this.clearPoll()
    this._pollTimer = setTimeout(() => this.checkJob(jobId), 2000)
  },

  async checkJob(jobId) {
    this._pollCount++
    if (this._pollCount > MAX_POLLS) {
      this.failGeneration('生成超时，请稍后重试', jobId)
      return
    }

    try {
      const res = await api.get('/jobs/' + jobId)
      if (res.status === 0) {
        this.setData({ statusText: '排队中，AI 即将开始...' })
        this.pollJob(jobId)
      } else if (res.status === 1) {
        this.setData({ statusText: 'AI 正在生成表单结构...' })
        this.pollJob(jobId)
      } else if (res.status === 2) {
        this.handleComplete(res.output)
      } else {
        this.failGeneration(res.output || 'AI 生成失败', jobId)
      }
    } catch (err) {
      this.failGeneration(err.message || '查询生成状态失败', jobId)
    }
  },

  handleComplete(output) {
    wx.removeStorageSync(STORAGE_KEY)
    let parsed
    try {
      parsed = JSON.parse(output || '{}')
    } catch (e) {
      this.setData({ generating: false, statusText: '' })
      wx.showToast({ title: '生成结果解析失败', icon: 'none' })
      return
    }

    const fields = schemaToFields(parsed.schema)
    const previewFields = fields.map(function (f) {
      return {
        key: f.key,
        label: f.label,
        type: f.type,
        typeLabel: TYPE_LABEL_MAP[f.type] || f.type,
        required: f.required,
      }
    })
    this.setData({
      generating: false,
      statusText: '',
      result: {
        title: parsed.title,
        description: parsed.description,
        schema: parsed.schema,
      },
      previewFields: previewFields,
    })
  },

  failGeneration(msg, jobId) {
    if (jobId) wx.removeStorageSync(STORAGE_KEY)
    this.setData({ generating: false, statusText: '' })
    wx.showToast({ title: msg, icon: 'none', duration: 3000 })
  },

  showError(err) {
    let msg = err.message || 'AI 生成失败'
    if (err.status === 503) msg = 'AI 服务暂未开启，请手动创建表单'
    wx.showToast({ title: msg, icon: 'none', duration: 3000 })
  },

  clearPoll() {
    if (this._pollTimer) {
      clearTimeout(this._pollTimer)
      this._pollTimer = null
    }
  },

  onRegenerate() {
    this.setData({ result: null, previewFields: [] })
  },

  onUseForm() {
    const app = getApp()
    app.globalData.tempFormDraft = {
      title: this.data.result.title,
      description: this.data.result.description,
      schema: this.data.result.schema,
    }
    wx.redirectTo({ url: '/pages/form-editor/form-editor' })
  },
})
