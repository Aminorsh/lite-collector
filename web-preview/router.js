/**
 * Simple client-side router for mini program page navigation.
 */
var router = (function() {
  var _stack = [];      // navigation stack
  var _currentPage = null;
  var _tabPages = ['index', 'profile'];

  function parseUrl(url) {
    // "/pages/form-detail/form-detail?formId=1" -> { page: 'form-detail', params: {formId:'1'} }
    var clean = url.replace(/^\/pages\//, '');
    var parts = clean.split('?');
    var pageName = parts[0].split('/')[0];
    var params = {};
    if (parts[1]) {
      parts[1].split('&').forEach(function(kv) {
        var pair = kv.split('=');
        params[decodeURIComponent(pair[0])] = decodeURIComponent(pair[1] || '');
      });
    }
    return { page: pageName, params: params };
  }

  function updateNavBar(pageName, isTab) {
    var titles = {
      'index': '我的表单',
      'profile': '我的',
      'form-editor': '编辑表单',
      'form-detail': '表单详情',
      'form-fill': '填写表单',
      'submissions': '提交列表',
      'submission-detail': '提交详情',
      'base-data': '底表数据',
      'ai-generate': 'AI 生成表单',
      'report': '数据报告',
    };

    document.getElementById('nav-title').textContent = titles[pageName] || 'Lite Collector';
    document.getElementById('nav-back').style.display = isTab ? 'none' : 'block';
  }

  function updateTabBar(pageName) {
    var isTab = _tabPages.indexOf(pageName) !== -1;
    document.getElementById('tab-bar').style.display = isTab ? 'flex' : 'none';

    document.querySelectorAll('.tab-item').forEach(function(el) {
      if (el.dataset.tab === pageName) {
        el.classList.add('active');
      } else {
        el.classList.remove('active');
      }
    });
  }

  function renderPage(pageName, params) {
    var container = document.getElementById('page-container');
    container.innerHTML = '<div class="container" id="page-root"></div>';

    var isTab = _tabPages.indexOf(pageName) !== -1;
    updateNavBar(pageName, isTab);
    updateTabBar(pageName);

    // Call the page renderer
    if (pages[pageName]) {
      pages[pageName](document.getElementById('page-root'), params || {});
    } else {
      container.innerHTML = '<div class="container"><div class="empty-state"><div class="empty-icon">🚧</div><div class="empty-text">页面不存在: ' + pageName + '</div></div></div>';
    }

    _currentPage = pageName;
    window.scrollTo(0, 0);
  }

  return {
    navigateTo: function(url, replace) {
      var parsed = parseUrl(url);
      if (!replace) {
        _stack.push(_currentPage + (window._lastParams ? '?' + new URLSearchParams(window._lastParams).toString() : ''));
      }
      window._lastParams = parsed.params;
      renderPage(parsed.page, parsed.params);
    },

    back: function() {
      if (_stack.length > 0) {
        var prev = _stack.pop();
        var parsed = parseUrl('/pages/' + prev);
        window._lastParams = parsed.params;
        renderPage(parsed.page, parsed.params);
      } else {
        this.switchTab('index');
      }
    },

    switchTab: function(tabName) {
      _stack = [];
      window._lastParams = {};
      renderPage(tabName, {});
    },

    getCurrentPage: function() {
      return _currentPage;
    },
  };
})();
