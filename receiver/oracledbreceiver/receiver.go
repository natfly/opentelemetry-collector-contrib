// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oracledbreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver"
import (
	"context"

	"go.opentelemetry.io/collector/component"
)

type oracledbreceiver struct {
}

func (o oracledbreceiver) Start(_ context.Context, host component.Host) error {
	return nil
}

func (o oracledbreceiver) Shutdown(_ context.Context) error {
	return nil
}
