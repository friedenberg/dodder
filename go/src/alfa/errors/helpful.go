package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Helpful interface {
	error
	GetHelpfulError() Helpful
	ErrorCause() []string
	ErrorRecovery() []string
}

type Retryable interface {
	GetRetryableError() Retryable
	Recover(interfaces.RetryableContext, error)
}

func PrintHelpful(printer interfaces.Printer, helpful Helpful) {
	printer.Printf("Error: %s", helpful.Error())
	printer.Printf("\nCause:")

	for _, causeLine := range helpful.ErrorCause() {
		printer.Print(causeLine)
	}

	printer.Printf("\nRecovery:")

	for _, recoveryLine := range helpful.ErrorRecovery() {
		printer.Print(recoveryLine)
	}
}
