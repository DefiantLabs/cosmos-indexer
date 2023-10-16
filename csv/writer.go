package csv

import (
	"bytes"
	"encoding/csv"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/csv/parsers"
)

// Create the CSV and write it to byte buffer
func ToCsv(rows []parsers.CsvRow, headers []string) (bytes.Buffer, error) {
	var b bytes.Buffer
	w := csv.NewWriter(&b)

	if err := w.Write(headers); err != nil {
		config.Log.Error("Error writing header to csv", err)
		return b, err
	}

	// write the accointing rows to the csv
	for _, row := range rows {
		csvForRow := row.GetRowForCsv()
		if err := w.Write(csvForRow); err != nil {
			config.Log.Error("Error writing header to csv", err)
			return b, err
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		config.Log.Error("Error writing header to csv", err)
		return b, err
	}

	return b, nil
}
