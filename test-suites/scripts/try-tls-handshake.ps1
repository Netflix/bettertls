# Warning!  This script must be run in a fresh PowerShell process.  Otherwise,
# PowerShell will cache any successful cert validation results, so you'll be
# getting fictitious results.

param (
  $url,
  $capath
)

$caname = ((& "certutil.exe" "-f" "-enterprise" "-addstore" "Root" "$capath" | Select-String -Pattern 'Certificate ".*"').Matches[0].Value | Select-String -Pattern '".*"').Matches[0].Value.Trim('"')
If (!$?) {
  Write-Host "certificate trust failed"
  exit 1
}

try {
  Invoke-WebRequest -Uri "$url" -Method GET -UseBasicParsing
  $success = $?
}
catch {
  $success = $false
}

& "certutil.exe" "-enterprise" "-delstore" "Root" "$caname"
If (!$?) {
  Write-Host "certificate untrust failed"
  exit 1
}

if ($success) {
  exit 0
}

exit 1
