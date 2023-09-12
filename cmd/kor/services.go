package kor

import (
	"github.com/spf13/cobra"
	"github.com/yonahd/kor/pkg/kor"
)

var serviceCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"svc"},
	Short:   "Gets unused services",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if outputFormat == "json" {
			kor.GetUnusedServicesJSON(namespace, kubeconfig)
		} else if slackWebhookURL != "" {
			kor.GetUnusedServicesSendToSlackWebhook(namespace, kubeconfig, slackWebhookURL)
		} else if slackChannel != "" && slackAuthToken != "" {
			kor.GetUnusedServicesSendToSlackAsFile(namespace, kubeconfig, slackChannel, slackAuthToken)
		} else {
			kor.GetUnusedServices(namespace, kubeconfig)
		}

	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
