package kor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getDeploymentsWithoutReplicas(kubeClient *kubernetes.Clientset, namespace string) ([]string, error) {
	deploymentsList, err := kubeClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deploymentsWithoutReplicas []string

	for _, deployment := range deploymentsList.Items {
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

func GetUnusedDeployments(namespace string, kubeconfig string) {
	var kubeClient *kubernetes.Clientset
	var namespaces []string

	kubeClient = GetKubeClient(kubeconfig)

	namespaces = SetNamespaceList(namespace, kubeClient)

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(kubeClient, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Deployments")
		fmt.Println(output)
		fmt.Println()
	}
}

func GetUnusedDeploymentsSendToSlackWebhook(namespace string, kubeconfig string, slackWebhookURL string) {
	var kubeClient *kubernetes.Clientset
	var namespaces []string

	kubeClient = GetKubeClient(kubeconfig)

	namespaces = SetNamespaceList(namespace, kubeClient)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(kubeClient, namespace)
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

func GetUnusedDeploymentsSendToSlackAsFile(namespace string, kubeconfig string, slackChannel string, slackAuthToken string) {
	var kubeClient *kubernetes.Clientset
	var namespaces []string

	kubeClient = GetKubeClient(kubeconfig)

	namespaces = SetNamespaceList(namespace, kubeClient)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(kubeClient, namespace)
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

func GetUnusedDeploymentsJSON(namespace string, kubeconfig string) (string, error) {
	var kubeClient *kubernetes.Clientset
	var namespaces []string

	kubeClient = GetKubeClient(kubeconfig)
	namespaces = SetNamespaceList(namespace, kubeClient)
	response := make(map[string]map[string][]string)

	for _, namespace := range namespaces {
		diff, err := ProcessNamespaceDeployments(kubeClient, namespace)
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

	return string(jsonResponse), nil
}
