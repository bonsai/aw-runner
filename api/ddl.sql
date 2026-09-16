-- aw-api タスクキュー（Cloud SQL PostgreSQL）
-- 適用: gcloud sql databases create aw_tasks;  psql -f ddl.sql

CREATE TABLE IF NOT EXISTS aw_tasks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow     TEXT        NOT NULL,                -- gh aw workflow id（例: addTopics）
    owner_repo   TEXT        NOT NULL DEFAULT 'bonsai/aw',
    engine       TEXT        NOT NULL DEFAULT 'claude' CHECK (engine IN ('claude','codex','copilot','gemini','pi')),
    inputs       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    status       TEXT        NOT NULL DEFAULT 'queued'
                 CHECK (status IN ('queued','claimed','dispatched','running','completed','failed','canceled')),
    gh_run_id    BIGINT,
    gh_run_url   TEXT,
    result       JSONB,                               -- Result JSON（safe-outputs）
    aic_used     NUMERIC,
    max_aic      NUMERIC,
    error        TEXT,
    retry_count  INT         NOT NULL DEFAULT 0,
    claimed_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_aw_tasks_claim
    ON aw_tasks (status, created_at);

-- worker 用 claim（SKIP LOCKED で並列安全）
-- UPDATE aw_tasks
--   SET status='claimed', claimed_at=now()
--   WHERE id = (
--     SELECT id FROM aw_tasks
--     WHERE status='queued'
--     ORDER BY created_at LIMIT 1
--     FOR UPDATE SKIP LOCKED
--   )
--   RETURNING *;