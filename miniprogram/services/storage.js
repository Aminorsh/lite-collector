/**
 * Storage service — thin wrapper over wx sync storage.
 */

const KEYS = {
  TOKEN: 'auth_token',
  USER: 'user_info',
}

function getToken() {
  return wx.getStorageSync(KEYS.TOKEN) || ''
}

function setToken(token) {
  wx.setStorageSync(KEYS.TOKEN, token)
}

function getUser() {
  return wx.getStorageSync(KEYS.USER) || null
}

function setUser(user) {
  wx.setStorageSync(KEYS.USER, user)
}

function clear() {
  wx.removeStorageSync(KEYS.TOKEN)
  wx.removeStorageSync(KEYS.USER)
}

module.exports = {
  getToken,
  setToken,
  getUser,
  setUser,
  clear,
}
