package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
)

const (
	addr = ":8089"
)

var mux map[string]func(http.ResponseWriter, *http.Request)

// SchedulerHandler implements custom handler
type SchedulerHandler struct{}

func (*SchedulerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.Path]; ok {
		h(w, r)
	}
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Your request URL is not found")
}

func main() {
	var masterURL, kubeConfig *string
	if home := homeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	masterURL = flag.String("masterURL", "http://127.0.0.1:8080", "kubernetes server address")
	flag.Parse()

	clientset, err := CreateClientset(*masterURL, *kubeConfig)
	if err != nil {
		glog.Fatalf("clientset error: %v", err)
	}

	extendedResourceScheduler := &ExtendedResourceScheduler{
		Clientset: clientset,
	}

	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = welcome
	mux["/predicates/ers"] = extendedResourceScheduler.Predicates

	server := http.Server{
		Addr:         addr,
		Handler:      &SchedulerHandler{},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	glog.V(2).Info("scheduler server is starting")

	if err := server.ListenAndServe(); err != nil {
		glog.Fatalf("scheduler server start failed: %v", err)
	}
}

func welcome(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "This is k8s extended resource scheduler")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
