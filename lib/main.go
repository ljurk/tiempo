package lib

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Period struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type Tiempo struct {
	Date    string   `yaml:"date"`
	Working []Period `yaml:"working"`
	Breaks  []Period `yaml:"breaks"`
}

const timeFormat string = "15:04"
const dateFormat string = "02.01.2006"

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

func Read(filepath string) ([]Tiempo, error) {
	// Open the YAML file
	file, err := os.Open(filepath)
	if err != nil {
		file, err = os.Create(filepath)
	}
	defer file.Close()

	// Read the file contents into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML into a slice of Tiempo structs
	var records []Tiempo
	if err := yaml.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func Write(filepath string, records []Tiempo) error {
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

func UpdateTime(filepath, timeField string, periodType string, newTime string) error {
	records, err := Read(filepath)
	if err != nil {
		return err
	}
	current_date := time.Now().Local().Format(dateFormat)
	recordFound := false

	// Update the time field for the matching date or add a new record
	for i, record := range records {
		if record.Date == current_date {
			recordFound = true
			switch periodType {
			case "working":
				switch timeField {
				case "Start":
					for j, _ := range record.Working {
						if record.Working[j].End == "" {
							return fmt.Errorf("End your last Period before starting a new")
						}
					}
					records[i].Working = append(record.Working, Period{Start: newTime})
				case "End":
					updated := false
					// Find the last Period without an End time and update it
					for j, _ := range record.Working {
						if record.Working[j].End == "" {
							records[i].Working[j].End = newTime
							updated = true
							break
						}
					}
					if !updated {
						return fmt.Errorf("There is no started Period to end")
					}
				}
			case "break":
				switch timeField {
				case "Start":
					for j, _ := range record.Breaks {
						if record.Breaks[j].End == "" {
							return fmt.Errorf("End your last Period before starting a new")
						}
					}
					records[i].Breaks = append(record.Breaks, Period{Start: newTime})
				case "End":
					updated := false
					// Find the last Period without an End time and update it
					for j, _ := range record.Breaks {
						if record.Breaks[j].End == "" {
							records[i].Breaks[j].End = newTime
							updated = true
							break
						}
					}
					if !updated {
						return fmt.Errorf("There is no started Period to end")
					}
				}
			}
		}
	}

	if !recordFound {
		// If the date doesn't exist and it's for Start, create a new entry
		if timeField == "Start" {
			records = append(records, Tiempo{Date: current_date, Working: []Period{{Start: newTime}}})
		} else {
			return fmt.Errorf("no existing record for the date, and cannot set End without a Start")
		}
	}

	if err := Write(filepath, records); err != nil {
		return fmt.Errorf("failed to write updated records to file %s: %w", filepath, err)
	}
	// Write the updated records back to file
	return nil
}
