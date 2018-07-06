package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1alpha1"
	"k8s.io/client-go/kubernetes"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api/v1"
)

// ExtendedResourceScheduler handle extendedresource scheduler
type ExtendedResourceScheduler struct {
	Clientset *kubernetes.Clientset
}

// Predicates implemented filter functions.
// The filter list is expected to be a subset of the supplied list.
func (ers *ExtendedResourceScheduler) Predicates(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)
	glog.V(2).Infof("body: %s", buf.String())

	var extenderArgs schedulerapi.ExtenderArgs
	var extenderFilterResult *schedulerapi.ExtenderFilterResult

	if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes:       nil,
			FailedNodes: nil,
			Error:       err.Error(),
		}
	} else {
		extenderFilterResult = filter(extenderArgs, ers.Clientset)
	}

	if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

func filter(extenderArgs schedulerapi.ExtenderArgs, clientset *kubernetes.Clientset) *schedulerapi.ExtenderFilterResult {
	pod := extenderArgs.Pod
	nodes := extenderArgs.Nodes.Items

	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	// default all node scheduling failed
	defaultNotSchedule := defaultFailedNodes(nodes)

	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: defaultNotSchedule,
		Error:       "",
	}

	extendedResourceClaims := make([]string, 0)
	for _, container := range pod.Spec.Containers {
		if len(container.ExtendedResourceClaims) != 0 {
			extendedResourceClaims = append(extendedResourceClaims, container.ExtendedResourceClaims...)
		}
	}
	if len(extendedResourceClaims) == 0 {
		result.Error = "extendedResourceClaims is empty"
		return &result
	}

	// Find ExtendedResourceClaim by extendedresourceclaim's name
	for _, ercName := range extendedResourceClaims {
		erc, err := getExtendedResourceClaim(clientset, pod.ObjectMeta.Namespace, ercName)
		if err != nil {
			result.Error = err.Error()
			return &result
		}
		rawResourceName := erc.Spec.RawResourceName
		erList, err := getERByRawResourceName(clientset, rawResourceName)
		if err != nil {
			result.Error = err.Error()
			return &result
		}

		extendedResourceNames := erc.Spec.ExtendedResourceNames
		if len(extendedResourceNames) != 0 {
			extendedResources := make([]v1alpha1.ExtendedResource, 0, len(extendedResourceNames))
			for _, name := range extendedResourceNames {
				er, err := getERByName(clientset, name)
				if err != nil {
					result.Error = err.Error()
					return &result
				}
				extendedResources = append(extendedResources, *er)
			}
			if err := checkExtendedResourceNodeAffinity(extendedResources); err != nil {
				result.Error = err.Error()
				return &result
			}
			erList.Items = append(erList.Items, extendedResources...)
		}

		extendedResourceNum := erc.Spec.ExtendedResourceNum
		// the number of nodes does not satisfy the extendedresourcenum required for pod
		if int64(len(nodes)) < extendedResourceNum {
			result.Error = "the number of nodes does not satisfy the extendedresourcenum required for pod"
			return &result
		}

		for _, er := range erList.Items {
			nodeAffinity := er.Spec.NodeAffinity
			if nodeAffinity == nil {
				result.Error = "extendedresource nodeAffinity is empty"
				return &result
			}
			nodeSelectorTerms := nodeAffinity.Required.NodeSelectorTerms
			for _, node := range nodes {
				// set er related with erc
				if nodeMatchesNodeSelectorTerms(&node, nodeSelectorTerms) {
					canSchedule = append(canSchedule, node)
					extendedResourceNames = append(extendedResourceNames, er.Name)
					er.Spec.ExtendedResourceClaimName = erc.Name
				}
				canNotSchedule[node.ObjectMeta.Name] = "node's label is not satisfy er's nodeAffinity"
			}
		}

		if int64(len(canSchedule)) < extendedResourceNum {
			result.Error = "The extendedresource that can be scheduled are insufficient"
			return &result
		}

		erc.Spec.ExtendedResourceNames = extendedResourceNames
		updateExtendedResourceClaim(clientset, pod.ObjectMeta.Namespace, erc)
		for _, er := range erList.Items {
			if er.Spec.ExtendedResourceClaimName != "" {
				updateExtendedResource(clientset, &er)
			}
		}
		canSchedule = canSchedule[:extendedResourceNum]
		result.FailedNodes = canNotSchedule
	}

	result.Nodes.Items = canSchedule
	return &result
}

// default set all node is fail
func defaultFailedNodes(nodes []v1.Node) map[string]string {
	canNotSchedule := make(map[string]string)
	for _, node := range nodes {
		canNotSchedule[node.ObjectMeta.Name] = ""
	}
	return canNotSchedule
}

// check whether the extendedresource is on the same node in extendedresourcenaems
func checkExtendedResourceNodeAffinity(extendedResources []v1alpha1.ExtendedResource) error {
	if len(extendedResources) == 0 || len(extendedResources) == 1 {
		return nil
	}
	extendedResource := extendedResources[0]
	for _, er := range extendedResources {
		if er.Spec.NodeAffinity != extendedResource.Spec.NodeAffinity {
			return errors.New("there are two cases in which the nodeAffinity is different in er")
		}
	}
	return nil
}
