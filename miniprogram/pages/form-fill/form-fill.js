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
        // Fill mode
        this.setData({
          loading: false,
          formTitle: form.title,
          formDesc: form.description,
          fields: fields,
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
