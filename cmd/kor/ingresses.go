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
		clientset := kor.GetKubeClient(kubeconfig)
		if outputFormat == "json" || outputFormat == "yaml" {
			if response, err := kor.GetUnusedIngressesStructured(includeExcludeLists, clientset, outputFormat); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(response)
			}
		} else {
			kor.GetUnusedIngresses(includeExcludeLists, clientset)
		}
	},
}

func init() {
	rootCmd.AddCommand(ingressCmd)
}
