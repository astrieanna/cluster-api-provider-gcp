/*
Copyright 2018 The Kubernetes Authors.

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

package google_test

import (
	"testing"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"sigs.k8s.io/cluster-api-provider-gcp/pkg/cloud/google"
	"sigs.k8s.io/cluster-api/pkg/controller/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func aTestDelete(t *testing.T) {
	testCases := []struct {
		name                    string
		firewallsDeleteOpResult *compute.Operation
		firewallsDeleteErr      error
		expectedErrorMessage    string
	}{
		{"successs", &compute.Operation{}, nil, ""},
		{"error", nil, &googleapi.Error{Code: 408, Message: "request timeout"}, "error deleting firewall rule for internal cluster traffic: error deleting firewall rule: googleapi: Error 408: request timeout"},
		{"404/NotFound error should succeed", nil, &googleapi.Error{Code: 404, Message: "not found"}, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			computeServiceMock := GCEClientComputeServiceMock{
				mockFirewallsDelete: func(project string, name string) (*compute.Operation, error) {
					return tc.firewallsDeleteOpResult, tc.firewallsDeleteErr
				},
			}
			params := google.ClusterActuatorParams{ComputeService: &computeServiceMock}
			actuator := newClusterActuator(t, params)
			cluster := newDefaultClusterFixture(t)
			err := actuator.Delete(cluster)
			if err != nil || tc.expectedErrorMessage != "" {
				if err == nil {
					t.Errorf("unexpected error message")
				} else if err.Error() != tc.expectedErrorMessage {
					t.Errorf("error mismatch: got '%v', want '%v'", err, tc.expectedErrorMessage)
				}
			}
		})
	}
}

func newClusterActuator(t *testing.T, params google.ClusterActuatorParams) cluster.Actuator {
	t.Helper()
	m, err := manager.New(nil, manager.Options{})
	if err != nil {
		t.Fatalf("error creating manager: %v", err)
	}
	actuator, err := google.NewClusterActuator(m, params)
	if err != nil {
		t.Fatalf("error creating cluster actuator: %v", err)
	}
	return actuator
}
