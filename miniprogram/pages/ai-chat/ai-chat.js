// 配置你的 DeepSeek API Key
// 获取地址：https://platform.deepseek.com/
const DEEPSEEK_API_KEY = 'sk-0cf7d70499d14fdaad8923680669c292'  // 替换为你的真实 Key

Page({
  data: {
    messageList: [],
    inputValue: '',
    isLoading: false,
    scrollToView: '',
    userAvatar: '/assets/icons/1.png'  // 默认头像
  },

  onLoad() {
    // 获取用户头像
    const app = getApp()
    if (app.globalData.userInfo && app.globalData.userInfo.avatar_url) {
      this.setData({ userAvatar: app.globalData.userInfo.avatar_url })
    }
  },

  // 发送消息
  onSend(e) {
    const userMsg = e.detail.value
    if (!userMsg || !userMsg.trim()) return

    // 添加用户消息
    const userMessage = {
      role: 'user',
      content: userMsg,
      status: 'complete'
    }

    // 添加 AI 占位消息
    const aiMessage = {
      role: 'assistant',
      content: '',
      status: 'pending'
    }

    const newList = [...this.data.messageList, userMessage, aiMessage]
    this.setData({
      messageList: newList,
      inputValue: '',
      isLoading: true,
      scrollToView: `msg-${newList.length - 1}`
    })

    // 调用 DeepSeek API
    this.callDeepSeek()
  },

  // 调用 DeepSeek API
  async callDeepSeek() {
    // 构建历史消息（只取已完成的）
    const historyMessages = this.data.messageList
      .filter(msg => msg.status === 'complete')
      .map(msg => ({
        role: msg.role,
        content: msg.content
      }))

    wx.request({
      url: 'https://api.deepseek.com/v1/chat/completions',
      method: 'POST',
      header: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${DEEPSEEK_API_KEY}`
      },
      data: {
        model: 'deepseek-chat',
        messages: [
          { role: 'system', content: '你是一个友好、幽默的AI助手，用简洁有趣的方式回答问题。' },
          ...historyMessages
        ],
        temperature: 0.8,
        max_tokens: 1000
      },
      success: (res) => {
        if (res.statusCode === 200 && res.data.choices) {
          const reply = res.data.choices[0].message.content
          
          // 更新最后一条 AI 消息
          const list = this.data.messageList
          const lastIndex = list.length - 1
          list[lastIndex].content = reply
          list[lastIndex].status = 'complete'
          
          this.setData({
            messageList: list,
            isLoading: false,
            scrollToView: `msg-${lastIndex}`
          })
        } else {
          this.handleApiError()
        }
      },
      fail: (err) => {
        console.error('API 请求失败:', err)
        this.handleApiError()
      }
    })
  },

  // 处理 API 错误
  handleApiError() {
    const list = this.data.messageList
    const lastIndex = list.length - 1
    list[lastIndex].content = '抱歉，网络出了点问题，请稍后再试～'
    list[lastIndex].status = 'complete'
    
    this.setData({
      messageList: list,
      isLoading: false
    })
  },

  // 停止生成
  onStop() {
    this.setData({ isLoading: false })
  },

  // 分享功能（可选）
  onShareAppMessage() {
    return {
      title: 'AI 智能助手',
      path: '/pages/ai-chat/ai-chat'
    }
  }
})