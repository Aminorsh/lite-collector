# Frontend Developer Guide — Lite Collector

## What are we building?

Lite Collector is a lightweight intelligent data collection platform designed to replace Excel-based workflows. Think of it as a WeChat-native Google Forms with AI built in.

**The problem it solves:** teams currently collect data by sharing Excel files over WeChat. This is slow, error-prone, and impossible to aggregate cleanly.

**How it works:**
- A **form owner** (e.g. a manager) designs a form on their phone and publishes it
- **Submitters** (e.g. team members) receive a link, open it in WeChat, and fill it in — no registration needed, identity comes from WeChat OpenID
- The owner reviews all submissions in one place
- **AI runs in the background** to flag suspicious entries (anomaly detection) and generate summary reports on demand

**Platform:** WeChat Mini Program (frontend) + Go/Gin REST API (backend)

---

## Authentication

All users authenticate via WeChat. There is no username/password.

### Login flow

```
1. Mini Program calls wx.login() → gets a temporary `code`
2. Send code to our backend → get a JWT token back
3. Store the token locally
4. Attach token to every subsequent request as: Authorization: Bearer <token>
```

### Login endpoint

```
POST /api/v1/auth/wx-login
Content-Type: application/json

{ "code": "<code from wx.login()>" }
```

**Success response `200`:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": 1,
    "openid": "oXxxx...",
    "nickname": "WeChat User",
    "avatar_url": ""
  }
}
```

> **Note:** Token expiry is 24 hours. When a request returns `401`, call wx.login() again and refresh the token.

---

## Base URL & versioning

All API endpoints are under:

```
http://<host>/api/v1/
```

Every request except login requires:

```
Authorization: Bearer <token>
```

---

## Error format

All errors follow the same structure regardless of endpoint:

```json
{
  "error": {
    "code": "FORM_NOT_FOUND",
    "message": "form not found"
  }
}
```

You should switch on `code` in your error handling, not on `message` (messages may change). Common codes:

| Code | HTTP Status | Meaning |
|---|---|---|
| `BAD_REQUEST` | 400 | Missing or malformed fields |
| `UNAUTHORIZED` | 401 | Missing or expired token |
| `FORBIDDEN` | 403 | You don't have access to this resource |
| `FORM_NOT_FOUND` | 404 | Form doesn't exist |
| `FORM_FORBIDDEN` | 403 | You don't own this form |
| `SUBMISSION_NOT_FOUND` | 404 | No submission found |
| `SUBMISSION_CREATE_FAILED` | 500 | Server error saving submission |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## API Endpoints

### Forms

#### Create a form
```
POST /api/v1/forms
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "2024 Annual Department Report",
  "description": "Please fill in before Friday",
  "schema": "<JSON string — see Form Schema section below>"
}
```

**Success `201`:**
```json
{
  "id": 42,
  "title": "2024 Annual Department Report",
  "description": "Please fill in before Friday",
  "status": 0,
  "created_at": "2026-04-07T10:00:00Z"
}
```

---

#### List my forms
```
GET /api/v1/forms
Authorization: Bearer <token>
```

**Success `200`:**
```json
{
  "forms": [
    {
      "id": 42,
      "title": "2024 Annual Department Report",
      "status": 1,
      "created_at": "2026-04-07T10:00:00Z",
      "updated_at": "2026-04-07T12:00:00Z"
    }
  ]
}
```

---

#### Get a single form
```
GET /api/v1/forms/:formId
Authorization: Bearer <token>
```

**Success `200`:**
```json
{
  "id": 42,
  "title": "2024 Annual Department Report",
  "description": "Please fill in before Friday",
  "schema": "{...}",
  "status": 1,
  "created_at": "2026-04-07T10:00:00Z",
  "updated_at": "2026-04-07T12:00:00Z"
}
```

---

#### Update a form
```
PUT /api/v1/forms/:formId
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated title",
  "description": "Updated description",
  "schema": "<JSON string>"
}
```

**Success `200`:** same shape as Create response.

---

#### Publish a form
Moves status from `draft (0)` to `published (1)`. Once published, submitters can fill it in.

```
POST /api/v1/forms/:formId/publish
Authorization: Bearer <token>
```

**Success `200`:**
```json
{ "message": "form published successfully" }
```

---

### Submissions

#### Submit a form
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

The request body is a flat key→value map where keys are the `field_key` values defined in the form schema.

**Success `201`:**
```json
{
  "id": 7,
  "status": 0,
  "submitted_at": "2026-04-07T14:30:00Z"
}
```

> After submission, the backend automatically enqueues an AI anomaly detection job in the background. The `status` will update asynchronously (0 = pending, 1 = normal, 2 = has anomaly).

---

#### Get my submission for a form
Each user can only submit once per form. This returns that submission along with the values they entered.

```
GET /api/v1/forms/:formId/submissions/my
Authorization: Bearer <token>
```

**Success `200`:**
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

## Form Schema

The form schema is a JSON string stored in the `schema` field. It describes what fields the form contains.

Proposed structure (subject to finalisation with backend):

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

### Supported field types

| Type | Description | Notes |
|---|---|---|
| `text` | Single-line text | |
| `textarea` | Multi-line text | |
| `number` | Numeric input | Validate ≥ 0 where applicable |
| `select` | Single choice dropdown | Requires `options` array |
| `radio` | Single choice inline | Requires `options` array |
| `checkbox` | Multiple choice | Returns array of selected values |
| `date` | Date picker | ISO 8601 format: `YYYY-MM-DD` |
| `phone` | Phone number | Client-side format validation |
| `id_card` | Chinese ID card | 18-digit validation |
| `image` | Image upload | Returns file URL after upload |

> The `key` value (e.g. `f_001`) is what you use as the field key when submitting. Convention is `f_` + zero-padded index, but any unique string works.

---

## Form & Submission status codes

**Form status (`status` field):**
| Value | Meaning |
|---|---|
| `0` | Draft — only visible to owner, cannot be submitted |
| `1` | Published — open for submissions |
| `2` | Archived — closed |

**Submission status (`status` field):**
| Value | Meaning |
|---|---|
| `0` | Pending — AI review in progress |
| `1` | Normal — no anomalies found |
| `2` | Has anomaly — AI flagged something, show a warning indicator |

---

## Current API status

> **Important:** The backend is under active development. Some endpoints currently return mock/placeholder data while the database layer is being implemented. The **request/response shapes and error codes are final** — you can build against them safely.

| Endpoint | Status |
|---|---|
| `POST /auth/wx-login` | Mock (simulated OpenID, no real WeChat exchange yet) |
| `POST /forms` | Real (writes to DB) |
| `GET /forms` | Real (reads from DB) |
| `GET /forms/:id` | Real (reads from DB) |
| `PUT /forms/:id` | Real (writes to DB) |
| `POST /forms/:id/publish` | Real (writes to DB) |
| `POST /forms/:id/submissions` | Mock (returns placeholder, not persisted yet) |
| `GET /forms/:id/submissions/my` | Mock (returns hardcoded values) |

---

## Health check

```
GET /health
```

```json
{ "status": "ok" }
```

No auth required. Use this to check if the server is up.

---

## Local development

Start the backend locally with Docker Compose:

```bash
docker-compose up
```

Base URL: `http://localhost:8080`

The backend will be ready when you see:
```
Server starting on :8080
```
