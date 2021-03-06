package main

import (
	"testing"
)

func TestUpdateNode(t *testing.T) {
	master := "http://127.0.0.1:8080"
	var kubeConfig string
	clientset, err := CreateClientset(&master, &kubeConfig)
	if err != nil {
		t.Fatalf("create clientset failed: %v\n", err)
	}
	ers := &ExtendedResourceScheduler{
		Clientset: clientset,
	}
	node, err := ers.FindNode("127.0.0.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	extendedResources := []string{"er1", "er2", "er3", "er4", "er5", "er6"}
	node.Status.ExtendedResourceAllocatable = append(node.Status.ExtendedResourceAllocatable, extendedResources...)
	err = ers.updateNodeStatus(node)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestUpdateNodeER7(t *testing.T) {
	master := "http://127.0.0.1:8080"
	var kubeConfig string
	clientset, err := CreateClientset(&master, &kubeConfig)
	if err != nil {
		t.Fatalf("create clientset failed: %v\n", err)
	}
	ers := &ExtendedResourceScheduler{
		Clientset: clientset,
	}
	node, err := ers.FindNode("127.0.0.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	extendedResources := []string{"er7"}
	node.Status.ExtendedResourceAllocatable = append(node.Status.ExtendedResourceAllocatable, extendedResources...)
	err = ers.updateNodeStatus(node)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestCleanNodeER(t *testing.T) {
	master := "http://127.0.0.1:8080"
	var kubeConfig string
	clientset, err := CreateClientset(&master, &kubeConfig)
	if err != nil {
		t.Fatalf("create clientset failed: %v\n", err)
	}
	ers := &ExtendedResourceScheduler{
		Clientset: clientset,
	}
	node, err := ers.FindNode("127.0.0.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	node.Status.ExtendedResourceAllocatable = []string{}
	err = ers.updateNodeStatus(node)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestPrintNodeAllocatable(t *testing.T) {
	master := "http://127.0.0.1:8080"
	var kubeConfig string
	clientset, err := CreateClientset(&master, &kubeConfig)
	if err != nil {
		t.Fatalf("create clientset failed: %v\n", err)
	}
	ers := &ExtendedResourceScheduler{
		Clientset: clientset,
	}
	node, err := ers.FindNode("127.0.0.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	t.Errorf("node: %v", node.Status.ExtendedResourceAllocatable)
}
