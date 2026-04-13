const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')
const { FIELD_TYPES } = require('../../utils/constants')

var TYPE_LABEL_MAP = {}
FIELD_TYPES.forEach(function (t) { TYPE_LABEL_MAP[t.value] = t.label })

Page({
  data: {
    description: '',
    generating: false,
    result: null,
    previewFields: [],
  },

  onDescInput(e) {
    this.setData({ description: e.detail.value })
  },

  async onGenerate() {
    var desc = this.data.description.trim()
    if (!desc) {
      wx.showToast({ title: '请输入表单描述', icon: 'none' })
      return
    }

    this.setData({ generating: true, result: null })

    try {
      var res = await api.post('/forms/generate', { description: desc })
      var fields = schemaToFields(res.schema)
      var previewFields = fields.map(function (f) {
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
        result: res,
        previewFields: previewFields,
      })
    } catch (err) {
      this.setData({ generating: false })
      var msg = err.message || 'AI 生成失败'
      if (err.status === 503) msg = 'AI 服务暂未开启，请手动创建表单'
      wx.showToast({ title: msg, icon: 'none', duration: 3000 })
    }
  },

  onRegenerate() {
    this.setData({ result: null, previewFields: [] })
  },

  onUseForm() {
    var app = getApp()
    app.globalData.tempFormDraft = {
      title: this.data.result.title,
      description: this.data.result.description,
      schema: this.data.result.schema,
    }
    wx.redirectTo({ url: '/pages/form-editor/form-editor' })
  },
})
