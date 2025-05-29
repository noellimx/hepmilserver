package bytes

import (
	"bytes"
	"encoding/csv"
)

func TwoDStringAsBytes(data [][]string) (*bytes.Buffer, error) {
	// Create CSV file
	var buf bytes.Buffer

	// Create CSV writer
	writer := csv.NewWriter(&buf)

	// Write the data to the file
	err := writer.WriteAll(data)
	if err != nil {
		return nil, err
	}
	writer.Flush()

	// Return the file object

	return &buf, nil
}
