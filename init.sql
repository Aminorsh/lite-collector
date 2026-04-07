-- Database schema for lite-collector
-- This file is automatically executed when the MySQL container starts

-- Users table (WeChat OpenID based)
CREATE TABLE IF NOT EXISTS users (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    openid      VARCHAR(64) UNIQUE NOT NULL,
    nickname    VARCHAR(64),
    avatar_url  VARCHAR(255),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Forms table
CREATE TABLE IF NOT EXISTS forms (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    owner_id      BIGINT NOT NULL,
    title         VARCHAR(128) NOT NULL,
    description   TEXT,
    form_schema   JSON NOT NULL,
    status        TINYINT DEFAULT 0,       -- 0:draft 1:published 2:archived
    template_year YEAR NULL,               -- set only when used as annual template
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    INDEX idx_forms_owner_id (owner_id),
    INDEX idx_forms_status (status),
    INDEX idx_forms_template_year (template_year)
);

-- Form templates table (for annual cloning)
CREATE TABLE IF NOT EXISTS form_templates (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_form_id BIGINT NOT NULL,
    cloned_form_id BIGINT NOT NULL,
    year           YEAR NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Submissions table
CREATE TABLE IF NOT EXISTS submissions (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    form_id      BIGINT NOT NULL,
    submitter_id BIGINT NOT NULL,
    status       TINYINT DEFAULT 0,        -- 0:pending 1:normal 2:has_anomaly
    submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_id) REFERENCES forms(id),
    FOREIGN KEY (submitter_id) REFERENCES users(id),
    INDEX idx_submissions_form_id (form_id),
    INDEX idx_submissions_submitter_id (submitter_id),
    INDEX idx_submissions_status (status)
);

-- Submission values table (EAV model for flexible form fields)
CREATE TABLE IF NOT EXISTS submission_values (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT,
    submission_id  BIGINT NOT NULL,
    field_key      VARCHAR(64) NOT NULL,   -- matches key in form schema
    value          TEXT,                   -- all values stored as string
    is_anomaly     TINYINT DEFAULT 0,
    anomaly_reason VARCHAR(255),
    FOREIGN KEY (submission_id) REFERENCES submissions(id),
    INDEX idx_sv_submission_id (submission_id),
    INDEX idx_sv_field_key (field_key),
    INDEX idx_sv_is_anomaly (is_anomaly)
);

-- Base data table (for prefilling)
CREATE TABLE IF NOT EXISTS base_data (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    form_id    BIGINT NOT NULL,
    row_key    VARCHAR(64) NOT NULL,       -- lookup key (e.g. employee ID)
    data       JSON NOT NULL,              -- prefill values for this record
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_id) REFERENCES forms(id),
    INDEX idx_base_data_form_id (form_id),
    INDEX idx_base_data_row_key (row_key)
);

-- AI jobs table (for tracking async AI tasks)
CREATE TABLE IF NOT EXISTS ai_jobs (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    type        VARCHAR(32) NOT NULL,      -- generate_form | generate_report | detect_anomaly
    status      TINYINT DEFAULT 0,         -- 0:queued 1:processing 2:done 3:failed
    input       TEXT,
    output      TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME NULL,
    INDEX idx_ai_jobs_user_id (user_id),
    INDEX idx_ai_jobs_status (status),
    INDEX idx_ai_jobs_type (type)
);