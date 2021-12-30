package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os"
	"path/filepath"
)

type PythonRequestsRunner struct {
	version string
	tmpDir  string
}

func (c *PythonRequestsRunner) Name() string {
	return "python_requests"
}

func (c *PythonRequestsRunner) Initialize() error {
	var err error
	c.version, err = execAndCapture("python3", "--version")
	if err != nil {
		return err
	}

	c.tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.tmpDir, "foo.py"), []byte(`import requests
import sys
#from requests.packages.urllib3.contrib.pyopenssl import inject_into_urllib3
#inject_into_urllib3()
from requests.packages.urllib3.contrib.pyopenssl import extract_from_urllib3
extract_from_urllib3()
r = requests.get(sys.argv[2], verify=sys.argv[1])
if r.status_code != 200:
    sys.exit(1)
`), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *PythonRequestsRunner) Close() error {
	if c.tmpDir != "" {
		return os.RemoveAll(c.tmpDir)
	}
	return nil
}

func (c *PythonRequestsRunner) GetVersion() string {
	return c.version
}

func (c *PythonRequestsRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"python3", filepath.Join(c.tmpDir, "foo.py"), caPath,
			fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort),
		}
	})
}
