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
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders/oci"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.uber.org/zap"
)

const (
	// TypeStr is type of detector.
	TypeStr         = "oci"
	attributePrefix = "oci."
)

var _ internal.Detector = (*Detector)(nil)

// Detector is an OCI metadata detector
type Detector struct {
	provider        oci.Provider
	attributeJPaths []AttributeJPathConfig
	logger          *zap.Logger
}

// NewDetector creates a new OCI metadata detector
func NewDetector(p component.ProcessorCreateSettings, cfg internal.DetectorConfig) (internal.Detector, error) {
	config := cfg.(Config)
	return &Detector{
		provider:        oci.NewProvider(),
		logger:          p.Logger,
		attributeJPaths: config.AttributeJPaths,
	}, nil
}

// Detect detects system metadata and returns a resource with the available ones
func (d *Detector) Detect(ctx context.Context) (resource pcommon.Resource, schemaURL string, err error) {
	res := pcommon.NewResource()
	attrs := res.Attributes()

	oci, err := d.provider.Metadata(ctx)
	if err != nil {
		d.logger.Debug("OCI detector metadata retrieval failed", zap.Error(err))
		// return an empty Resource and no error
		return res, "", nil
	}

	attrs.InsertString(conventions.AttributeCloudProvider, "oci")
	attrs.InsertString(conventions.AttributeCloudAccountID, oci.TenantId)
	attrs.InsertString(conventions.AttributeCloudRegion, oci.CanonicalRegionName)
	attrs.InsertString(conventions.AttributeCloudAvailabilityZone, oci.AvailabilityDomain)
	attrs.InsertString(conventions.AttributeHostID, oci.Id)
	attrs.InsertString(conventions.AttributeHostImageID, oci.Image)
	attrs.InsertString("oci.compartment.id", oci.CompartmentId)
	attrs.InsertString("oci.shape", oci.Shape)

	if len(d.attributeJPaths) != 0 {
		for _, entry := range d.attributeJPaths {
			value := gjson.Get(oci.RawResponse, entry.Path)
			if value.Exists() {
				attrs.UpsertString(attributePrefix+entry.Name, value.String())
			}
		}
	}

	return res, conventions.SchemaURL, nil
}
