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

var clearCalendarCmd = &cobra.Command{
	Use:     "calendar",
	Example: "meetingepd clear calendar",
	Long:    "Clear the cache of the server and force it to fetch the latest info from the iCal",
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := client.RefreshCalendar(ctx, &pb.CalendarRequest{
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			zapLog.Fatal(fmt.Sprintf("Failed to talk to gRPC API (%s) %v", addr, err))
		}

		fmt.Print("Cleared calendar cache\n")
	},
}

var getCalendarCmd = &cobra.Command{
	Use:     "calendar",
	Example: "meetingepd get calendar",
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

		calendar, err := client.GetCalendar(ctx, &pb.CalendarRequest{Timestamp: time.Now().Unix()})
		if err != nil {
			zapLog.Fatal(fmt.Sprintf("Failed to talk to gRPC API (%s) %v", addr, err))
		}

		fmt.Printf("Got Calendar (last refreshed: %s)\n\n", time.Unix(calendar.LastUpdated, 0).Format(time.RFC822))
		for idx, item := range calendar.Entries {
			fmt.Printf("%d) ", idx)

			if item.Important {
				fmt.Print("!")
			}

			fmt.Printf("%s: [%s to %s] - %s", item.Title, time.Unix(item.Start, 0).Format(time.RFC822), time.Unix(item.End, 0).Format(time.RFC822), item.Busy.String())

			if item.AllDay {
				fmt.Print(" (all day)")
			}

			if len(item.Message) > 0 {
				fmt.Printf(": %s", item.Message)
			}

			fmt.Print("\n")
		}
	},
}

func init() {
	clearCmd.AddCommand(clearCalendarCmd)
	getCmd.AddCommand(getCalendarCmd)
}
