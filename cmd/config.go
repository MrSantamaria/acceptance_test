package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	selectors []string
)

func InitEnv(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().String("token", "", "OCM Token")
	rootCmd.PersistentFlags().String("env", "", "Environment")
	rootCmd.PersistentFlags().String("operator", "", "operatorName")
	rootCmd.PersistentFlags().StringSliceVar(&selectors, "selectors", nil, "comma-separated list of cluster deployment selectors")
	rootCmd.PersistentFlags().String("imagetag", "", "Image Tag")
	rootCmd.PersistentFlags().String("telemeterClientID", "", "TELEMETER_CLIENT_ID")
	rootCmd.PersistentFlags().String("telemeterSecret", "", "TELEMETER_SECRET")
	rootCmd.PersistentFlags().String("telemeterSearchTime", "10m", "TELEMETER_SEARCH_TIME")

	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("environment", rootCmd.PersistentFlags().Lookup("env"))
	viper.BindPFlag("operator", rootCmd.PersistentFlags().Lookup("operator"))
	viper.BindPFlag("selectors", rootCmd.PersistentFlags().Lookup("selectors"))
	viper.BindPFlag("imagetag", rootCmd.PersistentFlags().Lookup("imagetag"))
	viper.BindPFlag("telemeterClientID", rootCmd.PersistentFlags().Lookup("telemeterClientID"))
	viper.BindPFlag("telemeterSecret", rootCmd.PersistentFlags().Lookup("telemeterSecret"))
	viper.BindPFlag("telemeterSearchTime", rootCmd.PersistentFlags().Lookup("telemeterSearchTime"))

	viper.AutomaticEnv()
}
