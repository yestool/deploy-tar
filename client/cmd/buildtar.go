/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/yestool/deploy-tar/client/buildtar"
	"github.com/spf13/cobra"
)

// buildtarCmd represents the buildtar command
var buildtarCmd = &cobra.Command{
	Use:   "buildtar",
	Short: "Pack the directory as tar.gz",
	Long: `Pack the directory as tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		srcDir := "/root/go-code/deploy-tar/test"
		destFile := "/root/go-code/deploy-tar/web.tar.gz"
		buildtar.Tar(srcDir, destFile)
	},
}

func init() {
	rootCmd.AddCommand(buildtarCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildtarCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//  
}
