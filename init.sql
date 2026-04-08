-- Database initialisation for lite-collector
-- Applied automatically when the MySQL container first starts.
-- The Go app also runs GORM AutoMigrate on startup, so this file
-- primarily serves as documentation and a fast-path for fresh containers.

CREATE TABLE IF NOT EXISTS users (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    open_id     VARCHAR(64) UNIQUE NOT NULL,
    nickname    VARCHAR(64),
    avatar_url  VARCHAR(255),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forms (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    owner_id      BIGINT NOT NULL,
    title         VARCHAR(128) NOT NULL,
    description   TEXT,
    `schema`      JSON NOT NULL,
    status        TINYINT DEFAULT 0,
    template_year YEAR NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_forms_owner_id (owner_id),
    INDEX idx_forms_status (status),
    INDEX idx_forms_template_year (template_year)
);

CREATE TABLE IF NOT EXISTS submissions (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    form_id      BIGINT NOT NULL,
    submitter_id BIGINT NOT NULL,
    status       TINYINT DEFAULT 0,
    submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_submissions_form_id (form_id),
    INDEX idx_submissions_submitter_id (submitter_id),
    INDEX idx_submissions_status (status)
);

CREATE TABLE IF NOT EXISTS submission_values (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT,
    submission_id  BIGINT NOT NULL,
    field_key      VARCHAR(64) NOT NULL,
    value          TEXT,
    is_anomaly     TINYINT(1) DEFAULT 0,
    anomaly_reason VARCHAR(255),
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sv_submission_id (submission_id),
    INDEX idx_sv_field_key (field_key),
    INDEX idx_sv_is_anomaly (is_anomaly)
);

CREATE TABLE IF NOT EXISTS base_data (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    form_id    BIGINT NOT NULL,
    row_key    VARCHAR(64) NOT NULL,
    data       JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_base_data_form_id (form_id),
    INDEX idx_base_data_row_key (row_key)
);

CREATE TABLE IF NOT EXISTS ai_jobs (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    job_type    VARCHAR(32) NOT NULL,
    status      TINYINT DEFAULT 0,
    input       TEXT,
    output      TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME NULL,
    INDEX idx_ai_jobs_user_id (user_id),
    INDEX idx_ai_jobs_status (status),
    INDEX idx_ai_jobs_job_type (job_type)
);
