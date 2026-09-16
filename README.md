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
cmd/          # CLI / TUI（aw コントロールCLI, aw-tui コントロールサーフェス, repo select）
ghaw/         # `gh aw` サブプロセスラッパー（list/status/run/logs）
repos/        # GitHub repository observation
graph/        # リポジトリグラフ（organization.json / .mmd）
api/          # REST 設計草案（openapi / ddl / task.schema）
workflows/
└── evolve.md # OBSERVE → EVOLVE 共通ループ
```

## CLI（gh aw コントロールサーフェス）

`cmd/aw` は `gh aw` の実コマンドを内部で呼ぶだけ。思考や Solve はしない。

```terminal
go build -o aw ./cmd/aw

./aw list                     # ワークフロー一覧
./aw status [pattern]         # ワークフロー状態
./aw run <workflow>           # dispatch（workflow_dispatch）
./aw run <workflow> --dry-run # 実行せずプレビュー
./aw run <workflow> --raw-field foo=bar --raw-field env=prod
./aw logs [workflow] -o dir   # 実行ログ / アーティファクト取得
```

どのコマンドも `-r owner/repo` で対象リポジトリを指定できる（省略時はカレント）。

## TUI（aw-tui）

`[4] AW` タブで選択リポジトリのワークフロー一覧を表示し、`enter` でプレビュー / `d` で dispatch できる。

```terminal
go build -o aw-tui ./cmd/aw-tui && ./aw-tui
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