package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/groovy-sky/azemailsender"
)

// Formatter handles output formatting for different modes
type Formatter struct {
	JSON  bool
	Quiet bool
	Debug bool
}

// NewFormatter creates a new output formatter
func NewFormatter(jsonOutput, quiet, debug bool) *Formatter {
	return &Formatter{
		JSON:  jsonOutput,
		Quiet: quiet,
		Debug: debug,
	}
}

// PrintSendResponse formats and prints send response
func (f *Formatter) PrintSendResponse(response *azemailsender.SendResponse) error {
	if f.JSON {
		return f.printJSON(map[string]interface{}{
			"id":        response.ID,
			"status":    response.Status,
			"timestamp": response.Timestamp.Format(time.RFC3339),
		})
	}

	if !f.Quiet {
		fmt.Printf("Email sent successfully!\n")
		fmt.Printf("Message ID: %s\n", response.ID)
		if response.Status != "" {
			fmt.Printf("Status: %s\n", response.Status)
		}
	}
	return nil
}

// PrintStatusResponse formats and prints status response
func (f *Formatter) PrintStatusResponse(response *azemailsender.StatusResponse) error {
	if f.JSON {
		return f.printJSON(map[string]interface{}{
			"id":        response.ID,
			"status":    response.Status,
			"timestamp": response.Timestamp.Format(time.RFC3339),
			"error":     response.Error,
		})
	}

	if !f.Quiet {
		fmt.Printf("Message ID: %s\n", response.ID)
		fmt.Printf("Status: %s\n", response.Status)
		fmt.Printf("Timestamp: %s\n", response.Timestamp.Format(time.RFC3339))
		if response.Error != nil {
			fmt.Printf("Error: %s\n", response.Error.Message)
		}
	}
	return nil
}

// PrintError formats and prints error messages
func (f *Formatter) PrintError(err error) {
	if f.JSON {
		f.printJSON(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// PrintInfo prints informational messages (only if not quiet)
func (f *Formatter) PrintInfo(message string, args ...interface{}) {
	if f.Quiet {
		return
	}

	if f.JSON {
		f.printJSON(map[string]interface{}{
			"info": fmt.Sprintf(message, args...),
		})
		return
	}

	fmt.Printf(message+"\n", args...)
}

// PrintDebug prints debug messages (only if debug enabled)
func (f *Formatter) PrintDebug(message string, args ...interface{}) {
	if !f.Debug {
		return
	}

	if f.JSON {
		f.printJSON(map[string]interface{}{
			"debug": fmt.Sprintf(message, args...),
		})
		return
	}

	fmt.Printf("[DEBUG] "+message+"\n", args...)
}

// PrintSuccess prints success messages
func (f *Formatter) PrintSuccess(message string, args ...interface{}) error {
	if f.JSON {
		return f.printJSON(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf(message, args...),
		})
	}

	if !f.Quiet {
		fmt.Printf("✓ "+message+"\n", args...)
	}
	return nil
}

// PrintConfig prints configuration in a formatted way
func (f *Formatter) PrintConfig(config interface{}) error {
	if f.JSON {
		return f.printJSON(config)
	}

	// Pretty print configuration
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// printJSON prints data as JSON
func (f *Formatter) printJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// FormatRecipients formats recipient list for display
func FormatRecipients(recipients []string) string {
	if len(recipients) == 0 {
		return "none"
	}
	return strings.Join(recipients, ", ")
}