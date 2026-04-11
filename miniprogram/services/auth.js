/**
 * Auth service — handles WeChat silent login flow.
 *
 * Flow: wx.login() → get code → POST /auth/wx-login → save JWT + user info
 */

const storage = require('./storage')

/**
 * Perform silent login via WeChat.
 * @returns {Promise<Object>} user info { id, openid, nickname, avatar_url }
 */
function silentLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success(loginRes) {
        if (!loginRes.code) {
          reject(new Error('wx.login failed: no code'))
          return
        }
        // Lazy-require api to avoid circular dependency at module load time
        const api = require('./api')
        api.post('/auth/wx-login', { code: loginRes.code })
          .then((data) => {
            storage.setToken(data.token)
            storage.setUser(data.user)

            const app = getApp()
            if (app) {
              app.globalData.token = data.token
              app.globalData.userInfo = data.user
            }

            resolve(data.user)
          })
          .catch(reject)
      },
      fail(err) {
        reject(new Error('wx.login failed: ' + (err.errMsg || '')))
      },
    })
  })
}

/**
 * Check whether user is logged in (has a token).
 */
function isLoggedIn() {
  return !!storage.getToken()
}

/**
 * Log out — clear stored credentials.
 */
function logout() {
  storage.clear()
  const app = getApp()
  if (app) {
    app.globalData.token = ''
    app.globalData.userInfo = null
  }
}

module.exports = {
  silentLogin,
  isLoggedIn,
  logout,
}
