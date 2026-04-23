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
    submitSuccess: false,
    submitting: false,
    prefilled: false,
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
      // Load schema first so the UI structure is known.
      var form = await api.get('/forms/' + formId + '/schema')
      var fields = schemaToFields(form.schema)

      // Try to prefill from this user's previous submission.
      var prefillValues = {}
      var prefilled = false
      try {
        var mySubmission = await api.get('/forms/' + formId + '/submissions/my')
        if (mySubmission && mySubmission.values) {
          prefillValues = mySubmission.values
          prefilled = Object.keys(prefillValues).length > 0
        }
      } catch (err) {
        if (err.code !== 'SUBMISSION_NOT_FOUND' && err.status !== 404) {
          throw err
        }
      }

      this.setData({
        loading: false,
        formTitle: form.title,
        formDesc: form.description,
        fields: fields,
        values: prefillValues,
        prefilled: prefilled,
      })
    } catch (err) {
      console.error('[form-fill] init error:', err)
      var msg = '加载失败'
      if (err.code === 'FORBIDDEN' || err.status === 403) msg = '该表单暂不可填写'
      if (err.code === 'FORM_NOT_FOUND' || err.status === 404) msg = '表单不存在'
      this.setData({ loading: false, errorMsg: msg })
    }
  },

  onClearPrefill() {
    this.setData({ values: {}, prefilled: false, errors: {} })
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

  onShareAppMessage() {
    return {
      title: this.data.formTitle || '数据收集表单',
      path: '/pages/form-fill/form-fill?formId=' + this.data.formId,
    }
  },
})
