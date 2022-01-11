package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os"
	"path/filepath"
)

type NodeRunner struct {
	version string
	tmpDir  string
}

func (c *NodeRunner) Name() string {
	return "node"
}

func (c *NodeRunner) Initialize() error {
	var err error
	c.version, err = execAndCapture("node", "--version")
	if err != nil {
		return err
	}

	c.tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.tmpDir, "foo.js"), []byte(`
const fs = require('fs');
const https = require('https');

var rootCa = fs.readFileSync(process.argv[2]);
var targetUrl = process.argv[3];

var req = https.request(targetUrl, {ca: rootCa}, function(res) {
  res.on('data', function(chunk) {
    // Ignored
  });
  res.on('end', function() {
    if (res.statusCode === 200) {
      process.exit(0);
    } else {
      process.exit(1);
    }
  });
});
req.on('error', function(e) {
  process.exit(1);
});
req.end();
`), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *NodeRunner) Close() error {
	if c.tmpDir != "" {
		return os.RemoveAll(c.tmpDir)
	}
	return nil
}

func (c *NodeRunner) GetVersion() string {
	return c.version
}

func (c *NodeRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"node", filepath.Join(c.tmpDir, "foo.js"), caPath,
			fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort),
		}
	})
}
