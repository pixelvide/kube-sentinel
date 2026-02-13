package utils

import (
	"testing"
)

func TestGetImageRegistryAndRepo(t *testing.T) {
	testcase := []struct {
		image    string
		registry string
		repo     string
	}{
		{"nginx", "", "library/nginx"},
		{"nginx:latest", "", "library/nginx"},
		{"pixelvide/kube-sentinel:latest", "", "pixelvide/kube-sentinel"},
		{"docker.io/library/nginx", "docker.io", "library/nginx"},
		{"docker.io/library/nginx:latest", "docker.io", "library/nginx"},
		{"gcr.io/my-project/my-image", "gcr.io", "my-project/my-image"},
		{"gcr.io/my-project/my-image:tag", "gcr.io", "my-project/my-image"},
		{"quay.io/my-org/my-repo", "quay.io", "my-org/my-repo"},
		{"quay.io/my-org/my-repo:tag", "quay.io", "my-org/my-repo"},
		{"registry.example.com/my-repo/test", "registry.example.com", "my-repo/test"},
	}
	for _, tc := range testcase {
		registry, repo := GetImageRegistryAndRepo(tc.image)
		if registry != tc.registry || repo != tc.repo {
			t.Errorf("GetImageRegistryAndRepo(%q) = (%q, %q), want (%q, %q)", tc.image, registry, repo, tc.registry, tc.repo)
		}
	}
}

func TestIsSecureRegistry(t *testing.T) {
	testcases := []struct {
		host      string
		shouldErr bool
	}{
		{"docker.io", false},
		{"quay.io", false},
		{"gcr.io", false},
		{"localhost", true},
		{"localhost:5000", true},
		{"127.0.0.1", true},
		{"127.0.0.1:5000", true},
		{"[::1]", true},
		{"[::1]:5000", true},
		{"google.com", false},
		{"0.0.0.0", true},
	}

	for _, tc := range testcases {
		err := IsSecureRegistry(tc.host)
		if tc.shouldErr && err == nil {
			t.Errorf("IsSecureRegistry(%q) should have failed but passed", tc.host)
		}
		if !tc.shouldErr && err != nil {
			t.Errorf("IsSecureRegistry(%q) failed: %v", tc.host, err)
		}
	}
}

func TestGenerateNodeAgentName(t *testing.T) {
	testcase := []struct {
		nodeName string
	}{
		{"node1"},
		{"shortname"},
		{"a-very-long-node-name-that-exceeds-the-maximum-length-allowed-for-kubernetes-names"},
		{"node-with-63-characters-abcdefghijklmnopqrstuvwxyz-123456789101"},
	}

	for _, tc := range testcase {
		podName := GenerateNodeAgentName(tc.nodeName)
		if len(podName) > 63 {
			t.Errorf("GenerateNodeAgentName(%q) = %q, length %d exceeds 63", tc.nodeName, podName, len(podName))
		}
	}
}

func TestGetUserGlabConfigDir(t *testing.T) {
	// Skip execution as it requires root permissions to write to /data
	t.Skip("Skipping execution of GetUserGlabConfigDir as it requires root permissions to write to /data")
}
