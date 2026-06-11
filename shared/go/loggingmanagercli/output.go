package loggingmanagercli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

type outputWriter struct{}

// NewOutputWriter creates output formatting worker.
func NewOutputWriter() OutputWriter {
	return outputWriter{}
}

// WriteEvents renders events in requested output format.
func (w outputWriter) WriteEvents(
	writer io.Writer,
	events []loggingmanagercontract.ListAlertItem,
	jsonOutput bool,
) error {
	if jsonOutput {
		return writeJSON(writer, events)
	}

	table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "EVENT_ID\tCATEGORY\tSEVERITY\tSTATE\tACK\tMUTED\tSOURCE\tCREATED_AT"); err != nil {
		return err
	}

	for _, item := range events {
		if _, err := fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\n",
			item.EventID,
			item.Category,
			item.Severity,
			item.Status,
			item.Acknowledged,
			item.Muted,
			item.Source,
			item.CreatedAt,
		); err != nil {
			return err
		}
	}

	return table.Flush()
}

// WriteOK renders mutation acknowledgements.
func (w outputWriter) WriteOK(writer io.Writer, response loggingmanagercontract.OKResponse, jsonOutput bool) error {
	if jsonOutput {
		return writeJSON(writer, response)
	}

	_, err := fmt.Fprintln(writer, strings.ToUpper(fmt.Sprintf("ok=%t", response.OK)))
	return err
}

func writeJSON(writer io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintln(writer, string(data)); err != nil {
		return err
	}

	return nil
}
