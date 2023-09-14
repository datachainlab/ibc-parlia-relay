package main

import (
	"context"
	"github.com/datachainlab/ibc-parlia-relay/tool/testdata/internal"
	"github.com/spf13/cobra"
	"log"
)

func main() {
	cobra.EnableCommandSorting = false

	var rootCmd = &cobra.Command{}
	rootCmd.AddCommand(internal.CreateMisbehavior())
	rootCmd.AddCommand(internal.CreateCreateClient())
	rootCmd.AddCommand(internal.CreateUpdateClient())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.Panicf("Failed to run command : %+v", err)
		return
	}
}