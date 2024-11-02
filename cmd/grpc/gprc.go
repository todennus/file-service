package grpc

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/spf13/cobra"
	"github.com/todennus/file-service/adapter/grpc"
	"github.com/todennus/file-service/wiring"
)

var Command = &cobra.Command{
	Use:   "grpc",
	Short: "Start the gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		envPaths, err := cmd.Flags().GetStringArray("env")
		if err != nil {
			panic(err)
		}

		system, err := wiring.InitializeSystem(envPaths...)
		if err != nil {
			panic(err)
		}

		address := fmt.Sprintf("%s:%d", system.Config.Variable.Server.Host, system.Config.Variable.Server.Port)
		app := grpc.App(system.Config, system.Usecases)

		listener, err := net.Listen("tcp", address)
		if err != nil {
			panic(err)
		}

		slog.Info("gRPC server started", "address", address)
		if err := app.Serve(listener); err != nil {
			panic(err)
		}
	},
}
