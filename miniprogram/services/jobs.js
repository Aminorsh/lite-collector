/**
 * AI job tracking — pending/recent jobs used by the global banner.
 *
 * Status contract (matches backend):
 *   0 queued | 1 processing | 2 done | 3 failed
 *
 * Dismiss model: only finished jobs (2/3) are dismissable. In-flight jobs (0/1)
 * always show until they transition. Dismissed IDs are kept in wx storage so
 * the banner doesn't reappear after the user has already seen the result.
 */

const api = require('./api')

const ACK_KEY = 'ackJobIds'
const ACK_MAX = 100 // cap the ack set so it can't grow unbounded

function readAcked() {
  try {
    var raw = wx.getStorageSync(ACK_KEY)
    if (!raw || !Array.isArray(raw)) return []
    return raw
  } catch (e) {
    return []
  }
}

function writeAcked(ids) {
  var trimmed = ids.slice(-ACK_MAX)
  try {
    wx.setStorageSync(ACK_KEY, trimmed)
  } catch (e) {
    // ignore
  }
}

function ack(jobId) {
  var ids = readAcked()
  if (ids.indexOf(jobId) === -1) {
    ids.push(jobId)
    writeAcked(ids)
  }
}

function isAcked(jobId) {
  return readAcked().indexOf(jobId) !== -1
}

/**
 * Fetch the user's pending + recently-finished jobs, filtered by ack state.
 * Returns array of visible job items, or [] on error.
 */
function fetchVisible() {
  return api.get('/jobs/pending')
    .then(function (res) {
      var jobs = (res && res.jobs) || []
      return jobs.filter(function (j) {
        // In-flight jobs always visible.
        if (j.status === 0 || j.status === 1) return true
        // Finished jobs visible until acknowledged.
        return !isAcked(j.id)
      })
    })
    .catch(function (err) {
      console.warn('[jobs] fetchVisible error:', err && err.message)
      return []
    })
}

/**
 * Resolve a human label and resume target for a job.
 * Returns { title, path } or null if no navigation makes sense.
 */
function describe(job) {
  var inFlight = job.status === 0 || job.status === 1
  var failed = job.status === 3

  if (job.job_type === 'generate_form') {
    if (inFlight) {
      return { title: 'AI 正在生成表单…', path: '/pages/ai-generate/ai-generate' }
    }
    if (failed) {
      return { title: 'AI 表单生成失败，点击查看', path: '/pages/ai-generate/ai-generate' }
    }
    return { title: 'AI 表单生成完成，点击查看', path: '/pages/ai-generate/ai-generate' }
  }

  if (job.job_type === 'generate_report') {
    var formPath = job.form_id
      ? '/pages/report/report?formId=' + job.form_id
      : null
    if (!formPath) return null
    if (inFlight) return { title: 'AI 正在生成报告…', path: formPath }
    if (failed)   return { title: 'AI 报告生成失败，点击查看', path: formPath }
    return { title: 'AI 报告生成完成，点击查看', path: formPath }
  }

  // detect_anomaly runs in the background silently — not surfaced.
  return null
}

module.exports = {
  fetchVisible,
  describe,
  ack,
}
