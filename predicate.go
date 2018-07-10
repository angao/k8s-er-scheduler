package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
			extendedResourceScheduler := &ExtendedResourceScheduler{
				Clientset: clientset,
			}
			extenderFilterResult = filter(extenderArgs, extendedResourceScheduler)
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

func filter(extenderArgs schedulerapi.ExtenderArgs, extendedResourceScheduler *ExtendedResourceScheduler) *schedulerapi.ExtenderFilterResult {
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

	extendedResourceClaims, err := extendedResourceScheduler.FindExtendedResourceClaimList(pod)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// calculate how much extendedResource are needed for pod
	// TODO: Check whether the user's declared rawResourceName is the same as the declared rawResourceName of extended resource
	var extendedResourceNames = make([]string, 0)
	for _, erc := range extendedResourceClaims {
		extendedResourceNames = append(extendedResourceNames, erc.Spec.ExtendedResourceNames...)
	}

	glog.V(2).Info("start to filter node")

	// TODO: check the extended resources of the node asynchronously
	for _, node := range nodes {
		nodeName := node.Name
		extendedResourceAllocatable := node.Status.ExtendedResourceAllocatable
		if len(extendedResourceAllocatable) < len(extendedResourceNames) {
			canNotSchedule[nodeName] = "extended resources that can be allocated on this node are less than pod needs"
			continue
		}

		if ss, b := sliceInSlice(extendedResourceNames, extendedResourceAllocatable); !b {
			canNotSchedule[nodeName] = fmt.Sprintf("there are no such [%s] extended resource", strings.Join(ss, " "))
			continue
		}

		extendedResources, err := extendedResourceScheduler.FindExtendedResourceList(extendedResourceAllocatable)
		if err != nil {
			canNotSchedule[nodeName] = err.Error()
			continue
		}

		// filter out the er specified in erc and er status is not available
		scheduled := false
		extendedResourceAvailable := make([]*v1alpha1.ExtendedResource, 0)
		extendedResourceAvailable = append(extendedResourceAvailable, extendedResources...)
		if len(extendedResourceNames) > 0 {
		loop:
			for i := 0; i < len(extendedResourceAvailable); i++ {
				for _, name := range extendedResourceNames {
					if name != extendedResourceAvailable[i].Name {
						continue
					}
					if extendedResourceAvailable[i].Status.Phase != v1alpha1.ExtendedResourceAvailable {
						scheduled = true
						break loop
					}
					extendedResourceAvailable = append(extendedResourceAvailable[:i], extendedResourceAvailable[i+1:]...)
				}
			}
			if scheduled {
				canNotSchedule[nodeName] = "there are unavailable extended resources in extendedresourceclaim"
				continue
			}
		}

		for _, erc := range extendedResourceClaims {
			erNames := erc.Spec.ExtendedResourceNames
			erNum := erc.Spec.ExtendedResourceNum
			requirements := erc.Spec.MetadataRequirements
			rawResourceName := erc.Spec.RawResourceName

			for i := 0; i < len(extendedResourceAvailable); i++ {
				er := extendedResourceAvailable[i]
				prop := er.Spec.Properties
				if int64(len(erNames)) < erNum && rawResourceName == er.Spec.RawResourceName &&
					(mapInMap(requirements.MatchLabels, prop) ||
						labelMatchesLabelSelectorExpressions(requirements.MatchExpressions, prop)) {
					erNames = append(erNames, er.Name)
					er.Spec.ExtendedResourceClaimName = erc.Name
					extendedResourceAvailable = append(extendedResourceAvailable[:i], extendedResourceAvailable[i+1:]...)
				}
			}
			if erNum != 0 && int64(len(erNames)) < erNum {
				canNotSchedule[nodeName] = fmt.Sprintf("extended resource that can be allocated are not satisfy [%s] needs", erc.Name)
				continue
			}
			erc.Spec.ExtendedResourceNames = erNames
			erc.Status.Phase = v1alpha1.ExtendedResourceClaimPending
			erc.Status.Reason = "extended resource have been satisfied and waiting to bound"
		}

		for _, erc := range extendedResourceClaims {
			if erc.Status.Phase != v1alpha1.ExtendedResourceClaimPending {
				scheduled = true
				break
			}
		}
		if !scheduled {
			canSchedule = append(canSchedule, node)
		} else {
			canNotSchedule[nodeName] = "node can allocate extended resource are not satisfy pod needs"
			continue
		}
	}

	// pod can be scheduled, so erc need to update, erc is pending
	if len(canSchedule) > 0 {
		for _, erc := range extendedResourceClaims {
			extendedResourceScheduler.UpdateExtendedResourceClaim(pod.Namespace, erc)
		}
	}
	result.FailedNodes = canNotSchedule
	result.Nodes.Items = canSchedule
	return result
}

// default set all node is fail
func defaultFailedNodes(nodes []v1.Node) map[string]string {
	canNotSchedule := make(map[string]string)
	for _, node := range nodes {
		canNotSchedule[node.ObjectMeta.Name] = ""
	}
	return canNotSchedule
}
