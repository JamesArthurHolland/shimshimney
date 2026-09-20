package cmd

import (
	"fmt"
	"os"

	"github.com/shimshimney/pkg/client"
	"github.com/spf13/cobra"
)

var operatorURL string

var rootCmd = &cobra.Command{
	Use:   "shimshimney",
	Short: "Interact with the shimshimney operator and pods",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&operatorURL, "operator", "o", "http://localhost:8080", "operator API URL")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered pods and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(operatorURL)
		pods, err := c.ListPods()
		if err != nil {
			return err
		}
		if len(pods) == 0 {
			fmt.Println("No pods registered")
			return nil
		}
		for _, pod := range pods {
			fmt.Printf("%s | %s | %s:%d | %s\n", pod.PodID, pod.Name, pod.Host, pod.Port, pod.Status)
		}
		return nil
	},
}

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Trigger a rebuild of all registered pods",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(operatorURL)
		return c.Rebuild()
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check operator health",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(operatorURL)
		return c.Health()
	},
}

func init() {
	rootCmd.AddCommand(listCmd, rebuildCmd, healthCmd)
}
