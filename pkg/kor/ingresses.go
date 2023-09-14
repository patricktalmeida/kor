package kor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/yaml"
)

func validateServiceBackend(kubeClient *kubernetes.Clientset, namespace string, backend *v1.IngressBackend) bool {
	if backend.Service != nil {
		serviceName := backend.Service.Name

		_, err := kubeClient.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
		if err != nil {
			return false
		}
	}
	return true
}

func retrieveUsedIngress(kubeClient *kubernetes.Clientset, namespace string) ([]string, error) {
	ingresses, err := kubeClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	usedIngresses := []string{}

	for _, ingress := range ingresses.Items {
		if ingress.Labels["kor/used"] == "true" {
			continue
		}

		used := true

		if ingress.Spec.DefaultBackend != nil {
			used = validateServiceBackend(kubeClient, namespace, ingress.Spec.DefaultBackend)
		}
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				used = validateServiceBackend(kubeClient, namespace, &path.Backend)
				if used {
					break
				}
			}
			if used {
				break
			}
		}
		if used {
			usedIngresses = append(usedIngresses, ingress.Name)
		}
	}
	return usedIngresses, nil
}

func retrieveIngressNames(kubeClient *kubernetes.Clientset, namespace string) ([]string, error) {
	ingresses, err := kubeClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ingresses.Items))
	for _, ingress := range ingresses.Items {
		names = append(names, ingress.Name)
	}
	return names, nil
}

func processNamespaceIngresses(kubeClient *kubernetes.Clientset, namespace string) ([]string, error) {
	usedIngresses, err := retrieveUsedIngress(kubeClient, namespace)
	if err != nil {
		return nil, err
	}
	ingressNames, err := retrieveIngressNames(kubeClient, namespace)
	if err != nil {
		return nil, err
	}

	diff := CalculateResourceDifference(usedIngresses, ingressNames)
	return diff, nil

}

func GetUnusedIngresses(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	for _, namespace := range namespaces {
		diff, err := processNamespaceIngresses(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Ingresses")
		fmt.Println(output)
		fmt.Println()
	}
}

func GetUnusedIngressesSendToSlackWebhook(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, slackWebhookURL string) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := processNamespaceIngresses(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Ingresses")

		outputBuffer.WriteString(output)
		outputBuffer.WriteString("\n")
	}

	if err := SendToSlackWebhook(slackWebhookURL, outputBuffer.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send payload to Slack: %v\n", err)
	}
}

func GetUnusedIngressesSendToSlackAsFile(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, slackChannel string, slackAuthToken string) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)

	var outputBuffer bytes.Buffer

	for _, namespace := range namespaces {
		diff, err := processNamespaceIngresses(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		output := FormatOutput(namespace, diff, "Ingresses")

		outputBuffer.WriteString(output)
		outputBuffer.WriteString("\n")
	}

	outputFilePath, _ := writeOutputToFile(outputBuffer)

	if err := SendFileToSlack(outputFilePath, "Unused Ingresses", slackChannel, slackAuthToken); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send output to Slack: %v\n", err)
	}
}

func GetUnusedIngressesStructured(includeExcludeLists IncludeExcludeLists, clientset *kubernetes.Clientset, outputFormat string) (string, error) {
	var namespaces []string

	namespaces = SetNamespaceList(includeExcludeLists, clientset)
	response := make(map[string]map[string][]string)

	for _, namespace := range namespaces {
		diff, err := processNamespaceIngresses(clientset, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		resourceMap := make(map[string][]string)
		resourceMap["Ingresses"] = diff
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
