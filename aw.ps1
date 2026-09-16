# aw.ps1 — gh aw control surface for aw-runner (Windows launcher)
#
# Usage (PowerShell):
#   .\aw.ps1 list -r bonsai/yose-db
#   .\aw.ps1 status -r bonsai/yose-db
#   .\aw.ps1 compile rakugo-zenza-update -o wf.yaml   (ローカル repo 内で実行)
#   .\aw.ps1 run rakugo-zenza-update -r bonsai/yose-db --dry-run
#   .\aw.ps1 logs rakugo-zenza-update -r bonsai/yose-db -o /tmp/awlogs
#
# Execution policy note: if this script is blocked (unsigned / UNC path), run with
#   powershell -NoProfile -ExecutionPolicy Bypass -File .\aw.ps1 <args>

$cmd = ($args -join ' ')
if (!$cmd) {
  Write-Host '.\.ps1 <list|status|run|logs> [args]  例: .\aw.ps1 list -r bonsai/yose-db' -ForegroundColor DarkGray
}
$runScript = "cd /home/bons/aw-runner && [ -x ./aw ] || go build -o aw ./cmd/aw; exec ./aw $cmd"
wsl -d Ubuntu -- bash -lc $runScript