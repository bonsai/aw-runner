<#
  aw-tui.ps1 — aw-runner TUI launcher (Windows)

  Usage:
    .\aw-tui.ps1
    # [4] で AW ワークフロー一覧、enter=preview / d=dispatch
#>
wsl -d Ubuntu -- bash -lc "cd /home/bons/aw-runner && [ -x ./aw-tui ] || go build -o aw-tui ./cmd/aw-tui; exec ./aw-tui"