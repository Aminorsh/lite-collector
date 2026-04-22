/**
 * Services layer — mirrors miniprogram/services/*.js
 */
var services = (function() {

  var BASE_URL = 'http://localhost:8080/api/v1';
  var _refreshPromise = null;

  function normalizeApiPath(url) {
    if (!url || typeof url !== 'string') return url;

    // Gin routes in this backend use trailing slashes for collection resources.
    // Avoid browser-visible 301 redirects (which may not include CORS headers).
    if (url === '/forms') return '/forms/';
    if (/^\/forms\/[^/]+\/base-data$/.test(url)) return url + '/';
    if (/^\/forms\/[^/]+\/submissions$/.test(url)) return url + '/';

    return url;
  }

  // --- Storage ---
  var storage = {
    getToken: function() { return wx.getStorageSync('auth_token') || ''; },
    setToken: function(t) { wx.setStorageSync('auth_token', t); },
    getUser: function() { return wx.getStorageSync('user_info') || null; },
    setUser: function(u) { wx.setStorageSync('user_info', u); },
    clear: function() {
      wx.removeStorageSync('auth_token');
      wx.removeStorageSync('user_info');
    },
  };

  // --- API ---
  function request(options) {
    return new Promise(function(resolve, reject) {
      var header = Object.assign({ 'Content-Type': 'application/json' }, options.header || {});
      var apiPath = normalizeApiPath(options.url);
      if (!options.skipAuth) {
        var token = storage.getToken();
        if (token) header['Authorization'] = 'Bearer ' + token;
      }

      wx.request({
        url: BASE_URL + apiPath,
        method: options.method || 'GET',
        data: options.data,
        header: header,
        success: function(res) {
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve(res.data);
            return;
          }
          if (res.statusCode === 401 && !options._retried && !options.skipAuth) {
            handleUnauthorized()
              .then(function() { return request(Object.assign({}, options, { _retried: true })); })
              .then(resolve)
              .catch(reject);
            return;
          }
          var errData = res.data && res.data.error;
          var error = new Error(errData ? errData.message : 'Request failed');
          error.code = errData ? errData.code : 'UNKNOWN';
          error.status = res.statusCode;
          reject(error);
        },
        fail: function(err) {
          wx.showToast({ title: '网络连接失败', icon: 'none' });
          reject(new Error(err.errMsg || 'Network error'));
        },
      });
    });
  }

  function handleUnauthorized() {
    if (_refreshPromise) return _refreshPromise;
    _refreshPromise = auth.silentLogin().finally(function() { _refreshPromise = null; });
    return _refreshPromise;
  }

  var api = {
    get: function(url, data) { return request({ url: url, method: 'GET', data: data }); },
    post: function(url, data) { return request({ url: url, method: 'POST', data: data }); },
    put: function(url, data) { return request({ url: url, method: 'PUT', data: data }); },
    del: function(url, data) { return request({ url: url, method: 'DELETE', data: data }); },
    BASE_URL: BASE_URL,
  };

  // --- Auth ---
  var auth = {
    silentLogin: function() {
      return new Promise(function(resolve, reject) {
        wx.login({
          success: function(loginRes) {
            if (!loginRes.code) { reject(new Error('wx.login failed: no code')); return; }
            api.post('/auth/wx-login', { code: loginRes.code })
              .then(function(data) {
                storage.setToken(data.token);
                storage.setUser(data.user);
                var app = getApp();
                if (app) {
                  app.globalData.token = data.token;
                  app.globalData.userInfo = data.user;
                }
                resolve(data.user);
              })
              .catch(reject);
          },
          fail: function(err) {
            reject(new Error('wx.login failed: ' + (err.errMsg || '')));
          },
        });
      });
    },

    isLoggedIn: function() { return !!storage.getToken(); },

    logout: function() {
      storage.clear();
      var app = getApp();
      if (app) {
        app.globalData.token = '';
        app.globalData.userInfo = null;
      }
    },
  };

  // Initialize loginReady
  getApp().globalData.loginReady = Promise.resolve();

  return { storage: storage, api: api, auth: auth };
})();

// --- Utility: Constants ---
var CONSTANTS = {
  FORM_STATUS: { DRAFT: 0, PUBLISHED: 1, ARCHIVED: 2 },
  FORM_STATUS_LABEL: { 0: '草稿', 1: '已发布', 2: '已归档' },
  FORM_STATUS_COLOR: { 0: '#999999', 1: '#07C160', 2: '#FFA500' },
  SUBMISSION_STATUS: { PROCESSING: 0, NORMAL: 1, ANOMALY: 2 },
  SUBMISSION_STATUS_LABEL: { 0: '处理中', 1: '正常', 2: '异常' },
  SUBMISSION_STATUS_COLOR: { 0: '#999999', 1: '#07C160', 2: '#FA5151' },
  FIELD_TYPES: [
    { value: 'text', label: '单行文本' },
    { value: 'textarea', label: '多行文本' },
    { value: 'number', label: '数字' },
    { value: 'select', label: '下拉选择' },
    { value: 'radio', label: '单选' },
    { value: 'checkbox', label: '多选' },
    { value: 'date', label: '日期' },
    { value: 'phone', label: '手机号' },
    { value: 'id_card', label: '身份证号' },
    { value: 'image', label: '图片' },
  ],
  OPTION_TYPES: ['select', 'radio', 'checkbox'],
};

var TYPE_LABEL_MAP = {};
CONSTANTS.FIELD_TYPES.forEach(function(t) { TYPE_LABEL_MAP[t.value] = t.label; });

// --- Utility: Schema ---
var schemaUtils = {
  newField: function(type, index) {
    var key = 'f_' + String(index).padStart(3, '0');
    var field = { key: key, label: '', type: type, required: false };
    if (CONSTANTS.OPTION_TYPES.indexOf(type) !== -1) field.options = ['选项1'];
    return field;
  },

  schemaToFields: function(jsonString) {
    if (!jsonString) return [];
    try {
      var parsed = typeof jsonString === 'string' ? JSON.parse(jsonString) : jsonString;
      return parsed.fields || [];
    } catch(e) { return []; }
  },

  fieldsToSchema: function(fields) {
    return JSON.stringify({ fields: fields });
  },

  nextFieldIndex: function(fields) {
    if (!fields || fields.length === 0) return 1;
    var maxIndex = 0;
    for (var i = 0; i < fields.length; i++) {
      var match = fields[i].key.match(/^f_(\d+)$/);
      if (match) {
        var num = parseInt(match[1], 10);
        if (num > maxIndex) maxIndex = num;
      }
    }
    return maxIndex + 1;
  },
};

// --- Utility: Validator ---
var validator = {
  validateField: function(field, value) {
    if (field.required) {
      if (value === undefined || value === null || value === '') return field.label + '不能为空';
      if (Array.isArray(value) && value.length === 0) return '请选择' + field.label;
    }
    if (value === undefined || value === null || value === '') return null;
    if (Array.isArray(value) && value.length === 0) return null;
    switch (field.type) {
      case 'phone': if (!/^1\d{10}$/.test(value)) return '请输入正确的手机号'; break;
      case 'id_card': if (!/^\d{17}[\dXx]$/.test(value)) return '请输入正确的身份证号'; break;
      case 'number':
        var num = Number(value);
        if (isNaN(num)) return '请输入有效数字';
        if (num < 0) return field.label + '不能为负数';
        break;
    }
    return null;
  },

  validateForm: function(fields, values) {
    var errors = {};
    var firstErrorKey = null;
    for (var i = 0; i < fields.length; i++) {
      var field = fields[i];
      var err = validator.validateField(field, values[field.key]);
      if (err) {
        errors[field.key] = err;
        if (!firstErrorKey) firstErrorKey = field.key;
      }
    }
    return { valid: !firstErrorKey, errors: errors, firstErrorKey: firstErrorKey };
  },
};

// --- Utility: Format Time ---
function formatTime(isoStr) {
  if (!isoStr) return '';
  var d = new Date(isoStr);
  var month = d.getMonth() + 1;
  var day = d.getDate();
  var hour = d.getHours();
  var minute = d.getMinutes();
  return month + '月' + day + '日 ' +
    (hour < 10 ? '0' : '') + hour + ':' +
    (minute < 10 ? '0' : '') + minute;
}

// --- Component Helpers ---
function renderStatusBadge(status, type) {
  var label, color;
  if (type === 'submission') {
    label = CONSTANTS.SUBMISSION_STATUS_LABEL[status] || '';
    color = CONSTANTS.SUBMISSION_STATUS_COLOR[status] || '#999';
  } else {
    label = CONSTANTS.FORM_STATUS_LABEL[status] || '';
    color = CONSTANTS.FORM_STATUS_COLOR[status] || '#999';
  }
  return '<span class="badge" style="background-color:' + color + ';"><span class="badge-text">' + escHtml(label) + '</span></span>';
}

function renderEmptyState(icon, text, hint) {
  return '<div class="empty-state">' +
    '<div class="empty-icon">' + icon + '</div>' +
    '<div class="empty-text">' + escHtml(text) + '</div>' +
    (hint ? '<div class="empty-hint">' + escHtml(hint) + '</div>' : '') +
    '</div>';
}

function escHtml(str) {
  if (!str) return '';
  return String(str).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function renderFieldRenderer(fields, values, errors, opts) {
  var readonly = opts && opts.readonly;
  var onChange = opts && opts.onChange;
  var html = '';

  fields.forEach(function(field) {
    var val = values[field.key];
    var err = errors ? errors[field.key] : '';
    html += '<div class="field-item" data-key="' + field.key + '">';
    html += '<div class="field-label">' + escHtml(field.label);
    if (field.required) html += '<span class="required-mark">*</span>';
    html += '</div>';

    if (readonly) {
      html += '<div class="field-value-readonly">';
      if (field.type === 'checkbox') {
        html += escHtml(Array.isArray(val) ? val.join('、') : (val || '未填写'));
      } else if (field.type === 'image' && val) {
        html += '<img class="image-preview" src="' + escHtml(val) + '">';
      } else if (field.type === 'image') {
        html += '<span class="text-secondary">未上传</span>';
      } else {
        html += escHtml(val !== undefined && val !== '' && val !== null ? val : '未填写');
      }
      html += '</div>';
    } else {
      var inputId = 'field-' + field.key;
      var genericPh = field.placeholder || ('请输入' + field.label);
      var phonePh = field.placeholder || '请输入手机号';
      var idPh = field.placeholder || '请输入身份证号';
      switch (field.type) {
        case 'text':
          html += '<input class="field-input" id="' + inputId + '" placeholder="' + escHtml(genericPh) + '" value="' + escHtml(val || '') + '">';
          break;
        case 'textarea':
          html += '<textarea class="field-textarea" id="' + inputId + '" placeholder="' + escHtml(genericPh) + '">' + escHtml(val || '') + '</textarea>';
          break;
        case 'number':
          html += '<input class="field-input" id="' + inputId + '" type="number" step="any" placeholder="' + escHtml(genericPh) + '" value="' + escHtml(val != null ? val : '') + '">';
          break;
        case 'phone':
          html += '<input class="field-input" id="' + inputId + '" type="tel" maxlength="11" placeholder="' + escHtml(phonePh) + '" value="' + escHtml(val || '') + '">';
          break;
        case 'id_card':
          html += '<input class="field-input" id="' + inputId + '" maxlength="18" placeholder="' + escHtml(idPh) + '" value="' + escHtml(val || '') + '">';
          break;
        case 'select':
          html += '<select class="field-input" id="' + inputId + '">';
          html += '<option value="">请选择</option>';
          (field.options || []).forEach(function(opt) {
            html += '<option value="' + escHtml(opt) + '"' + (val === opt ? ' selected' : '') + '>' + escHtml(opt) + '</option>';
          });
          html += '</select>';
          break;
        case 'radio':
          html += '<div class="option-group">';
          (field.options || []).forEach(function(opt) {
            html += '<label class="option-item"><input type="radio" name="' + inputId + '" value="' + escHtml(opt) + '"' + (val === opt ? ' checked' : '') + '> ' + escHtml(opt) + '</label>';
          });
          html += '</div>';
          break;
        case 'checkbox':
          var checkedArr = Array.isArray(val) ? val : [];
          html += '<div class="option-group">';
          (field.options || []).forEach(function(opt) {
            html += '<label class="option-item"><input type="checkbox" name="' + inputId + '" value="' + escHtml(opt) + '"' + (checkedArr.indexOf(opt) !== -1 ? ' checked' : '') + '> ' + escHtml(opt) + '</label>';
          });
          html += '</div>';
          break;
        case 'date':
          html += '<input class="field-input" id="' + inputId + '" type="date" value="' + escHtml(val || '') + '">';
          break;
        case 'image':
          if (val) {
            html += '<div class="image-area"><img class="image-preview" src="' + escHtml(val) + '" id="' + inputId + '-preview"><input type="file" accept="image/*" id="' + inputId + '" style="display:none;"></div>';
          } else {
            html += '<div class="image-area"><div class="image-upload-btn" id="' + inputId + '-btn"><span class="image-upload-icon">+</span><span class="image-upload-text">选择图片</span></div><input type="file" accept="image/*" id="' + inputId + '" style="display:none;"></div>';
          }
          break;
        default:
          html += '<input class="field-input" id="' + inputId + '" value="' + escHtml(val || '') + '">';
      }
    }

    if (err) html += '<span class="field-error">' + escHtml(err) + '</span>';
    html += '</div>';
  });

  return html;
}

function bindFieldEvents(container, fields, values, onChange) {
  fields.forEach(function(field) {
    var inputId = 'field-' + field.key;
    switch (field.type) {
      case 'text': case 'textarea': case 'phone': case 'id_card': case 'date':
        var el = container.querySelector('#' + inputId);
        if (el) el.addEventListener('input', function() { onChange(field.key, el.value); });
        if (el && field.type === 'date') el.addEventListener('change', function() { onChange(field.key, el.value); });
        break;
      case 'number':
        var nEl = container.querySelector('#' + inputId);
        if (nEl) nEl.addEventListener('input', function() { onChange(field.key, nEl.value === '' ? '' : Number(nEl.value)); });
        break;
      case 'select':
        var sEl = container.querySelector('#' + inputId);
        if (sEl) sEl.addEventListener('change', function() { onChange(field.key, sEl.value); });
        break;
      case 'radio':
        container.querySelectorAll('input[name="' + inputId + '"]').forEach(function(r) {
          r.addEventListener('change', function() { onChange(field.key, r.value); });
        });
        break;
      case 'checkbox':
        container.querySelectorAll('input[name="' + inputId + '"]').forEach(function(c) {
          c.addEventListener('change', function() {
            var checked = [];
            container.querySelectorAll('input[name="' + inputId + '"]:checked').forEach(function(x) { checked.push(x.value); });
            onChange(field.key, checked);
          });
        });
        break;
      case 'image':
        var fileInput = container.querySelector('#' + inputId);
        var btn = container.querySelector('#' + inputId + '-btn');
        var preview = container.querySelector('#' + inputId + '-preview');
        var trigger = btn || preview;
        if (trigger && fileInput) {
          trigger.addEventListener('click', function() { fileInput.click(); });
          fileInput.addEventListener('change', function() {
            if (fileInput.files && fileInput.files[0]) {
              var url = URL.createObjectURL(fileInput.files[0]);
              onChange(field.key, url);
            }
          });
        }
        break;
    }
  });
}
