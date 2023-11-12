package main

import (
	"fmt"
	"os"

	"github.com/MrSantamaria/acceptance_test/cmd"
	"github.com/MrSantamaria/acceptance_test/workflows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "acceptance_test",
	Short: "acceptance_test is a component of the Hypershift Operator Promotion process",
	Long:  `acceptance_test is a tool used to validate Hypershift Operator Promotions ocurred successfully`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		err = workflows.SetUp(viper.GetString("token"), viper.GetString("environment"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = workflows.AcceptanceTest()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func main() {
	cmd.InitEnv(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
