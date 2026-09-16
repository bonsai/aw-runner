# aw-api 設計（Spec 草案 v0.1 — 2026-09-08）

決定事項（人間承認済）: **DB キュー主線 + GCP Cloud Run REST**。
下記は `設計.md`＝Spec。Gate 規則により、本ドキュメント承認後に PoC を開始する。

## ゴール

aw のワークフロー実行（gh aw run ≒ workflow_dispatch）を、ローカル CLI に依存せず
opencode / 各エージェント / curl から REST API で依頼できるようにする。
タスク/結果は DB で**正本化**（repos#7 の data 正本化の対象としても扱える）。

```
[Client: opencode / agents / curl]
        │ HTTPS (IAP)
        ▼
┌──────────────────────────┐  Cloud Run (min0)      ┌─────────────────┐
│  aw-api  REST            │  POST /v1/tasks        │  DB aw_tasks    │
│  enqueue/status/result   ├───────────────────────▶│  (Cloud SQL)    │
└──────────────────────────┘ ◀──────────────────────┘  JSONB          │
        ▲                                            ▲                  │
        │ poll: Cloud Scheduler→worker job          │ claim/readback   │
        ▼                                            │                  │
┌──────────────────────────┐  GitHub(dispatch)      └─────────────────┘
│  aw-worker (Cloud Run    │  gh aw run <wf> /       │
│  job, every 60s)         │  gh workflow run → GH Actions
└──────────────────────────┘  lock.yml running Claude/Codex
```

## コンポーネント

### 1. aw-api（Cloud Run、REST）

- `POST /v1/tasks` — タスク enqueue（202 + task row）
- `GET  /v1/tasks?status=&workflow=&limit=` — 一覧
- `GET  /v1/tasks/{id}` — 詳細（status / gh_run_url / result）
- `GET  /v1/tasks/{id}/result` — 完了 Result JSON（safe-outputs 成果物）
- `POST /v1/tasks/{id}/cancel` — キュー中のみキャンセル
- 制御面（gh aw ミラー）: `POST /v1/workflows/{id}/validate` `POST /v1/workflows/{id}/compile` `GET /v1/forecast?workflow=`
- `GET /healthz`
- 認証: Cloud Run IAM（人間=@bonsai 等）/ エージェント=SA OIDC token → IAP で通す

### 2. aw-worker（Cloud Run job、Cloud Scheduler で60s毎）

ループ:
1. **Claim**: `UPDATE aw_tasks SET status='claimed' WHERE id=(SELECT id … ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)`
2. **Compile 確認**: 指定 repo の `.github/workflows/<id>.lock.yml` を確認。無ければ `gh aw compile <id> --engine <engine>`
3. **Dispatch**: `gh aw run <id> --json --raw-field aw_task_id=<id> …`（= `gh workflow run <id>.yml` 経由）
4. **Poll**: `gh run view <gh_run_id> --json status,conclusion,url` を終端まで待つ
5. **Readback**: `gh run download` で `result.json`（または safe-outputs outcomes）→ DB に status / result / aic_used を書込み
6. 失敗時: retry_count+1、上限まで再試行。最終失敗は status=failed + error を result に

### 3. DB（Cloud SQL PostgreSQL、正本）

- `aw_tasks` 表（下記 ddl.sql）。タスク入力・Result は JSONB
- 代替: Firestore（serverless）も可。Cloud SQL を推奨（JOIN・保留制約・JSONB が素直）

## 契約（Task / Result）

- Task JSON: `api/task.schema.json`（`workflow`, `owner_repo`, `inputs`, `engine`, `max_aic`）
- Result JSON: 既存 `bonsai/gh-aw` の result 契約と整合（safe-outputs 成果物 = 検証済 evidence）
- 状態遷移: `queued → claimed → dispatched → running → completed | failed | canceled`

## GCP 配線（Deploy フェーズで実施、PoC 後）

- `gcloud run deploy aw-api --region asia-northeast1 --allow-unauthenticated`（IAP 前）
- Cloud Run job `aw-worker` + `gcloud scheduler jobs create http …`
- GitHub 認証: **GitHub App `bonsai-aw-app`**（actions:write / contents:read / issues:write / workflows:write）、
  1h token を worker が取得して dispatch に使用。PAT 非推奨
- コストガード: `max_aic` 上限 + `gh aw forecast` 確認ポリシー + worker 並列数制御

## 未決定（レビューで決定する）

1. DB = Cloud SQL PostgreSQL or Firestore
2. IAP導入時期（PoC=無効 → 本番=有効）
3. 認証の単位（個人/ org / 各 agent）
4. `gh aw run` 直送と DB キューの中間（即時 vs 定期）は aw-api の `POST /v1/tasks {schedule:none|cron}` で拡張