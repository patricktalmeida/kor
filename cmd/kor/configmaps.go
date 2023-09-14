package kor

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonahd/kor/pkg/kor"
)

var configmapCmd = &cobra.Command{
	Use:     "configmap",
	Aliases: []string{"cm", "configmaps"},
	Short:   "Gets unused configmaps",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		clientset := kor.GetKubeClient(kubeconfig)
		if outputFormat == "json" || outputFormat == "yaml" {
			if response, err := kor.GetUnusedConfigmapsStructured(includeExcludeLists, clientset, outputFormat); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(response)
			}
		} else if slackWebhookURL != "" {
			kor.GetUnusedConfigmapsSendToSlackWebhook(includeExcludeLists, clientset, slackWebhookURL)
		} else if slackChannel != "" && slackAuthToken != "" {
			kor.GetUnusedConfigmapsSendToSlackAsFile(includeExcludeLists, clientset, slackChannel, slackAuthToken)
		} else {
			kor.GetUnusedConfigmaps(includeExcludeLists, clientset)
		}

	},
}

func init() {
	rootCmd.AddCommand(configmapCmd)
}
