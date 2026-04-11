/**
 * Application-wide constants.
 */

// Form status values
const FORM_STATUS = {
  DRAFT: 0,
  PUBLISHED: 1,
  ARCHIVED: 2,
}

const FORM_STATUS_LABEL = {
  0: '草稿',
  1: '已发布',
  2: '已归档',
}

const FORM_STATUS_COLOR = {
  0: '#999999',   // gray
  1: '#07C160',   // green
  2: '#FFA500',   // orange
}

// Submission status values
const SUBMISSION_STATUS = {
  PROCESSING: 0,
  NORMAL: 1,
  ANOMALY: 2,
}

const SUBMISSION_STATUS_LABEL = {
  0: '处理中',
  1: '正常',
  2: '异常',
}

const SUBMISSION_STATUS_COLOR = {
  0: '#999999',   // gray
  1: '#07C160',   // green
  2: '#FA5151',   // red
}

// Supported field types
const FIELD_TYPES = [
  { value: 'text',     label: '单行文本' },
  { value: 'textarea', label: '多行文本' },
  { value: 'number',   label: '数字' },
  { value: 'select',   label: '下拉选择' },
  { value: 'radio',    label: '单选' },
  { value: 'checkbox', label: '多选' },
  { value: 'date',     label: '日期' },
  { value: 'phone',    label: '手机号' },
  { value: 'id_card',  label: '身份证号' },
  { value: 'image',    label: '图片' },
]

// Field types that require an options array
const OPTION_TYPES = ['select', 'radio', 'checkbox']

// Error codes from the backend
const ERROR_CODES = {
  BAD_REQUEST: 'BAD_REQUEST',
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  FORM_NOT_FOUND: 'FORM_NOT_FOUND',
  FORM_FORBIDDEN: 'FORM_FORBIDDEN',
  SUBMISSION_NOT_FOUND: 'SUBMISSION_NOT_FOUND',
  SUBMISSION_CREATE_FAILED: 'SUBMISSION_CREATE_FAILED',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
}

module.exports = {
  FORM_STATUS,
  FORM_STATUS_LABEL,
  FORM_STATUS_COLOR,
  SUBMISSION_STATUS,
  SUBMISSION_STATUS_LABEL,
  SUBMISSION_STATUS_COLOR,
  FIELD_TYPES,
  OPTION_TYPES,
  ERROR_CODES,
}
