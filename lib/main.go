package lib

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Period struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type Day struct {
	Date    string   `yaml:"date"`
	Working []Period `yaml:"working"`
	Breaks  []Period `yaml:"breaks"`
}

type Tiempo struct {
	Targets map[string]string
	Days    []Day `yaml:"days"`
}

const timeFormat string = "15:04"
const DateFormat string = "02.01.2006"

func PrintPeriods(periods []Period) string {
	result := ""
	for _, period := range periods {
		result += fmt.Sprintf("%s-%s;", period.Start, period.End)
	}
	if len(result) > 0 {
		return result[:len(result)-1] // Remove trailing semicolon
	}
	return result
}

func CalculateDuration(periods []Period) time.Duration {
	total := time.Duration(0)
	for _, period := range periods {
		start, err := time.Parse(timeFormat, period.Start)
		if err != nil {
			return total
		}
		end, err := time.Parse(timeFormat, period.End)
		if err != nil {
			return total
		}
		total += end.Sub(start)
	}
	return total
}

func Read(filepath string) (Tiempo, error) {
	// Open the YAML file
	file, err := os.Open(filepath)
	if err != nil {
		file, err = os.Create(filepath)
	}
	defer file.Close()

	// Read the file contents into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		return Tiempo{}, err
	}

	// Unmarshal the YAML into a slice of Tiempo structs
	var records Tiempo
	if err := yaml.Unmarshal(data, &records); err != nil {
		return Tiempo{}, err
	}
	// fill Targets if there are none
	if len(records.Targets) == 0 {
		for i := time.Monday; i <= time.Saturday; i++ {
			records.Targets[strings.ToLower(i.String())] = "0h"
		}
		records.Targets[strings.ToLower(time.Sunday.String())] = "0h"
	}

	return records, nil
}

func Write(filepath string, records Tiempo) error {
	// Marshal the slice of Tiempo structs to YAML
	data, err := yaml.Marshal(&records)
	if err != nil {
		return err
	}

	// Open the file for writing
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the YAML data to the file
	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func UpdateTime(filepath string, timeField string, periodType string, newTime string) error {
	records, err := Read(filepath)
	if err != nil {
		return err
	}
	currentDate := time.Now().Local().Format(DateFormat)
	recordFound := false

	updatePeriod := func(periods *[]Period, newTime string) error {
		switch timeField {
		case "Start":
			for _, period := range *periods {
				if period.End == "" {
					return fmt.Errorf("end your last period before starting a new one")
				}
			}
			*periods = append(*periods, Period{Start: newTime})
		case "End":
			updated := false
			for i := range *periods {
				if (*periods)[i].End == "" {
					(*periods)[i].End = newTime
					updated = true
					break
				}
			}
			if !updated {
				return fmt.Errorf("there is no started period to end")
			}
		}
		return nil
	}

	for i, record := range records.Days {
		if record.Date != currentDate {
			continue
		}
		recordFound = true
		switch periodType {
		case "working":
			if err := updatePeriod(&records.Days[i].Working, newTime); err != nil {
				return err
			}
		case "break":
			if err := updatePeriod(&records.Days[i].Breaks, newTime); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown period type: %s", periodType)
		}
		break
	}

	if !recordFound {
		if timeField == "Start" && periodType == "working" {
			records.Days = append(records.Days, Day{Date: currentDate, Working: []Period{{Start: newTime}}})
		} else {
			return fmt.Errorf("no existing record for the date, and cannot set End without a Start")
		}
	}

	if err := Write(filepath, records); err != nil {
		return fmt.Errorf("failed to write updated records to file %s: %w", filepath, err)
	}
	return nil
}
