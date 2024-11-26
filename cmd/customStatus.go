package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/cedi/meeting_epd/pkg/api"
	pb "github.com/cedi/meeting_epd/pkg/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	title       string
	description string
	icon        string
	iconSize    int32
)

var getCustomStatusCmd = &cobra.Command{
	Use:     "status",
	Example: "meetingepd get status",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("server.debug", true)
		undo, zapLog, otelZap := initTelemetry()
		defer zapLog.Sync()
		defer undo()

		addr := fmt.Sprintf("%s:%d", server, port)

		conn, client := api.NewGrpcApiClient(otelZap, addr)
		defer conn.Close()

		// Contact the server
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		customStatus, err := client.GetCustomStatus(ctx, &pb.CustomStatusRequest{Timestamp: time.Now().Unix()})
		if err != nil {
			zapLog.Fatal(fmt.Sprintf("Failed to talk to gRPC API (%s) %v", addr, err))
		}

		fmt.Print("Custom Status:")
		if len(customStatus.Title) > 0 {
			fmt.Printf("\n")
			fmt.Printf("  - Title: %s\n", customStatus.Title)
			fmt.Printf("  - Description: %s\n", customStatus.Description)
			fmt.Printf("  - Icon: %s (%dx%d)\n", customStatus.Icon, customStatus.IconSize, customStatus.IconSize)
		} else {
			fmt.Printf(" is not set\n")
		}
	},
}

var setCustomStatusCmd = &cobra.Command{
	Use:     "status",
	Example: "meetingepd set status",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("server.debug", true)
		undo, zapLog, otelZap := initTelemetry()
		defer zapLog.Sync()
		defer undo()

		addr := fmt.Sprintf("%s:%d", server, port)

		conn, client := api.NewGrpcApiClient(otelZap, addr)
		defer conn.Close()

		// Contact the server
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		customStatus, err := client.SetCustomStatus(ctx, &pb.CustomStatus{
			Title:       title,
			Description: description,
			Icon:        icon,
			IconSize:    iconSize,
		})
		if err != nil {
			zapLog.Fatal(fmt.Sprintf("Failed to talk to gRPC API (%s) %v", addr, err))
		}

		fmt.Print("Set Custom Status:")
		if len(customStatus.Title) > 0 {
			fmt.Printf("\n")
			fmt.Printf("  - Title: %s\n", customStatus.Title)
			fmt.Printf("  - Description: %s\n", customStatus.Description)
			fmt.Printf("  - Icon: %s (%dx%d)\n", customStatus.Icon, customStatus.IconSize, customStatus.IconSize)
		} else {
			fmt.Printf(" is not set\n")
		}
	},
}

var clearCustomStatusCmd = &cobra.Command{
	Use:     "status",
	Example: "meetingepd clear status",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("server.debug", true)
		undo, zapLog, otelZap := initTelemetry()
		defer zapLog.Sync()
		defer undo()

		addr := fmt.Sprintf("%s:%d", server, port)

		conn, client := api.NewGrpcApiClient(otelZap, addr)
		defer conn.Close()

		// Contact the server
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := client.SetCustomStatus(ctx, &pb.CustomStatus{
			Title:       "",
			Description: "",
			Icon:        "",
			IconSize:    0,
		})
		if err != nil {
			zapLog.Fatal(fmt.Sprintf("Failed to talk to gRPC API (%s) %v", addr, err))
		}

		fmt.Print("Cleared Custom Status\n")
	},
}

func init() {
	setCustomStatusCmd.Flags().StringVarP(&title, "title", "t", "", "Title of your custom status")
	setCustomStatusCmd.MarkFlagRequired("title")

	setCustomStatusCmd.Flags().StringVarP(&description, "description", "d", "", "Description of your custom status")

	setCustomStatusCmd.Flags().StringVarP(&icon, "icon", "i", "warning_icon", "Icon to use in custom status")
	setCustomStatusCmd.Flags().Int32Var(&iconSize, "icon_size", 196, "Icon size to display in the custom status")

	setCmd.AddCommand(setCustomStatusCmd)
	getCmd.AddCommand(getCustomStatusCmd)
	clearCmd.AddCommand(clearCustomStatusCmd)
}
