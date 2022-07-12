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
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders/oci/oke"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.uber.org/zap"
)

const (
	// TypeStr is type of detector.
	TypeStr = "oke"
)

var _ internal.Detector = (*Detector)(nil)

// Detector is an OKE metadata detector
type Detector struct {
	provider oke.Provider
	logger   *zap.Logger
}

// NewDetector creates a new OKE metadata detector
func NewDetector(p component.ProcessorCreateSettings, cfg internal.DetectorConfig) (internal.Detector, error) {
	return &Detector{
		provider: oke.NewProvider(),
		logger:   p.Logger,
	}, nil
}

// Detect detects system metadata and returns a resource with the available ones
func (d *Detector) Detect(ctx context.Context) (resource pcommon.Resource, schemaURL string, err error) {
	res := pcommon.NewResource()
	attrs := res.Attributes()

	oke, err := d.provider.Metadata(ctx)
	if err != nil {
		d.logger.Debug("OKE detector metadata retrieval failed", zap.Error(err))
		// return an empty Resource and no error
		return res, "", nil
	}

	attrs.InsertString(conventions.AttributeCloudProvider, "oci")
	attrs.InsertString(conventions.AttributeCloudPlatform, "oci_oke")
	attrs.InsertString(conventions.AttributeCloudRegion, oke.CanonicalRegionName)
	attrs.InsertString(conventions.AttributeK8SClusterName, oke.Metadata.OkeClusterDisplayName)
	attrs.InsertString(conventions.AttributeCloudAccountID, oke.Metadata.OkeTenancyId)
	attrs.InsertString("oci.oke.clusterid", oke.Metadata.OkeClusterId)
	attrs.InsertString("oci.oke.k8version", oke.Metadata.OkeK8Version)

	return res, conventions.SchemaURL, nil
}
