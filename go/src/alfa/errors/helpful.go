package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func PrintHelpful(printer interfaces.Printer, helpful interfaces.ErrorHelpful) {
	printer.Printf("Error: %s", helpful.Error())
	printer.Printf("\nCause:")

	for _, causeLine := range helpful.GetErrorCause() {
		printer.Print(causeLine)
	}

	printer.Printf("\nRecovery:")

	for _, recoveryLine := range helpful.GetErrorRecovery() {
		printer.Print(recoveryLine)
	}
}
