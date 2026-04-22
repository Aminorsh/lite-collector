/**
 * wx.* API shim for browser environment.
 * Replaces WeChat Mini Program APIs with browser equivalents.
 */

var _toastTimer = null;

var wx = {
  // --- Storage ---
  getStorageSync(key) {
    try {
      var val = localStorage.getItem(key);
      if (val === null) return '';
      try { return JSON.parse(val); } catch(e) { return val; }
    } catch(e) { return ''; }
  },

  setStorageSync(key, val) {
    localStorage.setItem(key, typeof val === 'object' ? JSON.stringify(val) : val);
  },

  removeStorageSync(key) {
    localStorage.removeItem(key);
  },

  // --- Login (mock) ---
  login(opts) {
    // Simulate wx.login by generating a mock code
    setTimeout(function() {
      if (opts.success) opts.success({ code: 'mock_code_' + Date.now() });
    }, 50);
  },

  // --- HTTP ---
  request(opts) {
    var init = {
      method: opts.method || 'GET',
      headers: opts.header || {},
    };

    var url = opts.url;

    if (opts.data && (init.method === 'GET' || init.method === 'HEAD')) {
      var params = new URLSearchParams(opts.data).toString();
      if (params) url += '?' + params;
    } else if (opts.data) {
      init.body = JSON.stringify(opts.data);
    }

    fetch(url, init)
      .then(function(resp) {
        return resp.text().then(function(text) {
          var data;
          try { data = JSON.parse(text); } catch(e) { data = text; }
          if (opts.success) opts.success({ statusCode: resp.status, data: data });
        });
      })
      .catch(function(err) {
        if (opts.fail) opts.fail({ errMsg: err.message });
      });
  },

  // --- UI: Toast ---
  showToast(opts) {
    var el = document.getElementById('toast');
    var iconEl = document.getElementById('toast-icon');
    var textEl = document.getElementById('toast-text');

    if (opts.icon === 'success') {
      iconEl.textContent = '\u2713';
      iconEl.style.display = 'block';
    } else {
      iconEl.style.display = 'none';
    }
    textEl.textContent = opts.title || '';
    el.classList.remove('toast-hidden');

    clearTimeout(_toastTimer);
    _toastTimer = setTimeout(function() {
      el.classList.add('toast-hidden');
    }, opts.duration || 1500);
  },

  // --- UI: Loading ---
  showLoading(opts) {
    var el = document.getElementById('loading-overlay');
    document.getElementById('loading-text').textContent = (opts && opts.title) || '加载中...';
    el.classList.remove('loading-hidden');
  },

  hideLoading() {
    document.getElementById('loading-overlay').classList.add('loading-hidden');
  },

  // --- UI: Modal ---
  showModal(opts) {
    var overlay = document.getElementById('modal-overlay');
    document.getElementById('modal-title').textContent = opts.title || '';
    document.getElementById('modal-content').textContent = opts.content || '';

    var actions = document.getElementById('modal-actions');
    actions.innerHTML = '';

    if (opts.showCancel !== false) {
      var cancelBtn = document.createElement('button');
      cancelBtn.textContent = opts.cancelText || '取消';
      cancelBtn.onclick = function() {
        overlay.classList.add('modal-hidden');
        if (opts.success) opts.success({ confirm: false, cancel: true });
      };
      actions.appendChild(cancelBtn);
    }

    var confirmBtn = document.createElement('button');
    confirmBtn.textContent = opts.confirmText || '确定';
    confirmBtn.onclick = function() {
      overlay.classList.add('modal-hidden');
      if (opts.success) opts.success({ confirm: true, cancel: false });
    };
    actions.appendChild(confirmBtn);

    overlay.classList.remove('modal-hidden');
  },

  // --- UI: Action Sheet ---
  showActionSheet(opts) {
    var overlay = document.getElementById('actionsheet-overlay');
    var box = document.getElementById('actionsheet-box');
    box.innerHTML = '';

    (opts.itemList || []).forEach(function(item, idx) {
      var div = document.createElement('div');
      div.className = 'actionsheet-item';
      div.textContent = item;
      div.onclick = function() {
        overlay.classList.add('modal-hidden');
        if (opts.success) opts.success({ tapIndex: idx });
      };
      box.appendChild(div);
    });

    var cancel = document.createElement('div');
    cancel.className = 'actionsheet-cancel';
    cancel.textContent = '取消';
    cancel.onclick = function() {
      overlay.classList.add('modal-hidden');
      if (opts.fail) opts.fail({ errMsg: 'showActionSheet:fail cancel' });
    };
    box.appendChild(cancel);

    overlay.classList.remove('modal-hidden');
  },

  // --- Navigation (delegated to router) ---
  navigateTo(opts) {
    if (typeof router !== 'undefined') router.navigateTo(opts.url);
  },

  navigateBack(opts) {
    if (typeof router !== 'undefined') {
      router.back();
    } else if (opts && opts.fail) {
      opts.fail();
    }
  },

  redirectTo(opts) {
    if (typeof router !== 'undefined') router.navigateTo(opts.url, true);
  },

  switchTab(opts) {
    if (typeof router !== 'undefined') {
      var page = opts.url.replace(/^\/pages\//, '').replace(/\/.*$/, '');
      router.switchTab(page);
    }
  },

  stopPullDownRefresh() {
    // no-op in browser
  },

  // --- File download / open (browser: fetch blob, stash url, open via openDocument) ---
  downloadFile(opts) {
    var headers = opts.header || {};
    fetch(opts.url, { headers: headers })
      .then(function(resp) {
        if (!resp.ok) {
          if (opts.fail) opts.fail({ errMsg: 'downloadFile:fail status ' + resp.status });
          return null;
        }
        return resp.blob().then(function(blob) {
          var tempFilePath = URL.createObjectURL(blob);
          if (opts.success) opts.success({ statusCode: resp.status, tempFilePath: tempFilePath });
        });
      })
      .catch(function(err) {
        if (opts.fail) opts.fail({ errMsg: err.message || 'downloadFile:fail' });
      });
  },

  openDocument(opts) {
    // In the browser, a blob URL from downloadFile opens in a new tab;
    // most browsers render PDFs inline and expose a download button.
    try {
      window.open(opts.filePath, '_blank');
      if (opts.success) opts.success({});
    } catch (err) {
      if (opts.fail) opts.fail({ errMsg: err.message || 'openDocument:fail' });
    }
  },

  // --- Media (file input shim) ---
  chooseMedia(opts) {
    var input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';
    input.onchange = function() {
      if (input.files && input.files[0]) {
        var url = URL.createObjectURL(input.files[0]);
        if (opts.success) opts.success({ tempFiles: [{ tempFilePath: url }] });
      }
    };
    input.click();
  },
};

// Global getApp
var _appData = {
  token: '',
  userInfo: null,
  loginReady: null,
  tempFormDraft: null,
};

function getApp() {
  return { globalData: _appData };
}
