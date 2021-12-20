package impltests

import (
	"encoding/base64"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PkijsRunner struct {
	tmpDir  string
	version string
}

func (p *PkijsRunner) Name() string {
	return "pkijs"
}

func (p *PkijsRunner) Initialize() error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	scriptPath := filepath.Join(tmpDir, "pkijs_test.js")
	err = ioutil.WriteFile(scriptPath, []byte(`
const asn1js = require('asn1js');
const pkijs = require('pkijs');

const { Crypto } = require('@peculiar/webcrypto');
const webcrypto = new Crypto();
pkijs.setEngine("newEngine", webcrypto, new pkijs.CryptoEngine({ name: "", crypto: webcrypto, subtle: webcrypto.subtle }));

function b64StringToCert(b64string) {
    let certificateBuffer = Buffer.from(b64string, 'base64');
    const asn1 = asn1js.fromBER(new Uint8Array(certificateBuffer).buffer);
    return new pkijs.Certificate({schema: asn1.result});
}

function verifyCertificate(trustRootAsB64String, certsAsArrayOfB64Strings) {
    const certificates = certsAsArrayOfB64Strings.map(b64StringToCert).reverse();
    const trustedCerts = [b64StringToCert(trustRootAsB64String)];

    const certChainVerificationEngine = new pkijs.CertificateChainValidationEngine({
        checkDate: new Date(),
        trustedCerts: trustedCerts,
        certs: certificates,
    });

    return certChainVerificationEngine.verify();
}

let trustRootAsB64String = process.argv[2];
let certsAsArrayOfB64Strings = process.argv[3].split(",");

let result = verifyCertificate(trustRootAsB64String, certsAsArrayOfB64Strings);

result.then(function(output) {
    if (output.result) {
        process.exit(0);
    } else {
        process.exit(1);
    }
}).catch(function(err) {
    console.log("Error: " + err);
    process.exit(1);
})
`), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("npm", "install", "pkijs", "@peculiar/webcrypto")
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return err
	}

	npmVersion, err := execAndCaptureInDir(tmpDir, "npm", "version")
	if err != nil {
		return err
	}

	modulesVersions, err := execAndCaptureInDir(tmpDir, "npm", "list")
	if err != nil {
		return err
	}

	p.version = npmVersion + "\n" + modulesVersions
	p.tmpDir = tmpDir
	return nil
}

func (p *PkijsRunner) Close() error {
	if p.tmpDir != "" {
		return os.RemoveAll(p.tmpDir)
	}
	return nil
}

func (p *PkijsRunner) GetVersion() string {
	return p.version
}

func (p *PkijsRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(p.tmpDir, "pkijs_test.js")
	return test_executor.ExecuteAllTestsLocal(ctx, suites, func(hostname string, certificates [][]byte) (bool, error) {
		certsB64 := make([]string, 0, len(certificates))
		for _, cert := range certificates {
			certsB64 = append(certsB64, base64.StdEncoding.EncodeToString(cert))
		}

		cmdParts := []string{"node", scriptPath,
			base64.StdEncoding.EncodeToString(suites.GetRootCert().Raw),
			strings.Join(certsB64, ",")}
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return false, err
		}
		err = cmdWaitWithTimeout(cmd, 5*time.Second)
		return err == nil, nil
	})
}
