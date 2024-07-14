package v1

import (
	"fmt"
	"testing"
)

func TestGetAerospikeInitContainerImageWithoutSpecifyingCustomInitContainer(t *testing.T) {
	aeroCluster := AerospikeCluster{}
	aeroCluster.Spec = AerospikeClusterSpec{}
	aeroCluster.Spec.PodSpec = AerospikePodSpec{}

	initContainerImage := GetAerospikeInitContainerImage(&aeroCluster)
	expectedImage := fmt.Sprintf("%s/%s/%s", AerospikeInitContainerDefaultRegistry, AerospikeInitContainerDefaultRegistryNamespace, AerospikeInitContainerDefaultRepoAndTag)
	if initContainerImage != expectedImage {
		t.Fatalf("By GetAerospikeInitContainerImage should have returned %s but got %s", expectedImage, initContainerImage)
	}
}

func TestGetAerospikeInitContainerImageWithCustomRegistryForInitContainer(t *testing.T) {
	aeroCluster := AerospikeCluster{}
	aeroCluster.Spec = AerospikeClusterSpec{}
	aeroCluster.Spec.PodSpec = AerospikePodSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec = &AerospikeInitContainerSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec.ImageRegistry = "myregistry.net"

	initContainerImage := GetAerospikeInitContainerImage(&aeroCluster)
	expectedImage := fmt.Sprintf("myregistry.net/%s/%s", AerospikeInitContainerDefaultRegistryNamespace, AerospikeInitContainerDefaultRepoAndTag)
	if initContainerImage != expectedImage {
		t.Fatalf("GetAerospikeInitContainerImage should have returned %s but got %s", expectedImage, initContainerImage)
	}
}

func TestGetAerospikeInitContainerImageWithCustomImageForInitContainer(t *testing.T) {
	aeroCluster := AerospikeCluster{}
	aeroCluster.Spec = AerospikeClusterSpec{}
	aeroCluster.Spec.PodSpec = AerospikePodSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec = &AerospikeInitContainerSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec.Image = "myinitimage:x.y.z"

	initContainerImage := GetAerospikeInitContainerImage(&aeroCluster)
	expectedImage := fmt.Sprintf("%s/myinitimage:x.y.z", AerospikeInitContainerDefaultRegistry)
	if initContainerImage != expectedImage {
		t.Fatalf("GetAerospikeInitContainerImage should have returned %s but got %s", expectedImage, initContainerImage)
	}
}

func TestGetAerospikeInitContainerImageWithCustomRegistryAndCustomImageForInitContainer(t *testing.T) {
	aeroCluster := AerospikeCluster{}
	aeroCluster.Spec = AerospikeClusterSpec{}
	aeroCluster.Spec.PodSpec = AerospikePodSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec = &AerospikeInitContainerSpec{}
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec.ImageRegistry = "myregistry.net"
	aeroCluster.Spec.PodSpec.AerospikeInitContainerSpec.Image = "myinitimage:x.y.z"

	initContainerImage := GetAerospikeInitContainerImage(&aeroCluster)
	expectedImage := "myregistry.net/myinitimage:x.y.z"
	if initContainerImage != expectedImage {
		t.Fatalf("GetAerospikeInitContainerImage should have returned %s but got %s", expectedImage, initContainerImage)
	}
}
