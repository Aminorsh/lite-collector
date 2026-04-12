Component({
  properties: {
    field: { type: Object, value: {} },
    value: { type: null, value: '' },
    error: { type: String, value: '' },
    readonly: { type: Boolean, value: false },
  },

  data: {
    pickerIndex: 0,
    checkedMap: {},
    displayValue: '',
  },

  observers: {
    'value, field': function (value, field) {
      // For select: find picker index
      if (field.type === 'select' && field.options && value) {
        var idx = field.options.indexOf(value)
        this.setData({ pickerIndex: idx >= 0 ? idx : 0 })
      }
      // For checkbox: build checked map and display string
      if (field.type === 'checkbox') {
        var arr = Array.isArray(value) ? value : []
        var map = {}
        arr.forEach(function (v) { map[v] = true })
        this.setData({
          checkedMap: map,
          displayValue: arr.join('、'),
        })
      }
    },
  },

  methods: {
    emitChange(val) {
      this.triggerEvent('change', { key: this.data.field.key, value: val })
    },

    onInput(e) {
      this.emitChange(e.detail.value)
    },

    onNumberInput(e) {
      var val = e.detail.value
      this.emitChange(val === '' ? '' : Number(val))
    },

    onPickerChange(e) {
      var idx = e.detail.value
      var val = this.data.field.options[idx]
      this.setData({ pickerIndex: idx })
      this.emitChange(val)
    },

    onRadioChange(e) {
      this.emitChange(e.detail.value)
    },

    onCheckboxChange(e) {
      this.emitChange(e.detail.value)
    },

    onDateChange(e) {
      this.emitChange(e.detail.value)
    },

    onChooseImage() {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (res) => {
          var tempPath = res.tempFiles[0].tempFilePath
          this.emitChange(tempPath)
        },
      })
    },
  },
})
