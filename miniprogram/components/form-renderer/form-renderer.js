Component({
  properties: {
    fields: { type: Array, value: [] },
    values: { type: Object, value: {} },
    errors: { type: Object, value: {} },
    readonly: { type: Boolean, value: false },
  },

  methods: {
    onFieldChange(e) {
      var detail = e.detail
      // Bubble up to parent page
      this.triggerEvent('change', detail)
    },
  },
})
