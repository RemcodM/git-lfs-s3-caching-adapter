package adapter

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ProcessStandaloneData is the primary endpoint for processing data with a
// standalone transfer agent. It reads input from the specified input file and
// produces output to the specified output file.
func ProcessData(input *os.File, output *os.File) error {
	var handler *cachingHandler

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		var msg inputMessage
		if err := json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&msg); err != nil {
			return errors.Wrapf(err, "error decoding JSON")
		}
		if handler == nil {
			var err error
			handler, err = newHandler(output, &msg)
			if err != nil {
				err := errors.Wrapf(err, "error creating handler")
				errMsg := outputErrorMessage{
					Error: errorMessage{
						Message: err.Error(),
					},
				}
				json.NewEncoder(output).Encode(errMsg)
				return err
			}
		}
		if !handler.dispatch(&msg) {
			break
		}
	}
	if handler != nil {
		os.RemoveAll(handler.tempdir)
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "error reading input")
	}
	return nil
}
