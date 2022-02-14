/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/samirgadkari/postgresService/pkg/conn"
	"github.com/samirgadkari/postgresService/pkg/conn/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the postgres service",
	Long: `Start the postgres service. This service will talk with the sidecar
to send and receive messages using protobufs. It will process the messages and talk
to the database to complete the requests. It will send a response message back for
each request message`,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		// thisServiceAddr := viper.GetString("thisServiceAddr")
		sidecarServiceAddr := viper.GetString("sidecarServiceAddr")
		_, sidecar, err := conn.Connect(sidecarServiceAddr)
		if err != nil {
			return
		}

		sidecar.Register()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
