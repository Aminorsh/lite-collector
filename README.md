# Lite Collector

An intelligent data collection platform with AI-powered anomaly detection and report generation. Built for WeChat Mini Program as the frontend client.

## Features

- **Form Management** - Create, publish, archive forms with flexible JSON schema
- **AI Form Generation** - Describe a form in natural language, get a ready-to-use schema
- **Data Collection** - EAV-based submission system supporting dynamic field types
- **AI Anomaly Detection** - Automatic data quality checks on every submission via DeepSeek
- **AI Report Generation** - On-demand summary reports with statistics and insights
- **Base Data** - Import reference data for form field prefilling (e.g. employee rosters)
- **WeChat Login** - Native `wx.login` integration with JWT session management

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Frontend | WeChat Mini Program (native) |
| Backend | Go + Gin + GORM |
| Database | MySQL 8.0 |
| Cache | Redis |
| AI | DeepSeek API |
| Auth | WeChat Mini Program + JWT |
| Docs | Swagger (swaggo) |
| Deployment | Docker Compose |

## Quick Start

### Prerequisites

- Go 1.20+
- MySQL 8.0
- Redis
- Docker & Docker Compose (for containerized deployment)

### Local Development

```bash
# Clone the repo
git clone https://github.com/Aminorsh/lite-collector.git
cd lite-collector

# Configure environment
cp .env.example .env
# Edit .env with your database credentials and API keys

# Initialize database
go run scripts/init_db.go

# Run the server
go run main.go
```

The server starts at `http://localhost:8080`. Swagger UI is available at `http://localhost:8080/swagger/index.html`.

### Frontend (WeChat Mini Program)

1. Open [WeChat DevTools](https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html) (macOS / Windows)
2. Import the `miniprogram/` directory as a Mini Program project
3. Set `appid` in `project.config.json` to your own, or use the test AppID
4. The app connects to `http://localhost:8080` by default — make sure the backend is running

> **Note:** WeChat DevTools is not available on Linux. The code is standard Mini Program JS/WXML/WXSS and can be developed on any platform, then previewed on a machine with DevTools.

### Docker Compose (Full Stack)

Run the entire stack (MySQL + Redis + backend) with a single command:

```bash
docker-compose -f docker-compose.prod.yml up --build
```

For development (MySQL + Redis only, run backend locally):

```bash
docker-compose up
go run main.go
```

## API Overview

All endpoints (except login) require JWT authentication via the `Authorization` header.

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/wx-login` | WeChat login, returns JWT |
| PUT | `/api/v1/user/profile` | Update nickname/avatar |

### Forms
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forms/` | Create a draft form |
| POST | `/api/v1/forms/generate` | AI: generate form from description |
| GET | `/api/v1/forms/` | List my forms |
| GET | `/api/v1/forms/:id` | Get form detail (owner) |
| GET | `/api/v1/forms/:id/schema` | Get published form (any user) |
| PUT | `/api/v1/forms/:id` | Update form |
| POST | `/api/v1/forms/:id/publish` | Publish form |
| POST | `/api/v1/forms/:id/archive` | Archive form |
| POST | `/api/v1/forms/:id/report` | AI: generate summary report |

### Submissions
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forms/:id/submissions/` | Submit form data |
| GET | `/api/v1/forms/:id/submissions/` | List all submissions (owner) |
| GET | `/api/v1/forms/:id/submissions/my` | Get my submission |
| GET | `/api/v1/forms/:id/submissions/overview` | Full table view with anomaly reasons |
| GET | `/api/v1/forms/:id/submissions/:sid` | Get one submission detail (owner) |

### Base Data
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forms/:id/base-data/` | Batch import reference data |
| GET | `/api/v1/forms/:id/base-data/` | List all reference data (owner) |
| GET | `/api/v1/forms/:id/base-data/lookup` | Lookup by row_key (prefill) |
| DELETE | `/api/v1/forms/:id/base-data/` | Clear reference data |

### AI Jobs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/jobs/:id` | Check AI job status |

## AI Features

### Anomaly Detection

Every submission automatically triggers an AI review. The DeepSeek model checks for:
- Values that don't match expected types or formats
- Unrealistic values (e.g. age = 500)
- Cross-field inconsistencies

Results are stored per-submission and surfaced via the overview endpoint.

### Report Generation

Form owners can trigger a summary report that analyzes all submissions and produces:
- Submission statistics (total, normal, anomalous)
- Key metrics for numeric fields (min, max, average)
- Distribution analysis for text fields
- Anomaly summary with details
- Improvement recommendations

### Form Generation

Describe a form in natural language and get a complete schema:

```bash
curl -X POST /api/v1/forms/generate \
  -d '{"description": "Employee travel expense report with traveler name, dates, destination, transport/hotel/meal costs, and reason"}'
```

Returns `title`, `description`, and `schema` ready to pass to `POST /forms/`.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `JWT_SECRET` | JWT signing secret | `change-me-in-production` |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL user | `root` |
| `DB_PASSWORD` | MySQL password | `password` |
| `DB_NAME` | Database name | `lite_collector` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `WX_APP_ID` | WeChat Mini Program AppID | (empty = simulated login) |
| `WX_APP_SECRET` | WeChat Mini Program AppSecret | (empty = simulated login) |
| `DEEPSEEK_API_KEY` | DeepSeek API key | (empty = AI features disabled) |

## Project Structure

```
.
├── main.go                  # Entry point, wiring
├── config/                  # Environment config loading
├── handlers/                # HTTP handlers (thin: bind → call → respond)
├── services/                # Business logic layer
├── repository/              # Data access layer (GORM)
├── models/                  # Database models
├── middleware/              # JWT auth middleware
├── routes/                  # Route registration
├── jobs/                    # Background AI worker
├── utils/                   # Typed app errors
├── docs/                    # Generated Swagger docs
├── docker-compose.yml       # Dev: MySQL + Redis
├── docker-compose.prod.yml  # Prod: MySQL + Redis + backend
├── Dockerfile               # Multi-stage build
└── miniprogram/             # WeChat Mini Program frontend
    ├── app.js / app.json    # App entry and config (2-tab layout)
    ├── services/            # API client, auth, storage
    ├── utils/               # Constants, schema helpers, validation
    ├── components/          # field-renderer, form-renderer, status-badge, empty-state
    └── pages/               # 10 pages (see below)
        ├── index/           # Form list + create FAB
        ├── form-editor/     # Visual schema editor (10 field types)
        ├── form-detail/     # Status actions (publish, share, archive)
        ├── form-fill/       # Form filling with validation + prefill
        ├── submissions/     # List + overview table with anomaly flags
        ├── submission-detail/ # Single submission read-only view
        ├── base-data/       # Import/list/clear reference data
        ├── ai-generate/     # AI form generation from description
        ├── report/          # AI report with async job polling
        └── profile/         # Nickname editing + logout
```

## License

MIT
