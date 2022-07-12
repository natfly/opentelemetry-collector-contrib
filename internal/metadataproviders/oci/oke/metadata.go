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

package oke

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// OKE Instance metadata endpoint
	metadataEndpointV2 = "http://169.254.169.254/opc/v2/instance/"
	metadataEndpointV1 = "http://169.254.169.254/opc/v1/instance/"
)

// Provider gets metadata from the Oke instance metadata endpoint.
type Provider interface {
	Metadata(context.Context) (*OkeMetadataReponse, error)
}

type okeProviderImpl struct {
	endpointV2 string
	endpointV1 string
	client     *http.Client
}

// NewProvider creates a new metadata provider
func NewProvider() Provider {
	return &okeProviderImpl{
		endpointV2: metadataEndpointV2,
		endpointV1: metadataEndpointV1,
		client:     &http.Client{},
	}
}

type OkeMetadataReponse struct {
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
	ShapeConfig         OkeShapeConfig `json:"shapeConfig"`
	Metadata            OkeMetadata    `json:"metadata"`
}

type OkeShapeConfig struct {
	OCpus                     float32 `json:"ocpus"`
	MemoryInGBs               float32 `json:"memoryInGBs"`
	NetworkingBandwidthInGbps float32 `json:"networkingBandwidthInGbps"`
	MaxVnicAttachments        int     `json:"maxVnicAttachments"`
}

// OkeMetadata is the OKE instance metadata response format
type OkeMetadata struct {
	OkeTm                 string `json:"oke-tm"`
	OkeK8Version          string `json:"oke-k8version"`
	OkePoolId             string `json:"oke-pool-id"`
	OkeTenancyId          string `json:"oke-tenancy-id"`
	OkeClusterDisplayName string `json:"oke-cluster-display-name"`
	OkeAvailabilityDomain string `json:"oke-ad"`
	OkeClusterId          string `json:"oke-cluster-id"`
	OkePrivateSubnet      string `json:"oke-is-on-private-subnet"`
	OkeImageName          string `json:"oke-image-name"`
}

// Metadata queries a given endpoint and parses the output
func (p *okeProviderImpl) Metadata(ctx context.Context) (*OkeMetadataReponse, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpointV2, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer Oracle")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Oke instance metadata endpoint v2: %w", err)
	} else if resp.StatusCode != 200 {
		// Try the v1 metadata instance endpoint
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, p.endpointV1, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Add("Authorization", "Bearer Oracle")
		resp, err = p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to query Oke instance metadata endpiont v1: %w", err)
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf("oke instance metadata endpoint v1 replied with status code: %s", resp.Status)
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read oke metadata instance endpoint reply: %w", err)
	}

	var metadata *OkeMetadataReponse
	err = json.Unmarshal(respBody, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Oke instance metadata reply: %w", err)
	}

	return metadata, nil
}
