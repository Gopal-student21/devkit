package servcmd

import (
	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/internal/web"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the visual dashboard",
		Long:  "Opens a web dashboard at localhost:8080 to manage all your CLI tools visually",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			port := cfg.Port
			if port == 0 {
				port = 8080
			}
			web.StartServer(port)
		},
	}
}
