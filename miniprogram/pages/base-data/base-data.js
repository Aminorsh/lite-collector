const api = require('../../services/api')

Page({
  data: {
    formId: null,
    loading: true,
    rows: [],
    showImport: false,
    importText: '',
    importing: false,
  },

  onLoad(options) {
    if (options.formId) {
      this.setData({ formId: options.formId })
      this.loadData()
    }
  },

  async loadData() {
    var app = getApp()
    await app.globalData.loginReady
    this.setData({ loading: true })

    try {
      var res = await api.get('/forms/' + this.data.formId + '/base-data')
      var rows = (res.rows || []).map(function (r) {
        var preview = ''
        if (r.data && typeof r.data === 'object') {
          var vals = Object.values(r.data)
          preview = vals.slice(0, 3).join(', ')
          if (vals.length > 3) preview += '...'
        }
        return { id: r.id, row_key: r.row_key, data: r.data, preview: preview }
      })
      this.setData({ rows: rows, loading: false })
    } catch (err) {
      console.error('[base-data] load error:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onShowImport() {
    this.setData({ showImport: true, importText: '' })
  },

  onCancelImport() {
    this.setData({ showImport: false, importText: '' })
  },

  onImportInput(e) {
    this.setData({ importText: e.detail.value })
  },

  async onDoImport() {
    var text = this.data.importText.trim()
    if (!text) {
      wx.showToast({ title: '请输入数据', icon: 'none' })
      return
    }

    var parsed
    try {
      parsed = JSON.parse(text)
    } catch (e) {
      wx.showToast({ title: 'JSON 格式错误', icon: 'none' })
      return
    }

    if (!Array.isArray(parsed) || parsed.length === 0) {
      wx.showToast({ title: '数据应为非空数组', icon: 'none' })
      return
    }

    // Validate structure
    for (var i = 0; i < parsed.length; i++) {
      if (!parsed[i].row_key || !parsed[i].data) {
        wx.showToast({ title: '第' + (i + 1) + '项缺少 row_key 或 data', icon: 'none' })
        return
      }
    }

    this.setData({ importing: true })
    try {
      var res = await api.post('/forms/' + this.data.formId + '/base-data', { rows: parsed })
      wx.showToast({ title: '成功导入 ' + res.imported + ' 条', icon: 'success' })
      this.setData({ showImport: false, importText: '', importing: false })
      this.loadData()
    } catch (err) {
      this.setData({ importing: false })
      wx.showToast({ title: err.message || '导入失败', icon: 'none' })
    }
  },

  onClearAll() {
    wx.showModal({
      title: '确认清空',
      content: '确定清空所有底表数据吗？此操作不可恢复。',
      success: (res) => {
        if (res.confirm) this.doClear()
      },
    })
  },

  async doClear() {
    wx.showLoading({ title: '清空中...' })
    try {
      await api.del('/forms/' + this.data.formId + '/base-data')
      wx.showToast({ title: '已清空', icon: 'success' })
      this.loadData()
    } catch (err) {
      wx.showToast({ title: err.message || '清空失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },
})
