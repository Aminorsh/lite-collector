/**
 * Page renderers — each function mirrors one mini program page.
 * Signature: pages.pageName(rootEl, params)
 */
var pages = {};

// ==================== Index ====================
pages['index'] = function(root, params) {
  var FILTER_KEY = 'formListFilter';
  var STATUS_TABS = [
    { label: '全部', value: '' },
    { label: '草稿', value: '0' },
    { label: '已发布', value: '1' },
    { label: '已归档', value: '2' },
  ];
  var SORT_OPTIONS = [
    { label: '最近更新', sort: 'updated_at', order: 'desc' },
    { label: '最近创建', sort: 'created_at', order: 'desc' },
    { label: '标题 A→Z', sort: 'title', order: 'asc' },
  ];

  var saved = wx.getStorageSync(FILTER_KEY) || {};
  var state = {
    forms: [],
    loading: true,
    query: saved.query || '',
    status: saved.status || '',
    sortIndex: typeof saved.sortIndex === 'number' ? saved.sortIndex : 0,
  };
  var debounceTimer = null;

  function persist() {
    wx.setStorageSync(FILTER_KEY, {
      query: state.query, status: state.status, sortIndex: state.sortIndex,
    });
  }

  async function loadForms() {
    state.loading = true;
    render();
    try {
      var sort = SORT_OPTIONS[state.sortIndex];
      var params = { sort: sort.sort, order: sort.order };
      if (state.query) params.q = state.query;
      if (state.status) params.status = state.status;
      var res = await services.api.get('/forms', params);
      state.forms = (res.forms || []).map(function(f) {
        return Object.assign({}, f, { updatedAtText: formatTime(f.updated_at || f.created_at) });
      });
    } catch(e) {
      console.error('[index] loadForms error:', e);
      state.forms = [];
    }
    state.loading = false;
    render();
  }

  function renderFilterBar() {
    var html = '<div class="filter-bar">';
    html += '<div class="search-row">';
    html += '<input class="search-input" id="f-query" placeholder="搜索表单标题" value="' + escHtml(state.query) + '" />';
    html += '<select class="sort-select" id="f-sort">';
    SORT_OPTIONS.forEach(function(opt, i) {
      html += '<option value="' + i + '"' + (i === state.sortIndex ? ' selected' : '') + '>' + escHtml(opt.label) + '</option>';
    });
    html += '</select>';
    html += '</div>';
    html += '<div class="status-tabs">';
    STATUS_TABS.forEach(function(t) {
      html += '<div class="status-tab' + (state.status === t.value ? ' active' : '') + '" data-value="' + escHtml(t.value) + '">' + escHtml(t.label) + '</div>';
    });
    html += '</div></div>';
    return html;
  }

  function render() {
    var html = renderFilterBar();

    if (state.loading) {
      html += '<div class="loading-state"><span class="text-secondary">加载中...</span></div>';
    } else if (state.forms.length === 0) {
      html += renderEmptyState('📋', '暂无表单', '点击右下方按钮创建第一个表单');
    } else {
      html += '<div class="form-list">';
      state.forms.forEach(function(f) {
        html += '<div class="form-card" data-id="' + f.id + '">';
        html += '<div class="form-card-header">';
        html += '<span class="form-title">' + escHtml(f.title) + '</span>';
        html += renderStatusBadge(f.status, 'form');
        html += '</div>';
        if (f.description) {
          html += '<div class="form-card-desc"><span class="text-secondary text-sm">' + escHtml(f.description) + '</span></div>';
        }
        html += '<div class="form-card-footer"><span class="text-secondary text-sm">' + escHtml(f.updatedAtText) + '</span></div>';
        html += '</div>';
      });
      html += '</div>';
    }
    html += '<div class="fab" id="fab-create"><span class="fab-icon">+</span></div>';
    root.innerHTML = html;

    var queryEl = document.getElementById('f-query');
    if (queryEl) queryEl.addEventListener('input', function() {
      state.query = queryEl.value;
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(function() { persist(); loadForms(); }, 300);
    });

    var sortEl = document.getElementById('f-sort');
    if (sortEl) sortEl.addEventListener('change', function() {
      state.sortIndex = Number(sortEl.value);
      persist(); loadForms();
    });

    root.querySelectorAll('.status-tab').forEach(function(el) {
      el.addEventListener('click', function() {
        var v = el.dataset.value;
        if (v === state.status) return;
        state.status = v;
        persist(); loadForms();
      });
    });

    root.querySelectorAll('.form-card').forEach(function(el) {
      el.addEventListener('click', function() {
        wx.navigateTo({ url: '/pages/form-detail/form-detail?formId=' + el.dataset.id });
      });
    });
    document.getElementById('fab-create').addEventListener('click', function() {
      wx.showActionSheet({
        itemList: ['手动创建', 'AI 创建'],
        success: function(res) {
          if (res.tapIndex === 0) wx.navigateTo({ url: '/pages/form-editor/form-editor' });
          else if (res.tapIndex === 1) wx.navigateTo({ url: '/pages/ai-generate/ai-generate' });
        },
      });
    });
  }

  loadForms();
};

// ==================== Profile ====================
pages['profile'] = function(root, params) {
  var state = {
    userInfo: services.storage.getUser(),
    editing: false,
    editNickname: '',
    saving: false,
  };

  function render() {
    var u = state.userInfo;
    var html = '';
    if (u) {
      html += '<div class="profile-card">';
      html += '<img class="avatar" src="' + escHtml(u.avatar_url || '') + '" onerror="this.style.display=\'none\'">';
      html += '<div class="user-name">' + escHtml(u.nickname || '微信用户') + '</div>';
      html += '<div class="user-id text-secondary text-sm">ID: ' + escHtml(u.id) + '</div>';
      html += '</div>';
      // Edit nickname menu
      html += '<div class="menu-section"><div class="menu-item" id="edit-nickname-btn">';
      html += '<span>修改昵称</span>';
      html += '<div class="menu-right"><span class="menu-value text-secondary text-sm">' + escHtml(u.nickname || '未设置') + '</span><span class="menu-arrow">›</span></div>';
      html += '</div></div>';
    } else {
      html += '<div class="profile-card"><div class="avatar"></div><div class="user-name text-secondary">未登录</div></div>';
    }

    if (state.editing) {
      html += '<div class="edit-panel">';
      html += '<div class="edit-title">修改昵称</div>';
      html += '<input class="edit-input" id="nickname-input" placeholder="请输入新昵称" value="' + escHtml(state.editNickname) + '" maxlength="20">';
      html += '<div class="edit-actions">';
      html += '<button class="edit-cancel" id="cancel-edit">取消</button>';
      html += '<button class="btn-primary" id="save-nickname"' + (state.saving ? ' disabled' : '') + '>' + (state.saving ? '保存中...' : '保存') + '</button>';
      html += '</div></div>';
    }

    html += '<div class="menu-section"><div class="menu-item" id="logout-btn"><span>退出登录</span><span class="menu-arrow">›</span></div></div>';

    root.innerHTML = html;

    // Bind
    var editBtn = document.getElementById('edit-nickname-btn');
    if (editBtn) editBtn.addEventListener('click', function() {
      state.editing = true;
      state.editNickname = state.userInfo ? state.userInfo.nickname || '' : '';
      render();
    });

    var cancelBtn = document.getElementById('cancel-edit');
    if (cancelBtn) cancelBtn.addEventListener('click', function() {
      state.editing = false;
      render();
    });

    var saveBtn = document.getElementById('save-nickname');
    if (saveBtn) saveBtn.addEventListener('click', async function() {
      var input = document.getElementById('nickname-input');
      var nickname = (input.value || '').trim();
      if (!nickname) { wx.showToast({ title: '昵称不能为空', icon: 'none' }); return; }
      state.saving = true;
      render();
      try {
        var res = await services.api.put('/user/profile', { nickname: nickname, avatar_url: '' });
        var user = services.storage.getUser();
        user.nickname = res.nickname || nickname;
        services.storage.setUser(user);
        getApp().globalData.userInfo = user;
        state.userInfo = user;
        state.saving = false;
        state.editing = false;
        wx.showToast({ title: '修改成功', icon: 'success' });
      } catch(e) {
        state.saving = false;
        wx.showToast({ title: e.message || '修改失败', icon: 'none' });
      }
      render();
    });

    var logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) logoutBtn.addEventListener('click', function() {
      wx.showModal({
        title: '提示',
        content: '确定退出登录吗？',
        success: function(res) {
          if (res.confirm) {
            services.auth.logout();
            state.userInfo = null;
            services.auth.silentLogin().then(function() {
              state.userInfo = services.storage.getUser();
              render();
            });
            render();
          }
        },
      });
    });
  }

  render();
};

// ==================== Form Editor ====================
pages['form-editor'] = function(root, params) {
  var state = {
    formId: params.formId || null,
    title: '',
    description: '',
    fields: [],
    saving: false,
  };

  function enrichFields(fields) {
    return fields.map(function(f) {
      return Object.assign({}, f, {
        typeLabel: TYPE_LABEL_MAP[f.type] || f.type,
        hasOptions: CONSTANTS.OPTION_TYPES.indexOf(f.type) !== -1,
      });
    });
  }

  async function init() {
    if (state.formId) {
      wx.showLoading({ title: '加载中...' });
      try {
        var res = await services.api.get('/forms/' + state.formId);
        var fields = schemaUtils.schemaToFields(res.schema);
        state.title = res.title;
        state.description = res.description;
        state.fields = enrichFields(fields);
      } catch(e) {
        wx.showToast({ title: '加载失败', icon: 'none' });
      }
      wx.hideLoading();
    } else {
      var app = getApp();
      if (app.globalData.tempFormDraft) {
        var draft = app.globalData.tempFormDraft;
        app.globalData.tempFormDraft = null;
        var fields = schemaUtils.schemaToFields(draft.schema);
        state.title = draft.title || '';
        state.description = draft.description || '';
        state.fields = enrichFields(fields);
      }
    }
    render();
  }

  function render() {
    var html = '';
    // Meta section
    html += '<div class="section">';
    html += '<div class="section-title">基本信息</div>';
    html += '<div class="input-group"><input class="input" id="form-title" placeholder="表单标题（必填）" value="' + escHtml(state.title) + '"></div>';
    html += '<div class="input-group"><textarea class="textarea" id="form-desc" placeholder="表单描述（选填）" maxlength="500">' + escHtml(state.description) + '</textarea></div>';
    html += '</div>';

    // Fields
    html += '<div class="section">';
    html += '<div class="section-header"><span class="section-title">字段列表</span><span class="field-count text-secondary text-sm">' + state.fields.length + ' 个字段</span></div>';

    state.fields.forEach(function(field, idx) {
      html += '<div class="field-card">';
      html += '<div class="field-header"><span class="field-type-label">' + escHtml(field.typeLabel) + '</span>';
      html += '<div class="field-actions">';
      if (idx > 0) html += '<button class="action-btn" data-action="up" data-index="' + idx + '">↑</button>';
      if (idx < state.fields.length - 1) html += '<button class="action-btn" data-action="down" data-index="' + idx + '">↓</button>';
      html += '<button class="action-btn action-delete" data-action="delete" data-index="' + idx + '">删除</button>';
      html += '</div></div>';

      html += '<div class="input-group"><input class="input field-label-input" data-index="' + idx + '" placeholder="字段名称" value="' + escHtml(field.label) + '"></div>';

      html += '<div class="toggle-row"><span>必填</span><input type="checkbox" class="field-required-toggle" data-index="' + idx + '"' + (field.required ? ' checked' : '') + '></div>';

      if (field.hasOptions) {
        html += '<div class="options-section"><div class="options-title">选项</div>';
        (field.options || []).forEach(function(opt, optIdx) {
          html += '<div class="option-row">';
          html += '<input class="input option-input" data-field-index="' + idx + '" data-opt-index="' + optIdx + '" value="' + escHtml(opt) + '" placeholder="选项内容">';
          if (field.options.length > 1) {
            html += '<button class="action-btn action-delete" data-action="delete-option" data-field-index="' + idx + '" data-opt-index="' + optIdx + '">×</button>';
          }
          html += '</div>';
        });
        html += '<div class="add-option-btn" data-action="add-option" data-field-index="' + idx + '">+ 添加选项</div>';
        html += '</div>';
      }
      html += '</div>';
    });
    html += '</div>';

    html += '<div class="add-field-btn" id="add-field">+ 添加字段</div>';
    html += '<div class="save-area"><button class="btn-primary save-btn" id="save-form"' + (state.saving ? ' disabled' : '') + '>' + (state.saving ? '保存中...' : '保存草稿') + '</button></div>';

    root.innerHTML = html;
    bindEvents();
  }

  function bindEvents() {
    document.getElementById('form-title').addEventListener('input', function(e) { state.title = e.target.value; });
    document.getElementById('form-desc').addEventListener('input', function(e) { state.description = e.target.value; });

    root.querySelectorAll('.field-label-input').forEach(function(el) {
      el.addEventListener('input', function() {
        state.fields[parseInt(el.dataset.index)].label = el.value;
      });
    });

    root.querySelectorAll('.field-required-toggle').forEach(function(el) {
      el.addEventListener('change', function() {
        state.fields[parseInt(el.dataset.index)].required = el.checked;
      });
    });

    root.querySelectorAll('.option-input').forEach(function(el) {
      el.addEventListener('input', function() {
        var fi = parseInt(el.dataset.fieldIndex);
        var oi = parseInt(el.dataset.optIndex);
        state.fields[fi].options[oi] = el.value;
      });
    });

    root.querySelectorAll('[data-action]').forEach(function(el) {
      el.addEventListener('click', function() {
        var action = el.dataset.action;
        var idx = parseInt(el.dataset.index);
        var fi = parseInt(el.dataset.fieldIndex);
        var oi = parseInt(el.dataset.optIndex);

        switch (action) {
          case 'up':
            if (idx > 0) { var tmp = state.fields[idx]; state.fields[idx] = state.fields[idx-1]; state.fields[idx-1] = tmp; }
            render(); break;
          case 'down':
            if (idx < state.fields.length - 1) { var tmp = state.fields[idx]; state.fields[idx] = state.fields[idx+1]; state.fields[idx+1] = tmp; }
            render(); break;
          case 'delete':
            state.fields.splice(idx, 1);
            render(); break;
          case 'delete-option':
            state.fields[fi].options.splice(oi, 1);
            render(); break;
          case 'add-option':
            state.fields[fi].options.push('选项' + (state.fields[fi].options.length + 1));
            render(); break;
        }
      });
    });

    document.getElementById('add-field').addEventListener('click', function() {
      var typeLabels = CONSTANTS.FIELD_TYPES.map(function(t) { return t.label; });
      wx.showActionSheet({
        itemList: typeLabels,
        success: function(res) {
          var type = CONSTANTS.FIELD_TYPES[res.tapIndex].value;
          var nextIdx = schemaUtils.nextFieldIndex(state.fields);
          var field = schemaUtils.newField(type, nextIdx);
          state.fields.push(enrichFields([field])[0]);
          render();
        },
      });
    });

    document.getElementById('save-form').addEventListener('click', async function() {
      var title = state.title.trim();
      if (!title) { wx.showToast({ title: '请填写表单标题', icon: 'none' }); return; }
      if (state.fields.length === 0) { wx.showToast({ title: '请至少添加一个字段', icon: 'none' }); return; }
      for (var i = 0; i < state.fields.length; i++) {
        if (!state.fields[i].label.trim()) {
          wx.showToast({ title: '请填写第' + (i+1) + '个字段的名称', icon: 'none' }); return;
        }
      }

      state.saving = true;
      render();

      var cleanFields = state.fields.map(function(f) {
        var clean = { key: f.key, label: f.label, type: f.type, required: f.required };
        if (f.hasOptions) clean.options = f.options.filter(function(o) { return o.trim() !== ''; });
        return clean;
      });
      var body = { title: title, description: state.description.trim(), schema: schemaUtils.fieldsToSchema(cleanFields) };

      try {
        if (state.formId) {
          await services.api.put('/forms/' + state.formId, body);
        } else {
          var res = await services.api.post('/forms', body);
          state.formId = res.id;
        }
        wx.showToast({ title: '保存成功', icon: 'success' });
        setTimeout(function() { router.back(); }, 800);
      } catch(e) {
        wx.showToast({ title: e.message || '保存失败', icon: 'none' });
      }
      state.saving = false;
      render();
    });
  }

  init();
};

// ==================== Form Detail ====================
pages['form-detail'] = function(root, params) {
  var formId = params.formId;
  if (!formId) { root.innerHTML = '<div class="loading-center"><span class="text-secondary">缺少表单 ID</span></div>'; return; }

  var state = { form: null, fieldCount: 0, createdAtText: '', updatedAtText: '' };

  async function loadForm() {
    try {
      var res = await services.api.get('/forms/' + formId);
      var fields = schemaUtils.schemaToFields(res.schema);
      state.form = res;
      state.fieldCount = fields.length;
      state.createdAtText = formatTime(res.created_at);
      state.updatedAtText = formatTime(res.updated_at);
    } catch(e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
    render();
  }

  function render() {
    if (!state.form) { root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载中...</span></div>'; return; }
    var f = state.form;
    var html = '<div class="detail-card">';
    html += '<div class="detail-header"><span class="detail-title">' + escHtml(f.title) + '</span>' + renderStatusBadge(f.status, 'form') + '</div>';
    if (f.description) html += '<span class="detail-desc">' + escHtml(f.description) + '</span>';
    html += '<div class="detail-meta"><span class="text-secondary text-sm">创建于 ' + escHtml(state.createdAtText) + '</span><span class="text-secondary text-sm">更新于 ' + escHtml(state.updatedAtText) + '</span></div>';
    html += '<div class="detail-stats"><span class="text-secondary text-sm">' + state.fieldCount + ' 个字段</span></div>';
    html += '</div>';

    html += '<div class="action-section">';
    if (f.status === 0) {
      html += '<button class="action-btn-full" id="btn-edit">编辑表单</button>';
      html += '<button class="action-btn-full" id="btn-base-data">底表数据</button>';
      html += '<button class="action-btn-full action-primary" id="btn-publish">发布表单</button>';
    } else if (f.status === 1) {
      html += '<button class="action-btn-full action-primary" id="btn-fill">填写表单（预览）</button>';
      html += '<button class="action-btn-full" id="btn-submissions">查看提交</button>';
      html += '<button class="action-btn-full" id="btn-base-data">底表数据</button>';
      html += '<button class="action-btn-full" id="btn-report">生成报告</button>';
      html += '<button class="action-btn-full action-warn" id="btn-archive">归档表单</button>';
    } else if (f.status === 2) {
      html += '<button class="action-btn-full" id="btn-submissions">查看提交</button>';
      html += '<button class="action-btn-full" id="btn-report">生成报告</button>';
    }
    html += '</div>';

    root.innerHTML = html;

    // Bind
    var editBtn = document.getElementById('btn-edit');
    if (editBtn) editBtn.addEventListener('click', function() { wx.navigateTo({ url: '/pages/form-editor/form-editor?formId=' + formId }); });

    var publishBtn = document.getElementById('btn-publish');
    if (publishBtn) publishBtn.addEventListener('click', function() {
      wx.showModal({ title: '确认发布', content: '发布后将开放填写，确定发布吗？', success: async function(res) {
        if (res.confirm) {
          wx.showLoading({ title: '发布中...' });
          try { await services.api.post('/forms/' + formId + '/publish'); wx.showToast({ title: '发布成功', icon: 'success' }); loadForm(); }
          catch(e) { wx.showToast({ title: e.message || '发布失败', icon: 'none' }); }
          wx.hideLoading();
        }
      }});
    });

    var archiveBtn = document.getElementById('btn-archive');
    if (archiveBtn) archiveBtn.addEventListener('click', function() {
      wx.showModal({ title: '确认归档', content: '归档后将不再接受新提交，确定归档吗？', success: async function(res) {
        if (res.confirm) {
          wx.showLoading({ title: '归档中...' });
          try { await services.api.post('/forms/' + formId + '/archive'); wx.showToast({ title: '已归档', icon: 'success' }); loadForm(); }
          catch(e) { wx.showToast({ title: e.message || '归档失败', icon: 'none' }); }
          wx.hideLoading();
        }
      }});
    });

    var subBtn = document.getElementById('btn-submissions');
    if (subBtn) subBtn.addEventListener('click', function() { wx.navigateTo({ url: '/pages/submissions/submissions?formId=' + formId }); });

    var bdBtn = document.getElementById('btn-base-data');
    if (bdBtn) bdBtn.addEventListener('click', function() { wx.navigateTo({ url: '/pages/base-data/base-data?formId=' + formId }); });

    var reportBtn = document.getElementById('btn-report');
    if (reportBtn) reportBtn.addEventListener('click', function() { wx.navigateTo({ url: '/pages/report/report?formId=' + formId }); });

    var fillBtn = document.getElementById('btn-fill');
    if (fillBtn) fillBtn.addEventListener('click', function() { wx.navigateTo({ url: '/pages/form-fill/form-fill?formId=' + formId }); });
  }

  loadForm();
};

// ==================== Form Fill ====================
pages['form-fill'] = function(root, params) {
  var formId = params.formId;
  if (!formId) { root.innerHTML = '<div class="error-state"><span class="text-secondary">缺少表单 ID</span></div>'; return; }

  var state = {
    loading: true, errorMsg: '', formTitle: '', formDesc: '',
    fields: [], values: {}, errors: {},
    submitted: false, submitSuccess: false, submitting: false,
    hasBaseData: true, lookupKey: '', lookingUp: false,
  };

  async function init() {
    try {
      var mySubmission = null;
      try { mySubmission = await services.api.get('/forms/' + formId + '/submissions/my'); }
      catch(e) { if (e.code !== 'SUBMISSION_NOT_FOUND' && e.status !== 404) throw e; }

      var form = await services.api.get('/forms/' + formId + '/schema');
      var fields = schemaUtils.schemaToFields(form.schema);

      if (mySubmission && mySubmission.id) {
        state.loading = false;
        state.formTitle = form.title;
        state.formDesc = form.description;
        state.fields = fields;
        state.values = mySubmission.values || {};
        state.submitted = true;
      } else {
        state.loading = false;
        state.formTitle = form.title;
        state.formDesc = form.description;
        state.fields = fields;
      }
    } catch(e) {
      state.loading = false;
      var msg = '加载失败';
      if (e.code === 'FORBIDDEN' || e.status === 403) msg = '该表单暂不可填写';
      if (e.code === 'FORM_NOT_FOUND' || e.status === 404) msg = '表单不存在';
      state.errorMsg = msg;
    }
    render();
  }

  function render() {
    if (state.loading) { root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载中...</span></div>'; return; }
    if (state.errorMsg) { root.innerHTML = '<div class="error-state"><span class="text-secondary">' + escHtml(state.errorMsg) + '</span></div>'; return; }

    var html = '';

    if (state.submitted) {
      html += '<div class="success-banner"><span class="success-icon">✓</span><span class="success-text">已提交</span></div>';
      html += '<div class="form-header-card"><span class="form-title" style="display:block;margin-bottom:8px;">' + escHtml(state.formTitle) + '</span>';
      if (state.formDesc) html += '<span class="text-secondary">' + escHtml(state.formDesc) + '</span>';
      html += '</div>';
      html += '<div class="form-body">' + renderFieldRenderer(state.fields, state.values, {}, { readonly: true }) + '</div>';
      root.innerHTML = html;
      return;
    }

    if (state.submitSuccess) {
      html += '<div class="success-full">';
      html += '<div class="success-full-icon">✓</div>';
      html += '<div class="success-full-title">提交成功</div>';
      html += '<div class="success-full-hint">感谢您的填写</div>';
      html += '<button class="btn-primary mt-lg" id="go-back" style="width:auto;padding:10px 40px;">返回</button>';
      html += '</div>';
      root.innerHTML = html;
      document.getElementById('go-back').addEventListener('click', function() { router.back(); });
      return;
    }

    // Fill mode
    html += '<div class="form-header-card"><span class="form-title" style="display:block;margin-bottom:8px;">' + escHtml(state.formTitle) + '</span>';
    if (state.formDesc) html += '<span class="text-secondary">' + escHtml(state.formDesc) + '</span>';
    html += '</div>';

    if (state.hasBaseData) {
      html += '<div class="lookup-card"><span class="lookup-title">查询预填</span>';
      html += '<div class="lookup-row">';
      html += '<input class="lookup-input" id="lookup-key" placeholder="输入查询键（如工号）" value="' + escHtml(state.lookupKey) + '">';
      html += '<button class="lookup-btn" id="lookup-btn"' + (state.lookingUp ? ' disabled' : '') + '>' + (state.lookingUp ? '查询中' : '查询') + '</button>';
      html += '</div></div>';
    }

    html += '<div class="form-body" id="form-fields">';
    html += renderFieldRenderer(state.fields, state.values, state.errors, { readonly: false });
    html += '</div>';

    html += '<div class="submit-area"><button class="btn-primary submit-btn" id="submit-btn"' + (state.submitting ? ' disabled' : '') + '>' + (state.submitting ? '提交中...' : '提交') + '</button></div>';

    root.innerHTML = html;

    // Bind field events
    var formBody = document.getElementById('form-fields');
    if (formBody) {
      bindFieldEvents(formBody, state.fields, state.values, function(key, val) {
        state.values[key] = val;
        state.errors[key] = '';
      });
    }

    var lookupInput = document.getElementById('lookup-key');
    if (lookupInput) lookupInput.addEventListener('input', function() { state.lookupKey = lookupInput.value; });

    var lookupBtn = document.getElementById('lookup-btn');
    if (lookupBtn) lookupBtn.addEventListener('click', async function() {
      var key = state.lookupKey.trim();
      if (!key) { wx.showToast({ title: '请输入查询键', icon: 'none' }); return; }
      state.lookingUp = true; render();
      try {
        var res = await services.api.get('/forms/' + formId + '/base-data/lookup', { row_key: key });
        if (res.data && typeof res.data === 'object') {
          Object.keys(res.data).forEach(function(k) { state.values[k] = res.data[k]; });
          wx.showToast({ title: '预填充成功', icon: 'success' });
        }
      } catch(e) {
        if (e.status === 404) wx.showToast({ title: '未找到匹配数据', icon: 'none' });
        else wx.showToast({ title: e.message || '查询失败', icon: 'none' });
      }
      state.lookingUp = false; render();
    });

    var submitBtn = document.getElementById('submit-btn');
    if (submitBtn) submitBtn.addEventListener('click', async function() {
      var result = validator.validateForm(state.fields, state.values);
      if (!result.valid) {
        state.errors = result.errors;
        render();
        wx.showToast({ title: result.errors[result.firstErrorKey], icon: 'none' });
        return;
      }
      state.submitting = true; state.errors = {}; render();
      wx.showLoading({ title: '提交中...' });
      try {
        await services.api.post('/forms/' + formId + '/submissions', state.values);
        wx.hideLoading();
        state.submitting = false;
        state.submitSuccess = true;
      } catch(e) {
        wx.hideLoading();
        state.submitting = false;
        wx.showToast({ title: e.message || '提交失败', icon: 'none' });
      }
      render();
    });
  }

  init();
};

// ==================== Submissions ====================
pages['submissions'] = function(root, params) {
  var formId = params.formId;
  var state = {
    loading: true, viewMode: 'list',
    submissions: [], anomalyCount: 0,
    overviewColumns: [], overviewData: [],
  };

  async function loadList() {
    state.loading = true; render();
    try {
      var res = await services.api.get('/forms/' + formId + '/submissions');
      state.submissions = (res.submissions || []).map(function(s) {
        return Object.assign({}, s, { submittedAtText: formatTime(s.submitted_at) });
      });
      state.anomalyCount = state.submissions.filter(function(s) { return s.status === 2; }).length;
    } catch(e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
    state.loading = false; render();
  }

  async function loadOverview() {
    state.loading = true; render();
    try {
      var res = await services.api.get('/forms/' + formId + '/submissions/overview');
      var fields = schemaUtils.schemaToFields(res.schema);
      state.overviewColumns = fields.map(function(f) { return { key: f.key, label: f.label }; });
      state.overviewData = (res.submissions || []).map(function(s) {
        var dv = {};
        if (s.values) Object.keys(s.values).forEach(function(k) {
          var v = s.values[k]; dv[k] = Array.isArray(v) ? v.join('、') : v;
        });
        return { id: s.id, status: s.status, values: dv, anomaly_reasons: s.anomaly_reasons || [] };
      });
      state.anomalyCount = state.overviewData.filter(function(s) { return s.status === 2; }).length;
    } catch(e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
    state.loading = false; render();
  }

  function render() {
    if (state.loading) { root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载中...</span></div>'; return; }

    var html = '';
    // Tabs
    html += '<div class="tab-bar-inner">';
    html += '<div class="tab-inner-item' + (state.viewMode === 'list' ? ' tab-inner-active' : '') + '" data-mode="list">列表视图</div>';
    html += '<div class="tab-inner-item' + (state.viewMode === 'overview' ? ' tab-inner-active' : '') + '" data-mode="overview">总览视图</div>';
    html += '</div>';

    var items = state.viewMode === 'list' ? state.submissions : state.overviewData;
    if (items.length > 0) {
      html += '<div class="stats-bar"><span class="text-secondary text-sm">共 ' + items.length + ' 条提交</span>';
      if (state.anomalyCount > 0) html += '<span class="anomaly-count text-sm">' + state.anomalyCount + ' 条异常</span>';
      html += '</div>';
    }

    if (items.length === 0) {
      html += renderEmptyState('📝', '暂无提交记录', '分享表单后等待用户填写');
    } else if (state.viewMode === 'list') {
      state.submissions.forEach(function(s) {
        html += '<div class="submission-card" data-id="' + s.id + '">';
        html += '<div class="submission-row"><span class="submission-id">#' + s.id + '</span>' + renderStatusBadge(s.status, 'submission') + '</div>';
        html += '<span class="text-secondary text-sm">' + escHtml(s.submittedAtText) + '</span>';
        html += '</div>';
      });
    } else {
      // Overview table
      html += '<div class="overview-wrap"><div class="table-container">';
      html += '<div class="table-fixed"><div class="table-th">#</div>';
      state.overviewData.forEach(function(row) {
        html += '<div class="table-td' + (row.status === 2 ? ' anomaly-row' : '') + '">' + row.id + '</div>';
      });
      html += '</div>';
      html += '<div class="table-scroll"><div class="table-scroll-inner">';
      html += '<div class="table-header-row">';
      state.overviewColumns.forEach(function(col) { html += '<div class="table-th">' + escHtml(col.label) + '</div>'; });
      html += '<div class="table-th">状态</div></div>';
      state.overviewData.forEach(function(row, idx) {
        html += '<div class="table-data-row' + (row.status === 2 ? ' anomaly-row' : '') + '" data-index="' + idx + '">';
        state.overviewColumns.forEach(function(col) {
          html += '<div class="table-td">' + escHtml(row.values[col.key] != null ? row.values[col.key] : '-') + '</div>';
        });
        html += '<div class="table-td">' + renderStatusBadge(row.status, 'submission') + '</div>';
        html += '</div>';
      });
      html += '</div></div></div></div>';
    }

    root.innerHTML = html;

    // Bind tabs
    root.querySelectorAll('.tab-inner-item').forEach(function(el) {
      el.addEventListener('click', function() {
        var mode = el.dataset.mode;
        if (mode === state.viewMode) return;
        state.viewMode = mode;
        if (mode === 'overview' && state.overviewData.length === 0) loadOverview();
        else render();
      });
    });

    // Bind list cards
    root.querySelectorAll('.submission-card').forEach(function(el) {
      el.addEventListener('click', function() {
        wx.navigateTo({ url: '/pages/submission-detail/submission-detail?formId=' + formId + '&submissionId=' + el.dataset.id });
      });
    });

    // Bind overview rows
    root.querySelectorAll('.table-data-row').forEach(function(el) {
      el.addEventListener('click', function() {
        var idx = parseInt(el.dataset.index);
        var row = state.overviewData[idx];
        if (row.status === 2 && row.anomaly_reasons.length > 0) {
          wx.showModal({ title: '异常原因', content: row.anomaly_reasons.join('\n'), showCancel: false });
        } else {
          wx.navigateTo({ url: '/pages/submission-detail/submission-detail?formId=' + formId + '&submissionId=' + row.id });
        }
      });
    });
  }

  loadList();
};

// ==================== Submission Detail ====================
pages['submission-detail'] = function(root, params) {
  var formId = params.formId;
  var submissionId = params.submissionId;

  async function load() {
    try {
      var [form, submission] = await Promise.all([
        services.api.get('/forms/' + formId),
        services.api.get('/forms/' + formId + '/submissions/' + submissionId),
      ]);
      var fields = schemaUtils.schemaToFields(form.schema);

      var html = '<div class="status-card"><div class="status-row"><span class="submission-label">提交 #' + escHtml(submission.id) + '</span>' + renderStatusBadge(submission.status, 'submission') + '</div>';
      html += '<span class="text-secondary text-sm">' + formatTime(submission.submitted_at) + '</span></div>';

      if (submission.status === 2) {
        html += '<div class="anomaly-card"><span class="anomaly-title">AI 检测到异常</span><span class="anomaly-hint text-sm">该提交的数据可能存在问题，请人工核实</span></div>';
      }

      html += '<div class="values-card">' + renderFieldRenderer(fields, submission.values || {}, {}, { readonly: true }) + '</div>';
      root.innerHTML = html;
    } catch(e) {
      root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载失败</span></div>';
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
  }

  root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载中...</span></div>';
  load();
};

// ==================== Base Data ====================
pages['base-data'] = function(root, params) {
  var formId = params.formId;
  var state = { loading: true, rows: [], showImport: false, importText: '', importing: false };

  async function loadData() {
    state.loading = true; render();
    try {
      var res = await services.api.get('/forms/' + formId + '/base-data');
      state.rows = (res.rows || []).map(function(r) {
        var preview = '';
        if (r.data && typeof r.data === 'object') {
          var vals = Object.values(r.data);
          preview = vals.slice(0, 3).join(', ');
          if (vals.length > 3) preview += '...';
        }
        return { id: r.id, row_key: r.row_key, data: r.data, preview: preview };
      });
    } catch(e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
    state.loading = false; render();
  }

  function render() {
    if (state.loading) { root.innerHTML = '<div class="loading-center"><span class="text-secondary">加载中...</span></div>'; return; }

    var html = '<div class="header-card">';
    html += '<div class="header-title">底表数据管理</div>';
    html += '<div class="text-secondary text-sm mt-sm">共 ' + state.rows.length + ' 条记录</div>';
    html += '<div class="header-actions mt-md">';
    html += '<button class="action-btn" id="show-import">导入数据</button>';
    if (state.rows.length > 0) html += '<button class="action-btn action-danger" id="clear-all">清空数据</button>';
    html += '</div></div>';

    if (state.showImport) {
      html += '<div class="import-panel">';
      html += '<div class="import-hint"><span class="text-sm">请粘贴 JSON 数组格式的数据，每项包含 row_key 和 data 字段：</span>';
      html += '<span class="text-sm text-secondary mt-sm">[{ "row_key":"EMP001", "data":{ "f_001":"张三" } }, ...]</span></div>';
      html += '<textarea class="import-textarea" id="import-text" placeholder=\'[{"row_key":"EMP001","data":{"f_001":"张三","f_002":"技术部"}}]\'>' + escHtml(state.importText) + '</textarea>';
      html += '<div class="import-actions">';
      html += '<button class="action-btn" id="cancel-import">取消</button>';
      html += '<button class="action-btn action-primary" id="do-import"' + (state.importing ? ' disabled' : '') + '>' + (state.importing ? '导入中...' : '确认导入') + '</button>';
      html += '</div></div>';
    }

    if (state.rows.length === 0 && !state.showImport) {
      html += renderEmptyState('📊', '暂无底表数据', '点击「导入数据」添加预填充数据');
    } else if (state.rows.length > 0) {
      html += '<div class="data-list">';
      state.rows.forEach(function(r) {
        html += '<div class="data-card"><div class="data-key">' + escHtml(r.row_key) + '</div><div class="data-preview text-secondary text-sm">' + escHtml(r.preview) + '</div></div>';
      });
      html += '</div>';
    }

    root.innerHTML = html;

    // Bind
    document.getElementById('show-import').addEventListener('click', function() { state.showImport = true; state.importText = ''; render(); });

    var cancelBtn = document.getElementById('cancel-import');
    if (cancelBtn) cancelBtn.addEventListener('click', function() { state.showImport = false; render(); });

    var importTextEl = document.getElementById('import-text');
    if (importTextEl) importTextEl.addEventListener('input', function() { state.importText = importTextEl.value; });

    var doImportBtn = document.getElementById('do-import');
    if (doImportBtn) doImportBtn.addEventListener('click', async function() {
      var text = state.importText.trim();
      if (!text) { wx.showToast({ title: '请输入数据', icon: 'none' }); return; }
      var parsed;
      try { parsed = JSON.parse(text); } catch(e) { wx.showToast({ title: 'JSON 格式错误', icon: 'none' }); return; }
      if (!Array.isArray(parsed) || parsed.length === 0) { wx.showToast({ title: '数据应为非空数组', icon: 'none' }); return; }
      for (var i = 0; i < parsed.length; i++) {
        if (!parsed[i].row_key || !parsed[i].data) { wx.showToast({ title: '第' + (i+1) + '项缺少 row_key 或 data', icon: 'none' }); return; }
      }
      state.importing = true; render();
      try {
        var res = await services.api.post('/forms/' + formId + '/base-data', { rows: parsed });
        wx.showToast({ title: '成功导入 ' + res.imported + ' 条', icon: 'success' });
        state.showImport = false; state.importText = ''; state.importing = false;
        loadData();
      } catch(e) {
        state.importing = false;
        wx.showToast({ title: e.message || '导入失败', icon: 'none' });
        render();
      }
    });

    var clearBtn = document.getElementById('clear-all');
    if (clearBtn) clearBtn.addEventListener('click', function() {
      wx.showModal({ title: '确认清空', content: '确定清空所有底表数据吗？此操作不可恢复。', success: async function(res) {
        if (res.confirm) {
          wx.showLoading({ title: '清空中...' });
          try { await services.api.del('/forms/' + formId + '/base-data'); wx.showToast({ title: '已清空', icon: 'success' }); loadData(); }
          catch(e) { wx.showToast({ title: e.message || '清空失败', icon: 'none' }); }
          wx.hideLoading();
        }
      }});
    });
  }

  loadData();
};

// ==================== AI Generate ====================
pages['ai-generate'] = function(root, params) {
  var STORAGE_KEY = 'aiGenerateJobId';
  var state = {
    description: '',
    generating: false,
    statusText: '正在排队...',
    jobId: null,
    result: null,
    previewFields: [],
  };
  var pollTimer = null;
  var pollCount = 0;

  function clearPoll() { if (pollTimer) { clearTimeout(pollTimer); pollTimer = null; } }

  function failGeneration(msg) {
    wx.removeStorageSync(STORAGE_KEY);
    state.generating = false;
    state.jobId = null;
    wx.showToast({ title: msg, icon: 'none', duration: 3000 });
    render();
  }

  function handleComplete(output) {
    wx.removeStorageSync(STORAGE_KEY);
    state.jobId = null;
    state.generating = false;

    var parsed;
    try { parsed = JSON.parse(output || '{}'); } catch (e) { parsed = null; }
    if (!parsed || !parsed.schema) {
      wx.showToast({ title: 'AI 返回数据解析失败', icon: 'none', duration: 3000 });
      render();
      return;
    }

    var fields = schemaUtils.schemaToFields(parsed.schema);
    state.previewFields = fields.map(function(f) {
      return { key: f.key, label: f.label, type: f.type, typeLabel: TYPE_LABEL_MAP[f.type] || f.type, required: f.required };
    });
    state.result = parsed;
    render();
  }

  function isAlive() { return root.isConnected; }

  function schedulePoll() {
    clearPoll();
    pollTimer = setTimeout(function() { if (isAlive()) checkJob(); }, 2000);
  }

  async function checkJob() {
    pollCount++;
    if (pollCount > 60) { failGeneration('生成超时，请稍后重试'); return; }
    try {
      var res = await services.api.get('/jobs/' + state.jobId);
      if (!isAlive()) return;
      if (res.status === 0) {
        state.statusText = '正在排队...'; render(); schedulePoll();
      } else if (res.status === 1) {
        state.statusText = 'AI 正在生成表单...'; render(); schedulePoll();
      } else if (res.status === 2) {
        handleComplete(res.output);
      } else {
        failGeneration(res.output || 'AI 生成失败');
      }
    } catch (e) {
      if (!isAlive()) return;
      failGeneration(e.message || '查询任务状态失败');
    }
  }

  function resumeIfAny() {
    var saved = wx.getStorageSync(STORAGE_KEY);
    if (!saved) return false;
    state.jobId = saved;
    state.generating = true;
    state.statusText = '恢复生成进度...';
    pollCount = 0;
    render();
    schedulePoll();
    return true;
  }

  function render() {
    var html = '';

    if (!state.result && !state.generating) {
      html += '<div class="input-card">';
      html += '<span class="card-title">描述你要创建的表单</span>';
      html += '<span class="card-hint text-secondary text-sm mt-sm">例如：员工信息登记表，包含姓名、年龄、部门、手机号、月薪</span>';
      html += '<textarea class="desc-textarea" id="ai-desc" placeholder="请描述表单内容..." maxlength="500">' + escHtml(state.description) + '</textarea>';
      html += '<button class="btn-primary generate-btn" id="gen-btn">AI 生成</button>';
      html += '</div>';
    }

    if (state.generating) {
      html += '<div class="generating-state"><span class="generating-icon">🤖</span>';
      html += '<span class="generating-title">' + escHtml(state.statusText) + '</span>';
      html += '<span class="generating-hint text-secondary text-sm mt-sm">你可以暂时离开，回到此页面会自动继续。</span></div>';
    }

    if (state.result && !state.generating) {
      html += '<div class="result-card">';
      html += '<span class="card-title">生成结果</span>';
      html += '<div class="preview-section mt-md"><span class="preview-label">标题</span><span class="preview-value">' + escHtml(state.result.title) + '</span></div>';
      if (state.result.description) {
        html += '<div class="preview-section"><span class="preview-label">描述</span><span class="preview-value">' + escHtml(state.result.description) + '</span></div>';
      }
      html += '<div class="preview-section"><span class="preview-label">字段 (' + state.previewFields.length + ')</span>';
      html += '<div class="field-preview-list">';
      state.previewFields.forEach(function(f) {
        html += '<div class="field-preview-item">';
        html += '<span class="field-preview-label">' + escHtml(f.label) + '</span>';
        html += '<span class="field-preview-type text-secondary text-sm">' + escHtml(f.typeLabel) + '</span>';
        if (f.required) html += '<span class="field-preview-required">必填</span>';
        html += '</div>';
      });
      html += '</div></div>';
      html += '<div class="result-actions mt-lg">';
      html += '<button class="action-btn-outline" id="regen-btn">重新生成</button>';
      html += '<button class="btn-primary" id="use-btn">使用此表单</button>';
      html += '</div></div>';
    }

    root.innerHTML = html;

    var descEl = document.getElementById('ai-desc');
    if (descEl) descEl.addEventListener('input', function() { state.description = descEl.value; });

    var genBtn = document.getElementById('gen-btn');
    if (genBtn) genBtn.addEventListener('click', async function() {
      var desc = state.description.trim();
      if (!desc) { wx.showToast({ title: '请输入表单描述', icon: 'none' }); return; }
      state.generating = true;
      state.statusText = '正在排队...';
      state.result = null;
      state.previewFields = [];
      render();
      try {
        var res = await services.api.post('/forms/generate', { description: desc });
        state.jobId = res.job_id;
        wx.setStorageSync(STORAGE_KEY, state.jobId);
        pollCount = 0;
        schedulePoll();
      } catch (e) {
        var msg = e.message || 'AI 生成失败';
        if (e.status === 503) msg = 'AI 服务暂未开启，请手动创建表单';
        failGeneration(msg);
      }
    });

    var regenBtn = document.getElementById('regen-btn');
    if (regenBtn) regenBtn.addEventListener('click', function() { state.result = null; state.previewFields = []; render(); });

    var useBtn = document.getElementById('use-btn');
    if (useBtn) useBtn.addEventListener('click', function() {
      var app = getApp();
      app.globalData.tempFormDraft = {
        title: state.result.title,
        description: state.result.description,
        schema: state.result.schema,
      };
      wx.redirectTo({ url: '/pages/form-editor/form-editor' });
    });
  }

  if (!resumeIfAny()) render();
};

// ==================== Report ====================
pages['report'] = function(root, params) {
  var formId = params.formId;
  var MODE = { LOADING: 'loading', EMPTY: 'empty', GENERATING: 'generating', COMPLETED: 'completed', FAILED: 'failed' };
  var state = {
    mode: MODE.LOADING,
    jobId: null,
    reportOutput: '',
    reportHTML: '',
    finishedAt: '',
    statusText: '正在排队...',
    statusHint: '请稍候，AI 正在准备分析数据',
    failedMsg: '',
  };
  var pollTimer = null;
  var pollCount = 0;

  function clearPoll() { if (pollTimer) { clearTimeout(pollTimer); pollTimer = null; } }
  function isAlive() { return root.isConnected; }

  function formatFinishedAt(s) {
    if (!s) return '';
    var d = new Date(s);
    if (isNaN(d.getTime())) return '';
    var pad = function(n) { return n < 10 ? '0' + n : String(n); };
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
      ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes());
  }

  function setReport(output, jobId, finishedAt) {
    state.mode = MODE.COMPLETED;
    state.jobId = jobId;
    state.reportOutput = output || '';
    state.reportHTML = markdown.render(output || '');
    state.finishedAt = formatFinishedAt(finishedAt);
    render();
  }

  async function loadLatest() {
    state.mode = MODE.LOADING;
    render();
    try {
      var res = await services.api.get('/forms/' + formId + '/report/latest');
      if (!isAlive()) return;
      if (!res || !res.output) {
        state.mode = MODE.EMPTY;
        render();
        return;
      }
      setReport(res.output, res.job_id, res.finished_at);
    } catch (e) {
      if (!isAlive()) return;
      // 204 surfaces as empty data via fetch; a hard error means something else is wrong.
      if (e.status === 404 || e.status === 204) {
        state.mode = MODE.EMPTY;
      } else {
        state.mode = MODE.FAILED;
        state.failedMsg = e.message || '加载报告失败';
      }
      render();
    }
  }

  async function startReport() {
    state.mode = MODE.GENERATING;
    state.statusText = '正在排队...';
    state.statusHint = '请稍候，AI 正在准备分析数据';
    state.failedMsg = '';
    render();

    try {
      var res = await services.api.post('/forms/' + formId + '/report');
      if (!isAlive()) return;
      state.jobId = res.job_id;
      pollCount = 0;
      schedulePoll();
    } catch (e) {
      if (!isAlive()) return;
      state.mode = MODE.FAILED;
      state.failedMsg = e.message || '无法创建报告任务';
      render();
    }
  }

  function schedulePoll() {
    clearPoll();
    pollTimer = setTimeout(function() { if (isAlive()) checkJob(); }, 2000);
  }

  async function checkJob() {
    pollCount++;
    if (pollCount > 60) {
      state.mode = MODE.FAILED;
      state.failedMsg = '生成超时，请稍后重试';
      render();
      return;
    }
    try {
      var res = await services.api.get('/jobs/' + state.jobId);
      if (!isAlive()) return;
      if (res.status === 0) {
        state.statusText = '排队中...';
        state.statusHint = '请稍候，AI 正在准备分析数据';
        render(); schedulePoll();
      } else if (res.status === 1) {
        state.statusText = 'AI 正在分析数据...';
        state.statusHint = '即将完成，请耐心等待';
        render(); schedulePoll();
      } else if (res.status === 2) {
        setReport(res.output, res.job_id, res.finished_at);
      } else {
        state.mode = MODE.FAILED;
        state.failedMsg = res.output || '报告生成失败';
        render();
      }
    } catch (e) {
      if (!isAlive()) return;
      state.mode = MODE.FAILED;
      state.failedMsg = e.message || '查询任务状态失败';
      render();
    }
  }

  function downloadPDF() {
    if (!state.jobId) { wx.showToast({ title: '暂无可下载的报告', icon: 'none' }); return; }
    var token = wx.getStorageSync('auth_token') || '';
    wx.showLoading({ title: '正在下载...' });
    wx.downloadFile({
      url: services.api.BASE_URL + '/jobs/' + state.jobId + '/pdf',
      header: { 'Authorization': 'Bearer ' + token },
      success: function(res) {
        wx.hideLoading();
        if (res.statusCode === 200 && res.tempFilePath) {
          wx.openDocument({
            filePath: res.tempFilePath,
            fileType: 'pdf',
            showMenu: true,
            fail: function() { wx.showToast({ title: '打开 PDF 失败', icon: 'none' }); },
          });
        } else {
          wx.showToast({ title: '下载失败 (HTTP ' + res.statusCode + ')', icon: 'none' });
        }
      },
      fail: function() {
        wx.hideLoading();
        wx.showToast({ title: '下载失败', icon: 'none' });
      },
    });
  }

  function render() {
    var html = '';

    if (state.mode === MODE.LOADING) {
      html += '<div class="generating-state"><span class="generating-icon">⏳</span>';
      html += '<span class="generating-title">加载报告中...</span></div>';
    }

    if (state.mode === MODE.EMPTY) {
      html += '<div class="empty-state"><div class="empty-icon">📝</div>';
      html += '<div class="empty-text">该表单尚未生成过报告</div>';
      html += '<button class="btn-primary mt-lg" id="gen-btn" style="margin-top:16px;width:auto;padding:10px 40px;">生成报告</button></div>';
    }

    if (state.mode === MODE.GENERATING) {
      html += '<div class="generating-state"><span class="generating-icon">🤖</span>';
      html += '<span class="generating-title">' + escHtml(state.statusText) + '</span>';
      html += '<span class="generating-hint text-secondary text-sm mt-sm">' + escHtml(state.statusHint) + '</span>';
      html += '<span class="generating-hint text-secondary text-sm mt-sm">你可以暂时离开，回到此页面会重新加载最新结果。</span></div>';
    }

    if (state.mode === MODE.FAILED) {
      html += '<div class="failed-state"><span class="failed-icon">⚠️</span>';
      html += '<span class="failed-title">操作失败</span>';
      html += '<span class="failed-msg text-secondary text-sm mt-sm">' + escHtml(state.failedMsg) + '</span>';
      html += '<button class="btn-primary mt-lg" id="retry-btn" style="width:auto;padding:10px 40px;">重试</button></div>';
    }

    if (state.mode === MODE.COMPLETED) {
      html += '<div class="report-card">';
      html += '<span class="report-title">数据分析报告</span>';
      if (state.finishedAt) {
        html += '<span class="text-secondary text-sm" style="display:block;margin-bottom:12px;">生成于 ' + escHtml(state.finishedAt) + '</span>';
      }
      html += '<div class="report-content md-body">' + state.reportHTML + '</div>';
      html += '<div class="result-actions mt-lg">';
      html += '<button class="action-btn-outline" id="pdf-btn">下载 PDF</button>';
      html += '<button class="btn-primary" id="regen-btn">重新生成</button>';
      html += '</div></div>';
    }

    root.innerHTML = html;

    var genBtn = document.getElementById('gen-btn');
    if (genBtn) genBtn.addEventListener('click', function() { startReport(); });

    var retryBtn = document.getElementById('retry-btn');
    if (retryBtn) retryBtn.addEventListener('click', function() {
      if (state.jobId) startReport(); else loadLatest();
    });

    var regenBtn = document.getElementById('regen-btn');
    if (regenBtn) regenBtn.addEventListener('click', function() { startReport(); });

    var pdfBtn = document.getElementById('pdf-btn');
    if (pdfBtn) pdfBtn.addEventListener('click', function() { downloadPDF(); });
  }

  loadLatest();
};
