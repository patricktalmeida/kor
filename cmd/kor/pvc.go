package kor

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonahd/kor/pkg/kor"
)

var pvcCmd = &cobra.Command{
	Use:     "persistentvolumeclaim",
	Aliases: []string{"pvc", "persistentvolumeclaims"},
	Short:   "Gets unused pvcs",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientset := kor.GetKubeClient(kubeconfig)
		if outputFormat == "json" || outputFormat == "yaml" {
			if response, err := kor.GetUnusedPvcsStructured(includeExcludeLists, clientset, outputFormat); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(response)
			}
		} else if slackWebhookURL != "" {
			kor.GetUnusedPvcsSendToSlackWebhook(includeExcludeLists, clientset, slackWebhookURL)
		} else if slackChannel != "" && slackAuthToken != "" {
			kor.GetUnusedPvcsSendToSlackAsFile(includeExcludeLists, clientset, slackChannel, slackAuthToken)
		} else {
			kor.GetUnusedPvcs(includeExcludeLists, clientset)
		}
	},
}

func init() {
	rootCmd.AddCommand(pvcCmd)
}
