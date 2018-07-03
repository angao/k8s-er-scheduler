package main

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

func getExtendedResourceClaim(clientset *kubernetes.Clientset, ercName string) (*v1alpha1.ExtendedResourceClaim, error) {
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

func updateExtendedResourceClaim(clientset *kubernetes.Clientset, erc *v1alpha1.ExtendedResourceClaim) error {
	_, err := clientset.ExtensionsV1alpha1().ExtendedResourceClaims("").Update(erc)
	return err
}

func getERByRawResourceName(clientset *kubernetes.Clientset, rawResourceName string) (*v1alpha1.ExtendedResourceList, error) {
	erList, err := clientset.ExtensionsV1alpha1().ExtendedResources().List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("not found extendedresource by rawresourcename: %v", err)
		return nil, err
	}
	var extendedResources *v1alpha1.ExtendedResourceList
	extendedResources.TypeMeta = erList.TypeMeta
	extendedResources.ListMeta = erList.ListMeta
	extendedResources.Items = make([]v1alpha1.ExtendedResource, 0, len(erList.Items))

	for _, er := range erList.Items {
		if er.Spec.RawResourceName == rawResourceName {
			extendedResources.Items = append(extendedResources.Items, er)
		}
	}
	return extendedResources, nil
}

func getERByName(clientset *kubernetes.Clientset, erName string) (*v1alpha1.ExtendedResource, error) {
	er, err := clientset.ExtensionsV1alpha1().ExtendedResources().Get(erName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("not found extendedresource by ername: %v", err)
		return nil, err
	}
	return er, nil
}

func updateExtendedResource(clientset *kubernetes.Clientset, er *v1alpha1.ExtendedResource) error {
	_, err := clientset.ExtensionsV1alpha1().ExtendedResources().Update(er)
	return err
}

// nodeMatchesNodeSelectorTerms checks if a node's labels satisfy a list of node selector terms,
// terms are ORed, and an empty list of terms will match nothing.
func nodeMatchesNodeSelectorTerms(node *v1.Node, nodeSelectorTerms []v1.NodeSelectorTerm) bool {
	for _, req := range nodeSelectorTerms {
		nodeSelector, err := nodeSelectorRequirementsAsSelector(req.MatchExpressions)
		if err != nil {
			glog.V(10).Infof("Failed to parse MatchExpressions: %+v, regarding as not match.", req.MatchExpressions)
			return false
		}
		if nodeSelector.Matches(labels.Set(node.Labels)) {
			return true
		}
	}
	return false
}

// nodeSelectorRequirementsAsSelector converts the []NodeSelectorRequirement api type into a struct that implements
// labels.Selector.
func nodeSelectorRequirementsAsSelector(nsm []v1.NodeSelectorRequirement) (labels.Selector, error) {
	if len(nsm) == 0 {
		return labels.Nothing(), nil
	}
	selector := labels.NewSelector()
	for _, expr := range nsm {
		var op selection.Operator
		switch expr.Operator {
		case v1.NodeSelectorOpIn:
			op = selection.In
		case v1.NodeSelectorOpNotIn:
			op = selection.NotIn
		case v1.NodeSelectorOpExists:
			op = selection.Exists
		case v1.NodeSelectorOpDoesNotExist:
			op = selection.DoesNotExist
		case v1.NodeSelectorOpGt:
			op = selection.GreaterThan
		case v1.NodeSelectorOpLt:
			op = selection.LessThan
		default:
			return nil, fmt.Errorf("%q is not a valid node selector operator", expr.Operator)
		}
		r, err := labels.NewRequirement(expr.Key, op, expr.Values)
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}
