package kor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func getDeploymentsWithoutReplicas(kubeClient *kubernetes.Clientset, namespace string) ([]string, error) {
	deploymentsList, err := kubeClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deploymentsWithoutReplicas []string

	for _, deployment := range deploymentsList.Items {
		if deployment.Labels["kor/used"] == "true" {
			continue
		}

		if *deployment.Spec.Replicas == 0 {
			deploymentsWithoutReplicas = append(deploymentsWithoutReplicas, deployment.Name)
		}
	}

	return deploymentsWithoutReplicas, nil
}

func ProcessNamespaceDeployments(clientset *kubernetes.Clientset, namespace string) ([]string, error) {
	usedServices, err := getDeploymentsWithoutReplicas(clientset, namespace)
	if err != nil {
		return nil, err
	}

	return usedServices, nil
}

func GetUnusedDeployments(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Deployments")
		fmt.Println(output)
		fmt.Println()
	}
}

func GetUnusedDeploymentsSendToSlackWebhook(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, slackWebhookURL string) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Deployments")

		outputBuffer.WriteString(output)
		outputBuffer.WriteString("\n")
	}

	if err := SendToSlackWebhook(slackWebhookURL, outputBuffer.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send payload to Slack: %v\n", err)
	}
}

func GetUnusedDeploymentsSendToSlackAsFile(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, slackChannel string, slackAuthToken string) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Deployments")

		outputBuffer.WriteString(output)
		outputBuffer.WriteString("\n")
	}

	outputFilePath, _ := writeOutputToFile(outputBuffer)

	if err := SendFileToSlack(outputFilePath, "Unused Deployments", slackChannel, slackAuthToken); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send output to Slack: %v\n", err)
	}
}

func GetUnusedDeploymentsStructured(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, outputFormat string) (string, error) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)
	response := make(map[string]map[string][]string)

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		resourceMap := make(map[string][]string)
		resourceMap["Deployments"] = diff
		response[namespace] = resourceMap
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}

	if outputFormat == "yaml" {
		yamlResponse, err := yaml.JSONToYAML(jsonResponse)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		return string(yamlResponse), nil
	} else {
		return string(jsonResponse), nil
	}
}
