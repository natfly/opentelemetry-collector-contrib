// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package oci

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders/oci"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal"
)

func TestNewDetector(t *testing.T) {
	d, err := NewDetector(componenttest.NewNopProcessorCreateSettings(), nil)
	require.NoError(t, err)
	assert.NotNil(t, d)
}

func TestDetectAzureAvailable(t *testing.T) {
	mp := &oci.MockProvider{}
	mp.On("Metadata").Return(&oci.OciMetadataReponse{
		AvailabilityDomain:  "availabilityDomain",
		FaultDomain:         "faultDomain",
		CompartmentId:       "compartmentId",
		DisplayName:         "displayName",
		Hostname:            "hostname",
		Id:                  "hostId",
		TenantId:            "tenantId",
		Image:               "hostImageId",
		Region:              "region",
		CanonicalRegionName: "canonicalRegionName",
		OciAdName:           "availability",
		Shape:               "shape",
	}, nil)

	detector := &Detector{provider: mp}
	res, schemaURL, err := detector.Detect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, conventions.SchemaURL, schemaURL)
	mp.AssertExpectations(t)
	res.Attributes().Sort()

	expected := internal.NewResource(map[string]interface{}{
		conventions.AttributeCloudProvider:         "oci",
		conventions.AttributeCloudPlatform:         "oci_compute",
		conventions.AttributeCloudAccountID:        "tenantId",
		conventions.AttributeCloudRegion:           "canonicalRegionName",
		conventions.AttributeCloudAvailabilityZone: "availabilityDomain",
		conventions.AttributeHostID:                "hostId",
		conventions.AttributeHostImageID:           "hostImageId",
		"oci.compartment.id":                       "compartmentId",
		"oci.shape":                                "shape",
	})

	expected.Attributes().Sort()

	assert.Equal(t, expected, res)
}

func TestDetectError(t *testing.T) {
	mp := &oci.MockProvider{}
	mp.On("Metadata").Return(&oci.OciMetadataReponse{}, fmt.Errorf("mock error"))

	detector := &Detector{provider: mp, logger: zap.NewNop()}
	res, _, err := detector.Detect(context.Background())
	assert.NoError(t, err)
	assert.True(t, internal.IsEmptyResource(res))
}
