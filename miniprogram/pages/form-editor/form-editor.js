const api = require('../../services/api')
const { FIELD_TYPES, OPTION_TYPES } = require('../../utils/constants')
const schema = require('../../utils/schema')

// Build a lookup: type value → Chinese label
const TYPE_LABEL_MAP = {}
FIELD_TYPES.forEach((t) => { TYPE_LABEL_MAP[t.value] = t.label })

Page({
  data: {
    formId: null,
    title: '',
    description: '',
    fields: [],
    saving: false,
  },

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
      this.loadForm(options.formId)
    } else {
      // Check if there's AI-generated data to pre-populate
      var app = getApp()
      if (app.globalData.tempFormDraft) {
        var draft = app.globalData.tempFormDraft
        app.globalData.tempFormDraft = null
        var fields = schema.schemaToFields(draft.schema)
        this.setData({
          title: draft.title || '',
          description: draft.description || '',
          fields: enrichFields(fields),
        })
      }
    }
  },

  async loadForm(formId) {
    var app = getApp()
    await app.globalData.loginReady

    wx.showLoading({ title: '加载中...' })
    try {
      var res = await api.get('/forms/' + formId)
      var fields = schema.schemaToFields(res.schema)
      this.setData({
        title: res.title,
        description: res.description,
        fields: enrichFields(fields),
      })
    } catch (err) {
      console.error('[form-editor] load error:', err)
      wx.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

  // --- Input handlers ---

  onTitleInput(e) {
    this.setData({ title: e.detail.value })
  },

  onDescInput(e) {
    this.setData({ description: e.detail.value })
  },

  onFieldLabelInput(e) {
    var idx = e.currentTarget.dataset.index
    this.setData({
      ['fields[' + idx + '].label']: e.detail.value,
    })
  },

  onFieldRequiredChange(e) {
    var idx = e.currentTarget.dataset.index
    this.setData({
      ['fields[' + idx + '].required']: e.detail.value,
    })
  },

  onOptionInput(e) {
    var fi = e.currentTarget.dataset.fieldIndex
    var oi = e.currentTarget.dataset.optIndex
    this.setData({
      ['fields[' + fi + '].options[' + oi + ']']: e.detail.value,
    })
  },

  // --- Field operations ---

  onAddField() {
    var typeLabels = FIELD_TYPES.map((t) => t.label)
    wx.showActionSheet({
      itemList: typeLabels,
      success: (res) => {
        var type = FIELD_TYPES[res.tapIndex].value
        var nextIdx = schema.nextFieldIndex(this.data.fields)
        var field = schema.newField(type, nextIdx)
        var enriched = enrichFields([field])[0]
        var fields = this.data.fields.concat([enriched])
        this.setData({ fields: fields })
      },
    })
  },

  onDeleteField(e) {
    var idx = e.currentTarget.dataset.index
    var fields = this.data.fields.slice()
    fields.splice(idx, 1)
    this.setData({ fields: fields })
  },

  onMoveUp(e) {
    var idx = e.currentTarget.dataset.index
    if (idx <= 0) return
    var fields = this.data.fields.slice()
    var tmp = fields[idx]
    fields[idx] = fields[idx - 1]
    fields[idx - 1] = tmp
    this.setData({ fields: fields })
  },

  onMoveDown(e) {
    var idx = e.currentTarget.dataset.index
    if (idx >= this.data.fields.length - 1) return
    var fields = this.data.fields.slice()
    var tmp = fields[idx]
    fields[idx] = fields[idx + 1]
    fields[idx + 1] = tmp
    this.setData({ fields: fields })
  },

  onAddOption(e) {
    var fi = e.currentTarget.dataset.fieldIndex
    var field = this.data.fields[fi]
    var opts = (field.options || []).concat(['选项' + (field.options.length + 1)])
    this.setData({
      ['fields[' + fi + '].options']: opts,
    })
  },

  onDeleteOption(e) {
    var fi = e.currentTarget.dataset.fieldIndex
    var oi = e.currentTarget.dataset.optIndex
    var opts = this.data.fields[fi].options.slice()
    opts.splice(oi, 1)
    this.setData({
      ['fields[' + fi + '].options']: opts,
    })
  },

  // --- Save ---

  async onSave() {
    var title = this.data.title.trim()
    if (!title) {
      wx.showToast({ title: '请填写表单标题', icon: 'none' })
      return
    }
    if (this.data.fields.length === 0) {
      wx.showToast({ title: '请至少添加一个字段', icon: 'none' })
      return
    }
    // Validate field labels
    for (var i = 0; i < this.data.fields.length; i++) {
      if (!this.data.fields[i].label.trim()) {
        wx.showToast({ title: '请填写第' + (i + 1) + '个字段的名称', icon: 'none' })
        return
      }
    }

    this.setData({ saving: true })

    // Strip UI-only props before sending
    var cleanFields = this.data.fields.map((f) => {
      var clean = {
        key: f.key,
        label: f.label,
        type: f.type,
        required: f.required,
      }
      if (f.hasOptions) {
        clean.options = f.options.filter((o) => o.trim() !== '')
      }
      return clean
    })

    var body = {
      title: title,
      description: this.data.description.trim(),
      schema: schema.fieldsToSchema(cleanFields),
    }

    try {
      if (this.data.formId) {
        await api.put('/forms/' + this.data.formId, body)
      } else {
        var res = await api.post('/forms', body)
        this.setData({ formId: res.id })
      }
      wx.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => { wx.navigateBack() }, 800)
    } catch (err) {
      console.error('[form-editor] save error:', err)
      wx.showToast({ title: err.message || '保存失败', icon: 'none' })
    } finally {
      this.setData({ saving: false })
    }
  },
})

/**
 * Add UI-only properties to fields for rendering.
 */
function enrichFields(fields) {
  return fields.map((f) => {
    return Object.assign({}, f, {
      typeLabel: TYPE_LABEL_MAP[f.type] || f.type,
      hasOptions: OPTION_TYPES.indexOf(f.type) !== -1,
    })
  })
}
