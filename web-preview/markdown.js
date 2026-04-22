/**
 * Browser markdown → HTML renderer.
 * Mirrors miniprogram/utils/markdown.js but emits an HTML string so we can
 * drop it into innerHTML instead of a <rich-text nodes> tree.
 *
 * Supports: h1-h4, paragraphs, bold (**), italic (*), inline code (`),
 * fenced code blocks, blockquotes, unordered/ordered lists, horizontal rules.
 */
var markdown = (function() {

  function esc(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  function renderInline(line) {
    var out = '';
    var i = 0;
    var buf = '';
    var flush = function() { if (buf) { out += esc(buf); buf = ''; } };

    while (i < line.length) {
      var ch = line[i];

      if (ch === '`') {
        var end = line.indexOf('`', i + 1);
        if (end > i) {
          flush();
          out += '<code class="md-code-inline">' + esc(line.substring(i + 1, end)) + '</code>';
          i = end + 1;
          continue;
        }
      }

      if (ch === '*' && line[i + 1] === '*') {
        var bEnd = line.indexOf('**', i + 2);
        if (bEnd > i + 1) {
          flush();
          out += '<strong class="md-bold">' + renderInline(line.substring(i + 2, bEnd)) + '</strong>';
          i = bEnd + 2;
          continue;
        }
      }

      if (ch === '*') {
        var iEnd = line.indexOf('*', i + 1);
        if (iEnd > i) {
          flush();
          out += '<em class="md-italic">' + renderInline(line.substring(i + 1, iEnd)) + '</em>';
          i = iEnd + 1;
          continue;
        }
      }

      buf += ch;
      i++;
    }
    flush();
    return out;
  }

  function render(md) {
    if (!md) return '';
    var lines = String(md).replace(/\r\n/g, '\n').split('\n');
    var out = '';

    var inCode = false;
    var codeBuf = [];

    var listItems = [];
    var listOrdered = false;
    var flushList = function() {
      if (listItems.length) {
        var tag = listOrdered ? 'ol' : 'ul';
        var cls = listOrdered ? 'md-ol' : 'md-ul';
        out += '<' + tag + ' class="' + cls + '">';
        listItems.forEach(function(it) {
          out += '<li class="md-li">' + renderInline(it) + '</li>';
        });
        out += '</' + tag + '>';
        listItems = [];
      }
    };

    var paraLines = [];
    var flushPara = function() {
      if (paraLines.length) {
        out += '<p class="md-p">' + renderInline(paraLines.join(' ')) + '</p>';
        paraLines = [];
      }
    };

    for (var i = 0; i < lines.length; i++) {
      var raw = lines[i];
      var line = raw.replace(/\s+$/, '');

      if (/^\s*```/.test(line)) {
        if (inCode) {
          flushPara(); flushList();
          out += '<pre class="md-pre">' + esc(codeBuf.join('\n')) + '</pre>';
          codeBuf = [];
          inCode = false;
        } else {
          flushPara(); flushList();
          inCode = true;
        }
        continue;
      }
      if (inCode) { codeBuf.push(raw); continue; }

      if (line === '') { flushPara(); flushList(); continue; }

      if (/^---+$/.test(line) || /^\*\*\*+$/.test(line)) {
        flushPara(); flushList();
        out += '<hr class="md-hr" />';
        continue;
      }

      var h = line.match(/^(#{1,4})\s+(.*)$/);
      if (h) {
        flushPara(); flushList();
        var level = h[1].length;
        out += '<h' + level + ' class="md-h' + level + '">' + renderInline(h[2]) + '</h' + level + '>';
        continue;
      }

      if (/^\s*>/.test(line)) {
        flushPara(); flushList();
        var q = line.replace(/^\s*>\s?/, '');
        out += '<blockquote class="md-blockquote">' + renderInline(q) + '</blockquote>';
        continue;
      }

      var ul = line.match(/^\s*[-*+]\s+(.*)$/);
      if (ul) {
        flushPara();
        if (listItems.length && listOrdered) flushList();
        listOrdered = false;
        listItems.push(ul[1]);
        continue;
      }
      var ol = line.match(/^\s*\d+\.\s+(.*)$/);
      if (ol) {
        flushPara();
        if (listItems.length && !listOrdered) flushList();
        listOrdered = true;
        listItems.push(ol[1]);
        continue;
      }

      flushList();
      paraLines.push(line);
    }
    flushPara(); flushList();
    if (inCode && codeBuf.length) {
      out += '<pre class="md-pre">' + esc(codeBuf.join('\n')) + '</pre>';
    }
    return out;
  }

  return { render: render };
})();
