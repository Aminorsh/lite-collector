/**
 * Field-level and form-level validation.
 */

/**
 * Validate a single field.
 * @param {Object} field - { key, label, type, required }
 * @param {*} value - the current value
 * @returns {string|null} error message or null if valid
 */
function validateField(field, value) {
  // Required check
  if (field.required) {
    if (value === undefined || value === null || value === '') return field.label + '不能为空'
    if (Array.isArray(value) && value.length === 0) return '请选择' + field.label
  }

  // Skip further checks if empty and not required
  if (value === undefined || value === null || value === '') return null
  if (Array.isArray(value) && value.length === 0) return null

  // Type-specific checks
  switch (field.type) {
    case 'phone':
      if (!/^1\d{10}$/.test(value)) return '请输入正确的手机号'
      break

    case 'id_card':
      if (!/^\d{17}[\dXx]$/.test(value)) return '请输入正确的身份证号'
      break

    case 'number':
      var num = Number(value)
      if (isNaN(num)) return '请输入有效数字'
      if (num < 0) return field.label + '不能为负数'
      break
  }

  return null
}

/**
 * Validate all fields in a form.
 * @param {Array} fields - fields array from schema
 * @param {Object} values - { key: value } map
 * @returns {{ valid: boolean, errors: Object, firstErrorKey: string|null }}
 */
function validateForm(fields, values) {
  var errors = {}
  var firstErrorKey = null

  for (var i = 0; i < fields.length; i++) {
    var field = fields[i]
    var err = validateField(field, values[field.key])
    if (err) {
      errors[field.key] = err
      if (!firstErrorKey) firstErrorKey = field.key
    }
  }

  return {
    valid: !firstErrorKey,
    errors: errors,
    firstErrorKey: firstErrorKey,
  }
}

module.exports = {
  validateField,
  validateForm,
}
