# 🏃 aw-runner — Dispatch / Execute

**責務: どう起動するか**

ここが `gh aw` の境界。

```text
Router
  ↓
Runner
  ↓
gh aw
  ↓
GitHub Actions
  ↓
Agent / Workflow
```

## 持つもの

- `gh aw` integration
- dispatch / run / status / logs / retry
- execution context

## やらない

- Agent の思考 / Solve

## 構成

```text
cmd/          # CLI / TUI（aw-tui コントロールサーフェス, repo select）
repos/        # GitHub repository observation
graph/        # リポジトリグラフ（organization.json / .mmd）
api/          # REST 設計草案（openapi / ddl / task.schema）
workflows/
└── evolve.md # OBSERVE → EVOLVE 共通ループ
```

## 実行境界

```text
gh aw → aw-runner
gh wf → bonsai/workflow
```

## ビルド

```terminal
go build ./cmd/aw-tui
```

## 境界

- 送り先決定 → **bonsai/aw-router**
- 生成（proposal → artifact）→ **bonsai/aw-generator**
- 親・仕様・入口 → **bonsai/aw**