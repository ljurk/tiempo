package lib

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

type Tiempo struct {
	Date          string `csv:"date"`
	Start         string `csv:"start"`
	End           string `csv:"end"`
	Breakduration string `csv:"break"`
}

// Expand ~ and environment variables in the path
func ExpandPath(path string) (string, error) {
	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand the ~ to the user's home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir + path[1:], nil
	}
	return path, nil
}

func (record *Tiempo) Duration() string {
	// Define the time format
	timeFormat := "15:04" // 24-hour format

	// Parse the time strings into time.Time objects
	startTime, err := time.Parse(timeFormat, record.Start)
	if err != nil {
		return "--:--"
	}
	endTime, err := time.Parse(timeFormat, record.End)
	if err != nil {
		return "--:--"
	}

	// Calculate the duration between the two times
	duration := endTime.Sub(startTime)

	return fmt.Sprintf("%d:%02d", int(duration.Hours()), int(duration.Minutes())%60)
}

func ContainsDate(tiempos []Tiempo, date string) bool {
	for _, tiempo := range tiempos {
		if tiempo.Date == date {
			return true
		}
	}
	return false
}

func PrintHeader() string {
	record := Tiempo{}
	t := reflect.TypeOf(record)

	out := ""
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("csv")
		out += fmt.Sprintf("%s,", tag)
	}
	return out[:len(out)-1] + "\n"
}

func PrintRecordWithTags(record interface{}) {
	v := reflect.ValueOf(record)
	t := reflect.TypeOf(record)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("csv")
		fmt.Printf("%s: %v\n", tag, field.Interface())
	}
	fmt.Println()
}

func Read(filepath string) []Tiempo {
	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the CSV file into a slice of Record structs
	var records []Tiempo
	if err := gocsv.UnmarshalFile(file, &records); err != nil {
		panic(err)
	}

	return records
}

func Write(filepath string, records []Tiempo) error {
	// Open the file for writing
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the slice of Tiempo structs to the file as CSV
	if err := gocsv.MarshalFile(&records, file); err != nil {
		return err
	}

	return nil
}

func UpdateTime(filepath, timeField string, newTime string) error {
	// Open the CSV file
	records := Read(filepath)

	current_date := time.Now().Local().Format("02.01.2006")

	// Update the time field for the matching date or add a new record
	if ContainsDate(records, current_date) {
		for i, record := range records {
			if record.Date == current_date {
				switch timeField {
				case "Start":
					records[i].Start = newTime
				case "End":
					records[i].End = newTime
				}
			}
		}
	} else {
		// If the date doesn't exist and it's for Start, create a new entry
		if timeField == "Start" {
			records = append(records, Tiempo{Date: current_date, Start: newTime})
		}
	}

	// Write the updated records back to the CSV file
	return Write(filepath, records)
}
