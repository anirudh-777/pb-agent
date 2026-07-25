package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const SchemaVersion = "1"

type Problem struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retriable bool   `json:"retriable"`
	Details   any    `json:"details,omitempty"`
}

type Envelope struct {
	SchemaVersion string   `json:"schemaVersion"`
	OK            bool     `json:"ok"`
	Command       string   `json:"command"`
	Data          any      `json:"data,omitempty"`
	Warnings      []string `json:"warnings"`
	AuditID       string   `json:"auditId,omitempty"`
	Error         *Problem `json:"error,omitempty"`
}

type CLIError struct {
	ExitCode  int
	Code      string
	Message   string
	Retriable bool
	Details   any
}

func (e *CLIError) Error() string { return e.Message }

func Write(w io.Writer, command string, data any, warnings []string, auditID string) error {
	if warnings == nil {
		warnings = []string{}
	}
	return json.NewEncoder(w).Encode(Envelope{
		SchemaVersion: SchemaVersion,
		OK:            true,
		Command:       command,
		Data:          data,
		Warnings:      warnings,
		AuditID:       auditID,
	})
}

func WriteHuman(w io.Writer, command string, data any, warnings []string, auditID string) error {
	if command == "connection.token-help" {
		return writeTokenHelp(w, data, warnings)
	}

	if _, err := fmt.Fprintf(w, "%s\n\n", command); err != nil {
		return err
	}
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(pretty)); err != nil {
		return err
	}
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(w, "\nWarning: %s\n", warning); err != nil {
			return err
		}
	}
	if auditID != "" {
		_, err = fmt.Fprintf(w, "\nAudit ID: %s\n", auditID)
	}
	return err
}

func WriteHumanError(w io.Writer, err error) int {
	message := "The operation failed."
	exitCode := 1
	if cliErr, ok := err.(*CLIError); ok {
		exitCode = cliErr.ExitCode
		message = cliErr.Message
	}
	_, _ = fmt.Fprintf(w, "Error: %s\n", message)
	return exitCode
}

func writeTokenHelp(w io.Writer, data any, warnings []string) error {
	values, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected token help data")
	}
	steps, _ := values["dashboardSteps"].([]string)
	security, _ := values["security"].([]string)

	if _, err := fmt.Fprintln(w, "Generate a PocketBase token for pb-agent"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	for index, step := range steps {
		if _, err := fmt.Fprintf(w, "%d. %s\n", index+1, step); err != nil {
			return err
		}
	}
	if command, ok := values["storeCommand"].(string); ok {
		if _, err := fmt.Fprintf(w, "\nStore it securely\n\n  %s\n", command); err != nil {
			return err
		}
	}
	if len(security) > 0 {
		if _, err := fmt.Fprintln(w, "\nSecurity"); err != nil {
			return err
		}
		for _, item := range security {
			if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
				return err
			}
		}
	}
	if revocation, ok := values["revocation"].(string); ok {
		if _, err := fmt.Fprintf(w, "\nRevocation\n%s\n", revocation); err != nil {
			return err
		}
	}
	if documentation, ok := values["documentation"].(string); ok {
		if _, err := fmt.Fprintf(w, "\nDocumentation\n%s\n", documentation); err != nil {
			return err
		}
	}
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(w, "\nWarning: %s\n", strings.TrimSpace(warning)); err != nil {
			return err
		}
	}
	return nil
}

func WriteError(w io.Writer, command string, err error) int {
	problem := &Problem{Code: "internal_error", Message: "The operation failed.", Retriable: false}
	exitCode := 1
	if cliErr, ok := err.(*CLIError); ok {
		exitCode = cliErr.ExitCode
		problem = &Problem{
			Code:      cliErr.Code,
			Message:   cliErr.Message,
			Retriable: cliErr.Retriable,
			Details:   cliErr.Details,
		}
	}
	_ = json.NewEncoder(w).Encode(Envelope{
		SchemaVersion: SchemaVersion,
		OK:            false,
		Command:       command,
		Warnings:      []string{},
		Error:         problem,
	})
	return exitCode
}

func Usage(message string) error {
	return &CLIError{ExitCode: 2, Code: "invalid_arguments", Message: message}
}

func Auth(message string) error {
	return &CLIError{ExitCode: 3, Code: "authentication_failed", Message: message}
}

func Policy(message string, details any) error {
	return &CLIError{ExitCode: 4, Code: "policy_denied", Message: message, Details: details}
}

func Conflict(message string, details any) error {
	return &CLIError{ExitCode: 5, Code: "plan_conflict", Message: message, Details: details}
}

func Connectivity(err error) error {
	return &CLIError{ExitCode: 6, Code: "connection_failed", Message: "Could not reach PocketBase.", Retriable: true, Details: fmt.Sprint(err)}
}

func Validation(message string, details any) error {
	return &CLIError{ExitCode: 7, Code: "pocketbase_validation_failed", Message: message, Details: details}
}
