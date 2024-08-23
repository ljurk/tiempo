package cmd

import (
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"
	"time"

	"github.com/ljurk/tiempo/lib"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "set Start",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("Error getting flag:", err)
			os.Exit(1)
		}
		// Expand the path
		filepath, err = lib.ExpandPath(filepath)
		if err != nil {
			fmt.Println("Error expanding path:", err)
			os.Exit(1)
		}
		current_time := time.Now().Local().Format("15:04")

		if err := lib.UpdateTime(filepath, "Start", current_time); err != nil {
			panic(err)
		}
		fmt.Println("Start time recorded:", current_time)
	},
}

var endCmd = &cobra.Command{
	Use:   "end",
	Short: "set End",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("Error getting flag:", err)
			os.Exit(1)
		}
		// Expand the path
		filepath, err = lib.ExpandPath(filepath)
		if err != nil {
			fmt.Println("Error expanding path:", err)
			os.Exit(1)
		}
		current_time := time.Now().Local().Format("15:04")

		if err := lib.UpdateTime(filepath, "End", current_time); err != nil {
			panic(err)
		}
		fmt.Println("End time recorded:", current_time)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show work hours per day",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("Error getting flag:", err)
			os.Exit(1)
		}
		// Expand the path
		filepath, err = lib.ExpandPath(filepath)
		if err != nil {
			fmt.Println("Error expanding path:", err)
			os.Exit(1)
		}
		// Create a tab writer
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer writer.Flush()
		// Print the table header
		header := "Date\tStart\tEnd\tDuration"
		fmt.Fprintln(writer, header)
		fmt.Fprintln(writer, regexp.MustCompile(`\w`).ReplaceAllString(header, "~"))

		for _, record := range lib.Read(filepath) {
			// Print the table row
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
				record.Date,
				record.Start,
				record.End,
				record.Duration(),
			)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("Error getting flag:", err)
			os.Exit(1)
		}
		// Expand the path
		filepath, err = lib.ExpandPath(filepath)
		if err != nil {
			fmt.Println("Error expanding path:", err)
			os.Exit(1)
		}
		if _, err := os.Stat(filepath); err == nil {
			fmt.Printf("%s already exists\n", filepath)
			return
		}

		file, err := os.Create(filepath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.WriteString(lib.PrintHeader())
	},
}

// Function to set up common flags
func setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "~/tiempo.csv", "path to CSV file")
}

func init() {
	setupFlags(initCmd)
	setupFlags(startCmd)
	setupFlags(endCmd)
	setupFlags(statusCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
	rootCmd.AddCommand(statusCmd)
}
