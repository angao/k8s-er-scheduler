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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	w.Header().Set("Content-Type", "application/json")
	if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

func filter(args schedulerapi.ExtenderArgs, clientset *kubernetes.Clientset) *schedulerapi.ExtenderFilterResult {
	pod := args.Pod
	canSchedule := make([]v1.Node, 0, len(args.Nodes.Items))
	canNotSchedule := make(map[string]string)
	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	extendedResourceClaims := make([]string, 0)
	for _, container := range pod.Spec.Containers {
		if len(container.ExtendedResourceClaim) != 0 {
			extendedResourceClaims = append(extendedResourceClaims, container.ExtendedResourceClaim...)
		}
	}
	if len(extendedResourceClaims) == 0 {
		result.Error = "extendedResourceClaims are empty"
		return &result
	}

	// Find ExtendedResourceClaim by extendedresourceclaim's name
	for _, ercName := range extendedResourceClaims {
		_, err := lookupExtendedResourceClaim(clientset, ercName)
		if err != nil {
			result.Error = err.Error()
			return &result
		}

	}

	return &result
}

func lookupExtendedResourceClaim(clientset *kubernetes.Clientset, ercName string) (*v1alpha1.ExtendedResourceClaim, error) {
	erc, err := clientset.ExtensionsV1alpha1().ExtendedResourceClaims("").Get(ercName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("not found extendedresourceclaim: %v", err)
		return nil, err
	}
	if len(erc.Spec.ExtendedResourceNames) == 0 && erc.Spec.ExtendedResourceNum == 0 {
		glog.Errorf("ExtendedResourceNames and ExtendedResourceNum are empty")
		return nil, errors.New("ExtendedResourceNames and ExtendedResourceNum are empty")
	}
	return erc, nil
}
