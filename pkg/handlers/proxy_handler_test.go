package handlers

import (
	"testing"
)

func TestValidateProxyRequest(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		podName   string // parameter name "name"
		path      string
		wantErr   bool
	}{
		{
			name:      "valid request",
			namespace: "default",
			podName:   "my-pod",
			path:      "/api/v1/health",
			wantErr:   false,
		},
		{
			name:      "invalid namespace with slash",
			namespace: "def/ault",
			podName:   "my-pod",
			path:      "/api/v1/health",
			wantErr:   true,
		},
		{
			name:      "invalid namespace with ..",
			namespace: "..",
			podName:   "my-pod",
			path:      "/api/v1/health",
			wantErr:   true,
		},
		{
			name:      "invalid pod name with slash",
			namespace: "default",
			podName:   "my/pod",
			path:      "/api/v1/health",
			wantErr:   true,
		},
		{
			name:      "invalid pod name with ..",
			namespace: "default",
			podName:   "../secrets",
			path:      "/api/v1/health",
			wantErr:   true,
		},
		{
			name:      "invalid path with ..",
			namespace: "default",
			podName:   "my-pod",
			path:      "/../secrets",
			wantErr:   true,
		},
		{
			name:      "invalid path with embedded ..",
			namespace: "default",
			podName:   "my-pod",
			path:      "/app/../config",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateProxyRequest(tt.namespace, tt.podName, tt.path); (err != nil) != tt.wantErr {
				t.Errorf("ValidateProxyRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
