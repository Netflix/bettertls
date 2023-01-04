package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Compatibility notes:
// This runner requires Administrator privileges so that it can add/remove root CA's.

type PowerShellRunner struct {
	version string
	tmpDir  string
}

func (c *PowerShellRunner) Name() string {
	return "powershell"
}

func (c *PowerShellRunner) Initialize() error {
	var err error
	c.version, err = execAndCapture("powershell", "$PSVersionTable.PSEdition + \" \" + $PSVersionTable.PSVersion")
	if err != nil {
		return err
	}

	c.tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.tmpDir, "try-tls-handshake.ps1"), []byte(`
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
`), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *PowerShellRunner) Close() error {
	if c.tmpDir != "" {
		return os.RemoveAll(c.tmpDir)
	}

	return nil
}

func (c *PowerShellRunner) GetVersion() string {
	return c.version
}

func (c *PowerShellRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"powershell", "-ExecutionPolicy", "Unrestricted", "-File", filepath.Join(c.tmpDir, "try-tls-handshake.ps1"), "-url", fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort), "-capath", caPath,
		}
	})
}
