/**
 * Form schema manipulation utilities.
 */

const { OPTION_TYPES } = require('./constants')

/**
 * Create a new field definition with a generated key.
 * @param {string} type - Field type (text, number, select, etc.)
 * @param {number} index - 1-based index for key generation
 * @returns {Object} field definition
 */
function newField(type, index) {
  const key = 'f_' + String(index).padStart(3, '0')
  const field = {
    key: key,
    label: '',
    type: type,
    required: false,
    placeholder: '',
  }
  if (OPTION_TYPES.indexOf(type) !== -1) {
    field.options = ['选项1']
  }
  return field
}

/**
 * Parse a JSON schema string into a fields array.
 * @param {string} jsonString - The schema JSON string from the backend
 * @returns {Array} fields array
 */
function schemaToFields(jsonString) {
  if (!jsonString) return []
  try {
    var parsed = typeof jsonString === 'string' ? JSON.parse(jsonString) : jsonString
    return parsed.fields || []
  } catch (e) {
    console.error('[schema] parse error:', e)
    return []
  }
}

/**
 * Serialize a fields array into a JSON schema string.
 * @param {Array} fields - The fields array
 * @returns {string} JSON string
 */
function fieldsToSchema(fields) {
  return JSON.stringify({ fields: fields })
}

/**
 * Get the next available field index from an existing fields array.
 * @param {Array} fields - Existing fields
 * @returns {number} next 1-based index
 */
function nextFieldIndex(fields) {
  if (!fields || fields.length === 0) return 1
  var maxIndex = 0
  for (var i = 0; i < fields.length; i++) {
    var match = fields[i].key.match(/^f_(\d+)$/)
    if (match) {
      var num = parseInt(match[1], 10)
      if (num > maxIndex) maxIndex = num
    }
  }
  return maxIndex + 1
}

module.exports = {
  newField,
  schemaToFields,
  fieldsToSchema,
  nextFieldIndex,
}
