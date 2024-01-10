/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yestool/deploy-tar/client/buildtar"
	"github.com/yestool/deploy-tar/client/config"
	"github.com/yestool/deploy-tar/client/upload"
)


var (
	cfgFile     string
	deployConfig config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deploy-tar",
	Short: "deploy tar.gz to server",
	Long: `deploy tar.gz to server`,
	Run: func(cmd *cobra.Command, args []string) { 

		deployConfig = config.Config{
			ApiKey: viper.GetString("apiKey"),
			Server: viper.GetString("server"),
			WebPath: viper.GetString("webPath"),
			TarPath: viper.GetString("tarPath"),
			WebSite: viper.GetString("webSite"),
		}
		fmt.Printf("deploy [%s] to remote server \n", deployConfig.TarPath)
		if !isTarGz(deployConfig.TarPath) {
			filePath := deployConfig.TarPath
			baseDir := filepath.Base(deployConfig.TarPath)
			deployConfig.TarPath = filepath.Join(os.TempDir(), fmt.Sprintf("%s.tar.gz", baseDir))
			buildtar.Tar(filePath, deployConfig.TarPath)
		}
		error := upload.UploadTar(deployConfig)
		if error != nil {
			fmt.Println(error)
		}
		os.Remove(deployConfig.TarPath)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
  rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /app/.deploy-tar.yaml)")
	viper.BindPFlag("apiKey", rootCmd.PersistentFlags().Lookup("apiKey"))
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("webPath", rootCmd.PersistentFlags().Lookup("webPath"))
	viper.BindPFlag("tarPath", rootCmd.PersistentFlags().Lookup("tarPath"))
  viper.SetDefault("author", "YesTool")
}

func initConfig() {
  // Don't forget to read config either from cfgFile or from home directory!
  if cfgFile != "" {
    // Use config file from the flag.
    viper.SetConfigFile(cfgFile)
  } else {
    viper.AddConfigPath("/app")
    viper.SetConfigName(".deploy-tar")
  }
	viper.AutomaticEnv()
  if err := viper.ReadInConfig(); err != nil {
    fmt.Println("Can not read config:", viper.ConfigFileUsed())
  }
}


func isTarGz(path string) bool {
	return strings.HasSuffix(path, ".tar.gz")
}

