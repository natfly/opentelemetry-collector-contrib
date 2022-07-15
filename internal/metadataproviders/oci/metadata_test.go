// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint:errcheck
package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCert = `-----BEGIN CERTIFICATE-----
MIIDHTCCAgWgAwIBAgIUN4v0jRyyOLMckguORjMhWiJEhAQwDQYJKoZIhvcNAQEL
BQAwHjEcMBoGA1UECwwTb3BjLXRlbmFudDp0ZW5hbnRJZDAeFw0yMjA3MTUxODU0
MDJaFw0zMjA0MTMxODU0MDJaMB4xHDAaBgNVBAsME29wYy10ZW5hbnQ6dGVuYW50
SWQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDO8dk3hCZTiskiUyXC
d6QuEnhseE432oh56wRdA/TPrCzgY9EXLJjttfbdcnXeOjwSRQs5hlP5Ob/OFjDq
LtbTVliwIf5gcoPiZzzR/D5sI5AUaW5uiqLbBGf8xzDeo3lxU6dU/eooTeA1kMYV
QtH3QdAkhp6P/Tb1Vg0chUzGk4Y5MrtV6uV3VWOVpJv8Simcw9bxQhAFUhIxRJ2n
cjWQaLGgdZD/z60WBBOU6L7KDghB9uekYz6bxnj431e9RttwaVTFXwyQ/A12K4gQ
1TJyejmFIIE6i6A12rZspRktrOfiAN9W8hHEaQ8WtVTfs75bdco5a4IuoRPwmYZx
4chZAgMBAAGjUzBRMB0GA1UdDgQWBBRZRBzARXbgsoHR36hkK4a4E4KHRzAfBgNV
HSMEGDAWgBRZRBzARXbgsoHR36hkK4a4E4KHRzAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4IBAQBmRE4xd11BEBhf+hN28VSwZNgVGyzti/4VO+NWnh/6
YchxIcZ02NTKC2XP/abnkpLlwdWGWtrCqNW84KVqCIkrmPlhXV1wHiTmdCFj/qip
ZCi3SlwZB5jBBm4zb9aSBfdhHPpiU+/jhlHiVG1cSw/oZ9W653B/V2brn45eKVyd
5ifrA5kMxx78DAekVmODNHwPmhuOMgEqqMTPfjyDWOsycG/SrHBqk2marRRFInh8
Oi1ecAgr6kQEzDYwdtU80GExTZkUS61Gzvt1d2uT4KJrdhXI6cdeBaXCOKhrvaiL
6zh0yyMtq6NV9C/SMSXSP5Kjpwcxib1+fnKSoD7oBfvO
-----END CERTIFICATE-----`
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	assert.NotNil(t, provider)
}

func TestQueryEndpointFailed(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	provider := &ociProviderImpl{
		endpointV2:           ts.URL,
		endpointV1:           ts.URL,
		identityCertEndpoint: ts.URL,
		client:               &http.Client{},
	}

	_, err := provider.Metadata(context.Background())
	assert.Error(t, err)
}

func TestQueryEndpointMalformed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{")
	}))
	defer ts.Close()

	provider := &ociProviderImpl{
		endpointV2:           ts.URL,
		endpointV1:           ts.URL,
		identityCertEndpoint: ts.URL,
		client:               &http.Client{},
	}

	_, err := provider.Metadata(context.Background())
	assert.Error(t, err)
}

func TestQueryEndpointCorrect(t *testing.T) {
	sentMetadata := &OciMetadataReponse{
		AvailabilityDomain:  "availabilityDomain",
		FaultDomain:         "faultDomain",
		CompartmentId:       "compartmentId",
		DisplayName:         "displayName",
		Hostname:            "hostname",
		Id:                  "id",
		TenantId:            "tenantId",
		Image:               "image",
		Region:              "region",
		CanonicalRegionName: "canonicalRegionName",
		OciAdName:           "ociAdName",
		Shape:               "shape",
	}
	marshalledMetadata, err := json.Marshal(sentMetadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(marshalledMetadata)
	}))
	defer ts.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testCert))
	}))
	defer ts2.Close()

	provider := &ociProviderImpl{
		endpointV1:           ts.URL,
		endpointV2:           ts.URL,
		identityCertEndpoint: ts2.URL,
		client:               &http.Client{},
	}

	recvMetadata, err := provider.Metadata(context.Background())

	require.NoError(t, err)
	assert.Equal(t, *sentMetadata, *recvMetadata)
}
