package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
)

const (
	addr = ":8080"
)

func main() {
	var kubeConfig *string
	if home := homeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	clientset, err := CreateClientset(kubeConfig)
	if err != nil {
		glog.Fatalf("clientset error: %v", err)
	}
	glog.V(2).Info("create clientset success")

	server := http.Server{
		Addr: addr,
		Handler: &ExtendedResourceScheduler{
			Clientset: clientset,
		},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	glog.V(2).Info("scheduler server is starting")

	if err := server.ListenAndServe(); err != nil {
		glog.Fatalf("scheduler server start failed: %v", err)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
