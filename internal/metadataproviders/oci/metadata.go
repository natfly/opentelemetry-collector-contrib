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

package oci

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	// Oci Instance metadata endpoint
	metadataEndpointV2   = "http://169.254.169.254/opc/v2/instance/"
	metadataEndpointV1   = "http://169.254.169.254/opc/v1/instance/"
	identityCertEndpoint = "http://169.254.169.254/opc/v2/identity/cert.pem"
)

// Provider gets metadata from the Oci instance metadata endpoint.
type Provider interface {
	Metadata(context.Context) (*OciMetadataReponse, error)
}

type ociProviderImpl struct {
	endpointV1           string
	endpointV2           string
	identityCertEndpoint string
	client               *http.Client
}

// NewProvider creates a new metadata provider
func NewProvider() Provider {
	return &ociProviderImpl{
		endpointV1:           metadataEndpointV1,
		endpointV2:           metadataEndpointV2,
		identityCertEndpoint: identityCertEndpoint,
		client:               &http.Client{},
	}
}

// OciMetadata is the OCI instance metadata response format
type OciMetadataReponse struct {
	TenantId            string
	AvailabilityDomain  string         `json:"availabilityDomain"`
	FaultDomain         string         `json:"faultDomain"`
	CompartmentId       string         `json:"compartmentId"`
	DisplayName         string         `json:"displayName"`
	Hostname            string         `json:"hostname"`
	Id                  string         `json:"id"`
	Image               string         `json:"image"`
	Region              string         `json:"region"`
	CanonicalRegionName string         `json:"canonicalRegionName"`
	OciAdName           string         `json:"ociAdName"`
	Shape               string         `json:"shape"`
	ShapeConfig         OciShapeConfig `json:"shapeConfig"`
}

type OciShapeConfig struct {
	OCpus                     float32 `json:"ocpus"`
	MemoryInGBs               float32 `json:"memoryInGBs"`
	NetworkingBandwidthInGbps float32 `json:"networkingBandwidthInGbps"`
	MaxVnicAttachments        int     `json:"maxVnicAttachments"`
}

// Metadata queries a given endpoint and parses the output
func (p *ociProviderImpl) Metadata(ctx context.Context) (*OciMetadataReponse, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpointV2, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer Oracle")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Oci instance metadata endpoint v2: %w", err)
	} else if resp.StatusCode != 200 {
		// Try the v1 metadata instance endpoint
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, p.endpointV1, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Add("Authorization", "Bearer Oracle")
		resp, err = p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to query Oci instance metadata endpiont v1: %w", err)
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf("oci instance metadata endpoint v1 replied with status code: %s", resp.Status)
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read oci metadata instance endpoint reply: %w", err)
	}

	var metadata *OciMetadataReponse
	err = json.Unmarshal(respBody, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Oci instance metadata reply: %w", err)
	}

	// Get tenant id from identity cert
	identityCert, err := getIdentityCertificate(ctx, p.client, p.identityCertEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve oci identity certificate: %w", err)
	}

	metadata.TenantId = extractTenancyIDFromCertificate(identityCert)

	return metadata, nil
}

func getIdentityCertificate(ctx context.Context, client *http.Client, certEndpoint string) (certificate *x509.Certificate, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer Oracle")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("oci instance certificate endpoint replied with status code: %s", resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read oci identity certificate endpoint reply: %w", err)
	}

	var block *pem.Block
	block, _ = pem.Decode(respBody)
	if block == nil {
		return nil, fmt.Errorf("failed to parse the certificate, not valid pem data")
	}

	if certificate, err = x509.ParseCertificate(block.Bytes); err != nil {
		return nil, fmt.Errorf("failed to parse the certificate: %s", err.Error())
	}

	return certificate, nil
}

func extractTenancyIDFromCertificate(cert *x509.Certificate) string {
	for _, nameAttr := range cert.Subject.Names {
		value := nameAttr.Value.(string)
		if strings.HasPrefix(value, "opc-tenant:") {
			return value[len("opc-tenant:"):]
		}
	}
	return ""
}
