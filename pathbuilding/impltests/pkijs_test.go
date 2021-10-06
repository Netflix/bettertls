package impltests

import (
	"encoding/base64"
	"github.com/Netflix/bettertls/pathbuilding"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPkiJs(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "pkijs_test.js")
	err := ioutil.WriteFile(scriptPath, []byte(`
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
		t.Fatal(err)
	}

	cmd := exec.Command("npm", "install", "pkijs", "@peculiar/webcrypto")
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	version, err := execAndCaptureInDir(tmpDir, "npm", "version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	version, err = execAndCaptureInDir(tmpDir, "npm", "list")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = pathbuilding.ExecuteTests(t, provider, func(testCaseName string) (bool, error) {
		testCase, err := provider.GetTestCase(testCaseName)
		if err != nil {
			return false, err
		}
		certsB64 := make([]string, 0, len(testCase.Certificates))
		for _, cert := range testCase.Certificates {
			certsB64 = append(certsB64, base64.StdEncoding.EncodeToString(cert))
		}

		cmdParts := []string{"node", scriptPath,
			base64.StdEncoding.EncodeToString(provider.GetRootCert().Raw),
			strings.Join(certsB64, ",")}
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return false, err
		}
		err = cmdWaitWithTimeout(t, cmd, 5*time.Second)
		return err == nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
