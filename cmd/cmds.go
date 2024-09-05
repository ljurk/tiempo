package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ljurk/tiempo/lib"
	"github.com/spf13/cobra"
)

func getFilePath(cmd *cobra.Command) (string, error) {
	filepath, err := cmd.Flags().GetString("file")
	if err != nil {
		return "", fmt.Errorf("error getting flag: %w", err)
	}
	// Expand environment variables
	filepath = os.ExpandEnv(filepath)

	// Expand the ~ to the user's home directory
	if strings.HasPrefix(filepath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir + filepath[1:], nil
	}
	return filepath, nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "set Start",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := getFilePath(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		periodType, err := cmd.Flags().GetString("type")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		current_time := time.Now().Local().Format("15:04")

		if err := lib.UpdateTime(filepath, "Start", periodType, current_time); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Start time recorded:", current_time)
	},
}

var endCmd = &cobra.Command{
	Use:   "end",
	Short: "set End",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, err := getFilePath(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		periodType, err := cmd.Flags().GetString("type")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		current_time := time.Now().Local().Format("15:04")

		if err := lib.UpdateTime(filepath, "End", periodType, current_time); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("End time recorded:", current_time)
	},
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	if d < 0 {
		d = -d
	}
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show work hours per day",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filepath, err := getFilePath(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Create a tab writer
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer writer.Flush()

		// Print the table header
		header := "Date\tNetWorkingTime\tNetBreakTime\tTargetDuration\tDiff"
		if verbose {
			header += "\tWorking\tBreaks"
		}
		fmt.Fprintln(writer, header)
		// underline header
		fmt.Fprintln(writer, regexp.MustCompile(`\w`).ReplaceAllString(header, "~"))

		records, err := lib.Read(filepath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}

		totalNetWorkingTime := time.Duration(0)
		totalBreakTime := time.Duration(0)
		totalDiff := time.Duration(0)
		for _, record := range records.Days {
			// Print the table row
			date, err := time.Parse(lib.DateFormat, record.Date)
			if err != nil {
				fmt.Println("Error parsing date:", err)
				return
			}

			// Get the day of the week
			dayOfWeek := strings.ToLower(date.Weekday().String())
			targetDuration, _ := time.ParseDuration(records.Targets[dayOfWeek])

			// Print row
			fmt.Fprintf(writer,
				"%s,%s\t%s\t%s\t%s\t%s",
				dayOfWeek,
				record.Date,
				fmtDuration(lib.CalculateDuration(record.Working)-lib.CalculateDuration(record.Breaks)),
				fmtDuration(lib.CalculateDuration(record.Breaks)),
				fmtDuration(targetDuration),
				fmtDuration(lib.CalculateDuration(record.Working)-lib.CalculateDuration(record.Breaks)-targetDuration))
			if verbose {
				fmt.Fprintf(writer,
					"\t%s\t%s",
					lib.PrintPeriods(record.Working),
					lib.PrintPeriods(record.Breaks))
			}
			fmt.Fprintf(writer, "\n")

			totalNetWorkingTime += lib.CalculateDuration(record.Working) - lib.CalculateDuration(record.Breaks)
			totalBreakTime += lib.CalculateDuration(record.Breaks)
			totalDiff += lib.CalculateDuration(record.Working) - lib.CalculateDuration(record.Breaks) - targetDuration
		}

		// Print summary
		fmt.Fprintln(writer, regexp.MustCompile(`\w`).ReplaceAllString(header, "="))
		fmt.Fprintf(writer,
			"%s\t%s\t%s\t\t%s\n",
			"Total",
			totalNetWorkingTime,
			totalBreakTime,
			totalDiff)

	},
}

// Function to set up common flags
func setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "~/tiempo.yml", "path to CSV file")
}

func init() {
	startCmd.Flags().StringP("type", "t", "working", "either [working, break]")
	endCmd.Flags().StringP("type", "t", "working", "either [working, break]")
	statusCmd.Flags().BoolP("verbose", "v", false, "")
	setupFlags(startCmd)
	setupFlags(endCmd)
	setupFlags(statusCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
	rootCmd.AddCommand(statusCmd)
}
