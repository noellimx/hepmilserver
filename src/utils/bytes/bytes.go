package bytes

import (
	"bytes"
	"encoding/csv"
)

func TwoDStringAsBytes(data [][]string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	defer writer.Flush()

	err := writer.WriteAll(data)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}
