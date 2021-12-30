package nameconstraints

import (
	"crypto/x509/pkix"
	"encoding/asn1"
)

const (
	nameTypeDNS = 2
	nameTypeIP  = 7
)

func buildSanExtension(critical bool, sans []*ExtraSan) (pkix.Extension, error) {
	ext := pkix.Extension{
		Id:       []int{2, 5, 29, 17},
		Critical: critical,
	}

	var rawValues []asn1.RawValue
	for _, san := range sans {
		rawValues = append(rawValues, asn1.RawValue{Tag: san.tag, Class: 2, Bytes: san.value})
	}

	var err error
	ext.Value, err = asn1.Marshal(rawValues)
	return ext, err
}
