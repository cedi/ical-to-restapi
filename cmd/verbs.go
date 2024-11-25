package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	server string
	port   int
)

var getCmd = &cobra.Command{
	Use:     "get",
	Example: "meetingepd get",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var setCmd = &cobra.Command{
	Use:     "set",
	Example: "meetingepd set",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var clearCmd = &cobra.Command{
	Use:     "clear",
	Example: "meetingepd clear",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func addConnFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&server, "server", "s", "", "Port of the gRPC API of the Server")
	cmd.MarkFlagRequired("server")

	cmd.Flags().IntVar(&port, "port", 50051, "Port of the gRPC API of the Server")
	viper.BindPFlag("server.grpcPort", cmd.Flags().Lookup("port"))
}

func init() {
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
}
