const api = require('../../services/api')
const { schemaToFields } = require('../../utils/schema')
const { validateForm } = require('../../utils/validator')

Page({
  data: {
    formId: null,
    loading: true,
    errorMsg: '',
    formTitle: '',
    formDesc: '',
    fields: [],
    values: {},
    errors: {},
    submitted: false,
    submitSuccess: false,
    submitting: false,
    hasBaseData: false,
    lookupKey: '',
    lookingUp: false,
  },

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
      this.init(options.formId)
    } else {
      this.setData({ loading: false, errorMsg: '缺少表单 ID' })
    }
  },

  async init(formId) {
    var app = getApp()
    await app.globalData.loginReady

    try {
      // Check if already submitted
      var mySubmission = null
      try {
        mySubmission = await api.get('/forms/' + formId + '/submissions/my')
      } catch (err) {
        // 404 means not submitted yet — that's fine
        if (err.code !== 'SUBMISSION_NOT_FOUND' && err.status !== 404) {
          throw err
        }
      }

      // Load form schema
      var form = await api.get('/forms/' + formId + '/schema')
      var fields = schemaToFields(form.schema)

      if (mySubmission && mySubmission.id) {
        // Already submitted — show read-only
        this.setData({
          loading: false,
          formTitle: form.title,
          formDesc: form.description,
          fields: fields,
          values: mySubmission.values || {},
          submitted: true,
        })
      } else {
        // Fill mode — check if base data exists for prefill
        var hasBaseData = false
        try {
          var bdRes = await api.get('/forms/' + formId + '/base-data/lookup', { row_key: '__probe__' })
          hasBaseData = true
        } catch (e) {
          // 404 is expected — but if the endpoint responds at all, base data is configured
          // We'll just try a lookup; if 404 it means no match, not "no base data"
          // A simpler heuristic: always show lookup for published forms
          hasBaseData = true
        }
        this.setData({
          loading: false,
          formTitle: form.title,
          formDesc: form.description,
          fields: fields,
          hasBaseData: hasBaseData,
        })
      }
    } catch (err) {
      console.error('[form-fill] init error:', err)
      var msg = '加载失败'
      if (err.code === 'FORBIDDEN' || err.status === 403) msg = '该表单暂不可填写'
      if (err.code === 'FORM_NOT_FOUND' || err.status === 404) msg = '表单不存在'
      this.setData({ loading: false, errorMsg: msg })
    }
  },

  onLookupInput(e) {
    this.setData({ lookupKey: e.detail.value })
  },

  async onLookup() {
    var key = this.data.lookupKey.trim()
    if (!key) {
      wx.showToast({ title: '请输入查询键', icon: 'none' })
      return
    }
    this.setData({ lookingUp: true })
    try {
      var res = await api.get('/forms/' + this.data.formId + '/base-data/lookup', { row_key: key })
      if (res.data && typeof res.data === 'object') {
        var newValues = Object.assign({}, this.data.values)
        Object.keys(res.data).forEach(function (k) {
          newValues[k] = res.data[k]
        })
        this.setData({ values: newValues })
        wx.showToast({ title: '预填充成功', icon: 'success' })
      }
    } catch (err) {
      if (err.status === 404) {
        wx.showToast({ title: '未找到匹配数据', icon: 'none' })
      } else {
        wx.showToast({ title: err.message || '查询失败', icon: 'none' })
      }
    } finally {
      this.setData({ lookingUp: false })
    }
  },

  onFieldChange(e) {
    var key = e.detail.key
    var val = e.detail.value
    this.setData({
      ['values.' + key]: val,
      ['errors.' + key]: '',
    })
  },

  async onSubmit() {
    var result = validateForm(this.data.fields, this.data.values)
    if (!result.valid) {
      this.setData({ errors: result.errors })
      wx.showToast({ title: result.errors[result.firstErrorKey], icon: 'none' })
      return
    }

    this.setData({ submitting: true, errors: {} })
    wx.showLoading({ title: '提交中...' })

    try {
      await api.post('/forms/' + this.data.formId + '/submissions', this.data.values)
      wx.hideLoading()
      this.setData({ submitting: false, submitSuccess: true })
    } catch (err) {
      wx.hideLoading()
      this.setData({ submitting: false })
      wx.showToast({ title: err.message || '提交失败', icon: 'none' })
    }
  },

  onGoBack() {
    wx.navigateBack({ fail: function () { wx.switchTab({ url: '/pages/index/index' }) } })
  },
})
