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
			fmt.Printf("%s | %s | %s | %s:%d | %s\n", pod.Namespace, pod.PodID, pod.Name, pod.Host, pod.Port, pod.Status)
		}
		return nil
	},
}

var rebuildCmd = &cobra.Command{
	Use:   "rebuild <namespace>",
	Short: "Trigger a rebuild of registered pods in one namespace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(operatorURL)
		results, err := c.Rebuild(args[0])
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("No pods registered")
			return nil
		}
		for _, result := range results {
			fmt.Println(result)
		}
		return nil
	},
}

var shimURL string

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check operator or shim health",
	RunE: func(cmd *cobra.Command, args []string) error {
		if shimURL != "" {
			c := client.New(shimURL)
			if err := c.Health(); err != nil {
				return fmt.Errorf("shim at %s is unhealthy: %w", shimURL, err)
			}
			fmt.Printf("shim at %s is healthy\n", shimURL)
			return nil
		}
		c := client.New(operatorURL)
		if err := c.Health(); err != nil {
			return fmt.Errorf("operator at %s is unhealthy: %w", operatorURL, err)
		}
		fmt.Printf("operator at %s is healthy\n", operatorURL)
		return nil
	},
}

func init() {
	healthCmd.Flags().StringVar(&shimURL, "shim", "", "check the health of a specific shim by URL instead of the operator")
	rootCmd.AddCommand(listCmd, rebuildCmd, healthCmd)
}
