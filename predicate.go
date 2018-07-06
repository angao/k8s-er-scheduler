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

// Predicates implemented filter functions.
// The filter list is expected to be a subset of the supplied list.
func Predicates(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			extenedResourceScheduler := &ExtendedResourceScheduler{
				Clientset: clientset,
			}
			extenderFilterResult = filter(extenderArgs, extenedResourceScheduler)
		}

		w.Header().Set("Content-Type", "application/json")
		if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func filter(extenderArgs schedulerapi.ExtenderArgs, extenedResourceScheduler *ExtendedResourceScheduler) *schedulerapi.ExtenderFilterResult {
	pod := extenderArgs.Pod
	nodes := extenderArgs.Nodes.Items

	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	// default all node scheduling failed
	defaultNotSchedule := defaultFailedNodes(nodes)

	result := &schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: defaultNotSchedule,
		Error:       "",
	}

	ercs, err := findExtendedResourceClaims(pod, extenedResourceScheduler)
	if err != nil {
		glog.Errorf("find extendedresourceclaims failed: %v", err)
		result.Error = err.Error()
		return result
	}
	for _, erc := range ercs {
		erList, err := findExtendedResourceList(erc, extenedResourceScheduler)
		if err != nil {
			glog.Errorf("erc: %v error: %v", erc, err)
			result.Error = err.Error()
			return result
		}

		extendedResources := &v1alpha1.ExtendedResourceList{
			Items: make([]v1alpha1.ExtendedResource, 0),
		}
		for _, er := range erList.Items {
			nodeAffinity := er.Spec.NodeAffinity
			requirements := erc.Spec.MetadataRequirements

			if nodeAffinity == nil ||
				!mapInMap(requirements.MatchLabels, er.Spec.Properties) ||
				!labelMatchesLabelSelectorExpressions(requirements.MatchExpressions, er.Spec.Properties) {
				continue
			}

			nodeSelectorTerms := nodeAffinity.Required.NodeSelectorTerms
			for _, node := range nodes {
				// set er related with erc
				if nodeMatchesNodeSelectorTerms(&node, nodeSelectorTerms) {
					erc.Spec.ExtendedResourceNames = append(erc.Spec.ExtendedResourceNames, er.Name)
					er.Spec.ExtendedResourceClaimName = erc.Name
					extendedResources.Items = append(extendedResources.Items, er)
					canSchedule = append(canSchedule, node)
				}
				canNotSchedule[node.ObjectMeta.Name] = "node's label is not satisfy er's nodeAffinity"
			}
		}

		extendedResourceNum := erc.Spec.ExtendedResourceNum

		extenedResourceScheduler.UpdateExtendedResourceClaim(pod.Namespace, erc)
		for _, er := range extendedResources.Items {
			if er.Spec.ExtendedResourceClaimName != "" {
				extenedResourceScheduler.UpdateExtendedResource(&er)
			}
		}
		canSchedule = canSchedule[:extendedResourceNum]
		result.FailedNodes = canNotSchedule
	}

	result.Nodes.Items = canSchedule
	return result
}

func findExtendedResourceClaims(pod v1.Pod, extenedResourceScheduler *ExtendedResourceScheduler) ([]*v1alpha1.ExtendedResourceClaim, error) {
	extendedResourceClaimNames := make([]string, 0)
	for _, container := range pod.Spec.Containers {
		if len(container.ExtendedResourceClaims) != 0 {
			extendedResourceClaimNames = append(extendedResourceClaimNames, container.ExtendedResourceClaims...)
		}
	}
	if len(extendedResourceClaimNames) == 0 {
		return nil, errors.New("extendedresourceclaims not set")
	}
	extendedResourceClaims := make([]*v1alpha1.ExtendedResourceClaim, 0)
	for _, ercName := range extendedResourceClaimNames {
		erc, err := extenedResourceScheduler.FindExtendedResourceClaim(pod.Namespace, ercName)
		if err != nil {
			return nil, err
		}
		extendedResourceClaims = append(extendedResourceClaims, erc)
	}
	return extendedResourceClaims, nil
}

func findExtendedResourceList(erc *v1alpha1.ExtendedResourceClaim, extenedResourceScheduler *ExtendedResourceScheduler) (*v1alpha1.ExtendedResourceList, error) {
	rawResourceName := erc.Spec.RawResourceName
	erList, err := extenedResourceScheduler.FindERByRawResourceName(rawResourceName)
	if err != nil {
		return nil, err
	}
	extendedResourceNames := erc.Spec.ExtendedResourceNames
	if len(extendedResourceNames) != 0 {
		extendedResources := make([]v1alpha1.ExtendedResource, 0, len(extendedResourceNames))
		for _, name := range extendedResourceNames {
			er, err := extenedResourceScheduler.FindERByName(name)
			if err != nil {
				return nil, err
			}
			extendedResources = append(extendedResources, *er)
		}
		erList.Items = append(erList.Items, extendedResources...)
	}
	return erList, nil
}

// default set all node is fail
func defaultFailedNodes(nodes []v1.Node) map[string]string {
	canNotSchedule := make(map[string]string)
	for _, node := range nodes {
		canNotSchedule[node.ObjectMeta.Name] = ""
	}
	return canNotSchedule
}
