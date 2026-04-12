const { FORM_STATUS_LABEL, FORM_STATUS_COLOR, SUBMISSION_STATUS_LABEL, SUBMISSION_STATUS_COLOR } = require('../../utils/constants')

Component({
  properties: {
    status: {
      type: Number,
      value: 0,
    },
    // 'form' or 'submission'
    type: {
      type: String,
      value: 'form',
    },
  },

  observers: {
    'status, type': function (status, type) {
      if (type === 'submission') {
        this.setData({
          label: SUBMISSION_STATUS_LABEL[status] || '',
          color: SUBMISSION_STATUS_COLOR[status] || '#999999',
        })
      } else {
        this.setData({
          label: FORM_STATUS_LABEL[status] || '',
          color: FORM_STATUS_COLOR[status] || '#999999',
        })
      }
    },
  },

  data: {
    label: '',
    color: '#999999',
  },
})
