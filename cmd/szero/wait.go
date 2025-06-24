package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/jadolg/szero/pkg"
	"k8s.io/client-go/kubernetes"
)

func waitForResourcesOrFatal(ctx context.Context, clientset kubernetes.Interface, downscaled bool) {
	waitFor := len(namespaces) * 3 // deployments, statefulsets, and daemonsets per namespace
	errors := make(chan error, 1)
	done := make(chan bool, waitFor)

	log.Infof("Waiting for all resources to reach the desired state in %d namespaces (timeout %v)", len(namespaces), timeout)

	for _, namespace := range namespaces {
		if !skipDeployments {
			go func(errors chan error) {
				waitForDeployments(ctx, clientset, namespace, downscaled, done, errors)
			}(errors)
		} else {
			waitFor--
		}

		if !skipStatefulsets {
			go func(errors chan error) {
				waitForStatefulSets(ctx, clientset, namespace, downscaled, done, errors)
			}(errors)
		} else {
			waitFor--
		}

		if !skipDaemonsets {
			go func(errors chan error) {
				waitForDaemonSets(ctx, clientset, namespace, downscaled, done, errors)
			}(errors)
		} else {
			waitFor--
		}
	}

	for {
		select {
		case err := <-errors:
			if err != nil {
				log.Fatalf("Error waiting for resources to reach desired state: %v", err)
			}
		case <-done:
			waitFor--
			if waitFor == 0 {
				return
			}
		}
	}
}

func waitForDaemonSets(ctx context.Context, clientset kubernetes.Interface, namespace string, downscaled bool, done chan bool, errors chan error) {
	daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
	if err != nil {
		errors <- err
		return
	}
	err = pkg.WaitForDaemonSets(ctx, clientset, daemonsets, timeout, downscaled)
	if err != nil {
		errors <- fmt.Errorf("could not wait for DaemonSets in namespace %s: %w", namespace, err)
		return
	}
	done <- true
}

func waitForStatefulSets(ctx context.Context, clientset kubernetes.Interface, namespace string, downscaled bool, done chan bool, errors chan error) {
	statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
	if err != nil {
		errors <- err
		return
	}
	err = pkg.WaitForStatefulSets(ctx, clientset, statefulsets, timeout, downscaled)
	if err != nil {
		errors <- fmt.Errorf("could not wait for StatefulSets in namespace %s: %w", namespace, err)
		return
	}
	done <- true
}

func waitForDeployments(ctx context.Context, clientset kubernetes.Interface, namespace string, downscaled bool, done chan bool, errors chan error) {
	deployments, err := pkg.GetDeployments(ctx, clientset, namespace)
	if err != nil {
		errors <- err
		return
	}
	err = pkg.WaitForDeployments(ctx, clientset, deployments, timeout, downscaled)
	if err != nil {
		errors <- fmt.Errorf("could not wait for Deployments in namespace %s: %w", namespace, err)
		return
	}
	done <- true
}
