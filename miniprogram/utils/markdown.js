/**
 * Minimal markdown → WeChat <rich-text> nodes parser.
 *
 * Supports headings (h1-h4), paragraphs, unordered/ordered lists, bold (**),
 * italic (*), inline code (`), fenced code blocks, blockquotes, and
 * horizontal rules. Tables are flattened into preformatted blocks.
 *
 * Why roll our own: towxml pulls ~1.5 MB into the mini-program bundle and
 * requires the "构建 npm" DevTools step. For the narrow subset the backend
 * emits (see services/ai_service.go prompt), a small home-grown parser is
 * simpler to audit and maintain.
 */

function escapeText(s) {
  // rich-text renders text literally; nothing to escape at the text node level.
  return s
}

function textNode(text) {
  return { type: 'text', text: escapeText(text) }
}

/**
 * Parse inline markdown (bold, italic, code) into rich-text children.
 * Ordered longest-delimiter-first so `**` beats `*`.
 */
function parseInline(line) {
  var nodes = []
  var i = 0
  var buf = ''
  var flush = function () {
    if (buf) { nodes.push(textNode(buf)); buf = '' }
  }

  while (i < line.length) {
    var ch = line[i]

    // Inline code `...`
    if (ch === '`') {
      var end = line.indexOf('`', i + 1)
      if (end > i) {
        flush()
        nodes.push({
          name: 'code',
          attrs: { class: 'md-code-inline' },
          children: [textNode(line.substring(i + 1, end))],
        })
        i = end + 1
        continue
      }
    }

    // Bold **...**
    if (ch === '*' && line[i + 1] === '*') {
      var bEnd = line.indexOf('**', i + 2)
      if (bEnd > i + 1) {
        flush()
        nodes.push({
          name: 'strong',
          attrs: { class: 'md-bold' },
          children: parseInline(line.substring(i + 2, bEnd)),
        })
        i = bEnd + 2
        continue
      }
    }

    // Italic *...*
    if (ch === '*') {
      var iEnd = line.indexOf('*', i + 1)
      if (iEnd > i) {
        flush()
        nodes.push({
          name: 'em',
          attrs: { class: 'md-italic' },
          children: parseInline(line.substring(i + 1, iEnd)),
        })
        i = iEnd + 1
        continue
      }
    }

    buf += ch
    i++
  }
  flush()
  return nodes
}

function makeList(ordered, items) {
  return {
    name: ordered ? 'ol' : 'ul',
    attrs: { class: ordered ? 'md-ol' : 'md-ul' },
    children: items.map(function (it) {
      return {
        name: 'li',
        attrs: { class: 'md-li' },
        children: parseInline(it),
      }
    }),
  }
}

/**
 * Parse full markdown into rich-text compatible node array.
 * Pass the result to <rich-text nodes="{{nodes}}" />.
 */
function parse(md) {
  if (!md) return []
  var lines = String(md).replace(/\r\n/g, '\n').split('\n')
  var out = []
  var i = 0

  var inCode = false
  var codeBuf = []

  var listItems = []
  var listOrdered = false
  var flushList = function () {
    if (listItems.length) {
      out.push(makeList(listOrdered, listItems))
      listItems = []
    }
  }

  var paraLines = []
  var flushPara = function () {
    if (paraLines.length) {
      out.push({
        name: 'p',
        attrs: { class: 'md-p' },
        children: parseInline(paraLines.join(' ')),
      })
      paraLines = []
    }
  }

  for (i = 0; i < lines.length; i++) {
    var raw = lines[i]
    var line = raw.replace(/\s+$/, '')

    if (/^\s*```/.test(line)) {
      if (inCode) {
        flushPara(); flushList()
        out.push({
          name: 'pre',
          attrs: { class: 'md-pre' },
          children: [textNode(codeBuf.join('\n'))],
        })
        codeBuf = []
        inCode = false
      } else {
        flushPara(); flushList()
        inCode = true
      }
      continue
    }
    if (inCode) { codeBuf.push(raw); continue }

    if (line === '') { flushPara(); flushList(); continue }

    if (/^---+$/.test(line) || /^\*\*\*+$/.test(line)) {
      flushPara(); flushList()
      out.push({ name: 'hr', attrs: { class: 'md-hr' } })
      continue
    }

    var h = line.match(/^(#{1,4})\s+(.*)$/)
    if (h) {
      flushPara(); flushList()
      var level = h[1].length
      out.push({
        name: 'h' + level,
        attrs: { class: 'md-h' + level },
        children: parseInline(h[2]),
      })
      continue
    }

    if (/^\s*>/.test(line)) {
      flushPara(); flushList()
      var q = line.replace(/^\s*>\s?/, '')
      out.push({
        name: 'blockquote',
        attrs: { class: 'md-blockquote' },
        children: parseInline(q),
      })
      continue
    }

    var ul = line.match(/^\s*[-*+]\s+(.*)$/)
    if (ul) {
      flushPara()
      if (listItems.length && listOrdered) flushList()
      listOrdered = false
      listItems.push(ul[1])
      continue
    }
    var ol = line.match(/^\s*\d+\.\s+(.*)$/)
    if (ol) {
      flushPara()
      if (listItems.length && !listOrdered) flushList()
      listOrdered = true
      listItems.push(ol[1])
      continue
    }

    flushList()
    paraLines.push(line)
  }
  flushPara(); flushList()
  if (inCode && codeBuf.length) {
    out.push({
      name: 'pre',
      attrs: { class: 'md-pre' },
      children: [textNode(codeBuf.join('\n'))],
    })
  }
  return out
}

module.exports = { parse }
