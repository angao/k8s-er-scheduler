/*
Copyright 2016 The Kubernetes Authors.

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

// This file contains a collection of methods that can be used from go-restful to
// generate Swagger API documentation for its models. Please read this PR for more
// information on the implementation: https://github.com/emicklei/go-restful/pull/215
//
// TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
// they are on one line! For multiple line or blocks that you want to ignore use ---.
// Any context after a --- is ignored.
//
// Those methods can be generated by using hack/update-generated-swagger-docs.sh

// AUTO-GENERATED FUNCTIONS START HERE
var map_ExtendedResource = map[string]string{
	"":         "ExtendedResource represents a specific extended resource",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
	"spec":     "Spec is the desired state of the ExtendedResource.",
	"status":   "Status is the current state of the ExtendedResource. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status",
}

func (ExtendedResource) SwaggerDoc() map[string]string {
	return map_ExtendedResource
}

var map_ExtendedResourceClaim = map[string]string{
	"":         "ExtendedResourceClaim is used by users to ask for extended resources",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
	"spec":     "Spec is the desired state of the ExtendedResourceClaim.",
	"status":   "Status is the current state of the ExtendedResourceClaim. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status",
}

func (ExtendedResourceClaim) SwaggerDoc() map[string]string {
	return map_ExtendedResourceClaim
}

var map_ExtendedResourceClaimList = map[string]string{
	"":         "ExtendedResourceClaimList is a collection of ExtendedResourceClaim.",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
	"items":    "Items is the list of ExtendedResourceClaim.",
}

func (ExtendedResourceClaimList) SwaggerDoc() map[string]string {
	return map_ExtendedResourceClaimList
}

var map_ExtendedResourceClaimSpec = map[string]string{
	"": "ExtendedResourceClaimSpec describes the ExtendedResourceClaim the user wishes to exist.",
	"metadataRequirements":  "defines general resource property matching constraints. e.g.: zone in { us-west1-b, us-west1-c }; type: k80",
	"extendedResourceNames": "ExtendedResourceNames are the names of ExtendedResources",
	"rawResourceName":       "raw extended resource name, such as nvidia.com/gpu used for batch resources request",
	"extendResourceNum":     "number of extended resource, for example: request 8 nvidia.com/gpu at one time",
}

func (ExtendedResourceClaimSpec) SwaggerDoc() map[string]string {
	return map_ExtendedResourceClaimSpec
}

var map_ExtendedResourceClaimStatus = map[string]string{
	"":        "ExtendedResourceClaimStatus is the status of ExtendedResourceClaim",
	"phase":   "Phase indicates if the Extended Resource Claim is Lost, bound or pending",
	"message": "A human-readable message indicating details about why CRC is in this phase",
	"reason":  "Brief string that describes any failure, used for CLI etc",
}

func (ExtendedResourceClaimStatus) SwaggerDoc() map[string]string {
	return map_ExtendedResourceClaimStatus
}

var map_ExtendedResourceList = map[string]string{
	"":         "ExtendedResourceList is a collection of ExtendedResource.",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
	"items":    "Items is the list of ExtendedResource.",
}

func (ExtendedResourceList) SwaggerDoc() map[string]string {
	return map_ExtendedResourceList
}

var map_ExtendedResourceSpec = map[string]string{
	"":                          "ExtendedResourceSpec describes the ExtendedResource the user wishes to exist.",
	"nodeAffinity":              "NodeAffinity defines constraints that limit what nodes this resource can be accessed from. This field influences the scheduling of pods that use this resource.",
	"rawResourceName":           "Raw resource name. E.g.: nvidia.com/gpu",
	"deviceID":                  "device id",
	"properties":                "gpuType: k80 zone: us-west1-b Note Kubelet adds a special property corresponding to the above ResourceName field. This will allow a single ResourceClass (e.g., “gpu”) to match multiple types of resources (e.g., nvidia.com/gpu and amd.com/gpu) through general set selector.",
	"extendedResourceClaimName": "ExtendedResourceClaimName is the name of ExtendedResourceClaim that the ExtendedResource is bound to",
}

func (ExtendedResourceSpec) SwaggerDoc() map[string]string {
	return map_ExtendedResourceSpec
}

var map_ExtendedResourceStatus = map[string]string{
	"":            "ExtendedResourceStatus is the status of ExtendedResource",
	"capacity":    "Capacity is the capacity of this device",
	"allocatable": "Allocatable is the resource of this device that can be available for scheduling",
	"phase":       "Phase indicates if the extended Resource is available, bound or pending",
	"message":     "A human-readable message indicating details about why CR is in this phase",
	"reason":      "Brief string that describes any failure, used for CLI etc",
}

func (ExtendedResourceStatus) SwaggerDoc() map[string]string {
	return map_ExtendedResourceStatus
}

var map_ResourceNodeAffinity = map[string]string{
	"":         "ResourceNodeAffinity defines constraints that limit what nodes this extended resource can be accessed from.",
	"required": "Required specifies hard node constraints that must be met.",
}

func (ResourceNodeAffinity) SwaggerDoc() map[string]string {
	return map_ResourceNodeAffinity
}

// AUTO-GENERATED FUNCTIONS END HERE
