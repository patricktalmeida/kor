package kor

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonahd/kor/pkg/kor"
)

var ingressCmd = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"ing", "ingresses"},
	Short:   "Gets unused ingresses",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if outputFormat == "json" || outputFormat == "yaml" {
			if response, err := kor.GetUnusedIngressesStructured(includeExcludeLists, kubeconfig, outputFormat); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(response)
			}
		} else if slackWebhookURL != "" {
			kor.GetUnusedIngressesSendToSlackWebhook(includeExcludeLists, kubeconfig, slackWebhookURL)
		} else if slackChannel != "" && slackAuthToken != "" {
			kor.GetUnusedIngressesSendToSlackAsFile(includeExcludeLists, kubeconfig, slackChannel, slackAuthToken)
		} else {
			kor.GetUnusedIngresses(includeExcludeLists, kubeconfig)
		}
	},
}

func init() {
	rootCmd.AddCommand(ingressCmd)
}
