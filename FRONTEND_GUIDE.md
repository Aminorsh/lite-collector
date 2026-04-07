# 前端开发指南 — Lite Collector

## 我们在构建什么？

Lite Collector 是一个轻量级智能数据收集平台，旨在替代基于 Excel 的工作流程。可以把它理解为一个内置于微信的原生版 Google Forms，并集成了 AI 功能。

**它要解决的问题：** 目前团队通过微信分享 Excel 文件来收集数据。这种方式效率低下、容易出错，且无法进行清晰的数据汇总。

**工作流程：**

- **表单创建者**（如团队经理）在手机上设计表单并发布
- **提交者**（如团队成员）收到链接，在微信中打开并填写 — 无需注册，身份信息来自微信 OpenID
- 创建者可在统一界面查看所有提交记录
- **AI 在后台运行**，标记可疑条目（异常检测）并按需生成汇总报告

**平台：** 微信小程序（前端）+ Go/Gin REST API（后端）

---

## 身份认证

所有用户通过微信进行身份认证，不采用用户名/密码方式。

### 登录流程

```
1. 小程序调用 wx.login() → 获取临时 `code`
2. 将 code 发送至后端 → 获取 JWT token
3. 将 token 保存在本地
4. 在后续每个请求中携带 token：Authorization: Bearer <token>
```

### 登录接口

```
POST /api/v1/auth/wx-login
Content-Type: application/json

{ "code": "<wx.login() 获取的 code>" }
```

**成功响应 `200`：**

```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": 1,
    "openid": "oXxxx...",
    "nickname": "微信用户",
    "avatar_url": ""
  }
}
```

> **注意：** Token 有效期为 24 小时。当请求返回 `401` 时，需重新调用 wx.login() 并刷新 token。

---

## 基础 URL 与版本

所有 API 接口的基础路径为：

```
http://<host>/api/v1/
```

除登录接口外，每个请求都需要携带：

```
Authorization: Bearer <token>
```

---

## 错误格式

所有错误均遵循统一结构，与具体接口无关：

```json
{
  "error": {
    "code": "FORM_NOT_FOUND",
    "message": "form not found"
  }
}
```

错误处理时应根据 `code` 进行判断，而非 `message`（message 内容可能会变化）。常见错误码：

| Code                       | HTTP 状态码 | 含义                       |
| -------------------------- | ----------- | -------------------------- |
| `BAD_REQUEST`              | 400         | 缺少字段或字段格式错误     |
| `UNAUTHORIZED`             | 401         | 缺少 token 或 token 已过期 |
| `FORBIDDEN`                | 403         | 无权访问该资源             |
| `FORM_NOT_FOUND`           | 404         | 表单不存在                 |
| `FORM_FORBIDDEN`           | 403         | 你不是该表单的创建者       |
| `SUBMISSION_NOT_FOUND`     | 404         | 未找到提交记录             |
| `SUBMISSION_CREATE_FAILED` | 500         | 服务器保存提交记录失败     |
| `INTERNAL_ERROR`           | 500         | 服务器内部错误             |

---

## API 接口

### 表单

#### 创建表单

```
POST /api/v1/forms
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "2024 年度部门报告",
  "description": "请于周五前填写",
  "schema": "<JSON 字符串 — 详见下方表单结构说明>"
}
```

**成功响应 `201`：**

```json
{
  "id": 42,
  "title": "2024 年度部门报告",
  "description": "请于周五前填写",
  "status": 0,
  "created_at": "2026-04-07T10:00:00Z"
}
```

---

#### 获取我的表单列表

```
GET /api/v1/forms
Authorization: Bearer <token>
```

**成功响应 `200`：**

```json
{
  "forms": [
    {
      "id": 42,
      "title": "2024 年度部门报告",
      "status": 1,
      "created_at": "2026-04-07T10:00:00Z",
      "updated_at": "2026-04-07T12:00:00Z"
    }
  ]
}
```

---

#### 获取单个表单

```
GET /api/v1/forms/:formId
Authorization: Bearer <token>
```

**成功响应 `200`：**

```json
{
  "id": 42,
  "title": "2024 年度部门报告",
  "description": "请于周五前填写",
  "schema": "{...}",
  "status": 1,
  "created_at": "2026-04-07T10:00:00Z",
  "updated_at": "2026-04-07T12:00:00Z"
}
```

---

#### 更新表单

```
PUT /api/v1/forms/:formId
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "更新后的标题",
  "description": "更新后的描述",
  "schema": "<JSON 字符串>"
}
```

**成功响应 `200`：** 结构与创建响应相同。

---

#### 发布表单

将表单状态从 `草稿 (0)` 变更为 `已发布 (1)`。发布后，提交者即可填写。

```
POST /api/v1/forms/:formId/publish
Authorization: Bearer <token>
```

**成功响应 `200`：**

```json
{ "message": "form published successfully" }
```

---

### 提交记录

#### 提交表单

```
POST /api/v1/forms/:formId/submissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "f_001": "张三",
  "f_002": "技术部",
  "f_003": 42
}
```

请求体是一个扁平化的 key→value 映射，其中 key 是表单结构中定义的 `field_key` 值。

**成功响应 `201`：**

```json
{
  "id": 7,
  "status": 0,
  "submitted_at": "2026-04-07T14:30:00Z"
}
```

> 提交后，后端会自动在后台将 AI 异常检测任务加入队列。`status` 状态会异步更新（0 = 处理中，1 = 正常，2 = 存在异常）。

---

#### 获取我在某个表单中的提交记录

每个用户对每个表单只能提交一次。该接口返回用户的提交记录及填写的内容。

```
GET /api/v1/forms/:formId/submissions/my
Authorization: Bearer <token>
```

**成功响应 `200`：**

```json
{
  "id": 7,
  "status": 1,
  "submitted_at": "2026-04-07T14:30:00Z",
  "values": {
    "f_001": "张三",
    "f_002": "技术部",
    "f_003": 42
  }
}
```

---

## 表单结构

表单结构是一个存储在 `schema` 字段中的 JSON 字符串，用于描述表单包含的字段。

建议结构（待与后端最终确认）：

```json
{
  "fields": [
    {
      "key": "f_001",
      "label": "姓名",
      "type": "text",
      "required": true
    },
    {
      "key": "f_002",
      "label": "部门",
      "type": "select",
      "required": true,
      "options": ["技术部", "市场部", "财务部"]
    },
    {
      "key": "f_003",
      "label": "年龄",
      "type": "number",
      "required": false
    }
  ]
}
```

### 支持的字段类型

| 类型       | 说明           | 备注                        |
| ---------- | -------------- | --------------------------- |
| `text`     | 单行文本       |                             |
| `textarea` | 多行文本       |                             |
| `number`   | 数字输入       | 适用时验证 ≥ 0              |
| `select`   | 单选下拉框     | 需要 `options` 数组         |
| `radio`    | 单选框（行内） | 需要 `options` 数组         |
| `checkbox` | 多选框         | 返回选中值的数组            |
| `date`     | 日期选择器     | ISO 8601 格式：`YYYY-MM-DD` |
| `phone`    | 手机号         | 客户端进行格式校验          |
| `id_card`  | 身份证号       | 18 位校验                   |
| `image`    | 图片上传       | 上传成功后返回文件 URL      |

> `key` 值（如 `f_001`）是在提交时用作字段键的值。约定使用 `f_` + 补零后的索引，但任何唯一字符串均可。

---

## 表单与提交记录状态码

**表单状态（`status` 字段）：**

| 值   | 含义                          |
| ---- | ----------------------------- |
| `0`  | 草稿 — 仅创建者可见，不可提交 |
| `1`  | 已发布 — 开放提交             |
| `2`  | 已归档 — 关闭提交             |

**提交记录状态（`status` 字段）：**

| 值   | 含义                               |
| ---- | ---------------------------------- |
| `0`  | 处理中 — AI 审核进行中             |
| `1`  | 正常 — 未发现异常                  |
| `2`  | 存在异常 — AI 标记，需显示警告标识 |

---

## 当前 API 状态

> **重要提示：** 后端正在积极开发中。部分接口当前返回模拟/占位数据，数据库层正在实现中。**请求/响应结构和错误码已最终确定** — 可安全地基于此进行开发。

| 接口                            | 状态                                  |
| ------------------------------- | ------------------------------------- |
| `POST /auth/wx-login`           | 模拟（模拟 OpenID，尚未对接真实微信） |
| `POST /forms`                   | 真实（写入数据库）                    |
| `GET /forms`                    | 真实（从数据库读取）                  |
| `GET /forms/:id`                | 真实（从数据库读取）                  |
| `PUT /forms/:id`                | 真实（写入数据库）                    |
| `POST /forms/:id/publish`       | 真实（写入数据库）                    |
| `POST /forms/:id/submissions`   | 模拟（返回占位数据，尚未持久化）      |
| `GET /forms/:id/submissions/my` | 模拟（返回硬编码值）                  |

---

## 健康检查

```
GET /health
```

```json
{ "status": "ok" }
```

无需认证。用于检查服务器是否正常运行。

---

## 本地开发

使用 Docker Compose 启动本地后端：

```bash
docker-compose up
```

基础 URL：`http://localhost:8080`

当看到以下输出时，表示后端已就绪：

```
Server starting on :8080
```

接口在线文档（Swagger UI）：

```
http://localhost:8080/swagger/index.html
```

在 Swagger UI 中可以直接测试接口——点击右上角「Authorize」，输入登录后获取的 token，之后所有请求会自动携带认证信息。
