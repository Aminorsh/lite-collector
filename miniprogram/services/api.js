/**
 * HTTP client wrapper for the backend API.
 *
 * - Auto-injects Authorization header from storage
 * - On 401, triggers silent re-login then retries the original request once
 * - Surfaces error.code from the standard error envelope
 */

const storage = require('./storage')

const BASE_URL = 'http://localhost:8080/api/v1'

let _refreshPromise = null

/**
 * Core request function.
 * @param {Object} options - { url, method, data, header, skipAuth }
 * @returns {Promise<Object>} response data
 */
function request(options) {
  return new Promise((resolve, reject) => {
    const header = Object.assign({
      'Content-Type': 'application/json',
    }, options.header || {})

    if (!options.skipAuth) {
      const token = storage.getToken()
      if (token) {
        header['Authorization'] = 'Bearer ' + token
      }
    }

    wx.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header: header,
      success(res) {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
          return
        }

        // 401 → attempt silent re-login and retry once
        if (res.statusCode === 401 && !options._retried && !options.skipAuth) {
          handleUnauthorized()
            .then(() => {
              const retryOpts = Object.assign({}, options, { _retried: true })
              return request(retryOpts)
            })
            .then(resolve)
            .catch(reject)
          return
        }

        // Extract error from standard envelope
        const errData = res.data && res.data.error
        const error = new Error(errData ? errData.message : 'Request failed')
        error.code = errData ? errData.code : 'UNKNOWN'
        error.status = res.statusCode
        reject(error)
      },
      fail(err) {
        wx.showToast({ title: '网络连接失败', icon: 'none' })
        reject(new Error(err.errMsg || 'Network error'))
      },
    })
  })
}

/**
 * Handle 401 by re-running silent login. Deduplicates concurrent 401s.
 */
function handleUnauthorized() {
  if (_refreshPromise) return _refreshPromise

  const auth = require('./auth')
  _refreshPromise = auth.silentLogin()
    .finally(() => { _refreshPromise = null })

  return _refreshPromise
}

// Convenience methods
function get(url, data) {
  return request({ url, method: 'GET', data })
}

function post(url, data) {
  return request({ url, method: 'POST', data })
}

function put(url, data) {
  return request({ url, method: 'PUT', data })
}

function del(url, data) {
  return request({ url, method: 'DELETE', data })
}

module.exports = {
  request,
  get,
  post,
  put,
  del,
  BASE_URL,
}
