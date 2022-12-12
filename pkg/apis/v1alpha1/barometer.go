/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BarometerSpec describes the desired state of the Barometer
type BarometerSpec struct {
	// Selector matches to pods monitored by this barometer
	Selector map[string]string `json:"selector,omitempty"`
	// Threshold to trigger pressure warnings
	Threshold *PSI `json:"threshold"`
}

type ThresholdSpec struct {
	CPU    int `json:"cpu,omitempty"`
	Memory int `json:"memory,omitempty"`
	IO     int `json:"io,omitempty"`
}

// BarometerSpec describes the observed status of the Barometer
type BarometerStatus struct {
	// Pressurized containers that have breached the threshold
	Pressure map[string]*PSI `json:"pressure,omitempty"`
}

type PSI struct {
	// Timestamp apis.VolatileTime `json:"timestamp"`
	CPU    *StallMetrics `json:"cpu,omitempty"`
	Memory *StallMetrics `json:"memory,omitempty"`
	IO     *StallMetrics `json:"io,omitempty"`
}

type StallMetrics struct {
	Some *StallMetric `json:"some,omitempty"`
	Full *StallMetric `json:"full,omitempty"`
}

type StallMetric struct {
	Avg10  *int32 `json:"avg10,omitempty"`
	Avg60  *int32 `json:"avg60,omitempty"`
	Avg300 *int32 `json:"avg300,omitempty"`
}

// Barometer is the Schema for the Barometers API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=barometers,categories=karpenter
// +kubebuilder:subresource:status
type Barometer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BarometerSpec   `json:"spec,omitempty"`
	Status BarometerStatus `json:"status,omitempty"`
}

// BarometerList contains a list of Provisioner
// +kubebuilder:object:root=true
type BarometerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Barometer `json:"items"`
}
