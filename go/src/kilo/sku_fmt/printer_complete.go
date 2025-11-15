package sku_fmt

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type PrinterComplete struct {
	bufferedWriter *bufio.Writer
	pool           interfaces.Pool[sku.Transacted, *sku.Transacted]
	chObjects      chan *sku.Transacted
	chDone         chan struct{}
}

func MakePrinterComplete(envLocal env_local.Env) *PrinterComplete {
	printer := &PrinterComplete{
		chObjects:      make(chan *sku.Transacted),
		chDone:         make(chan struct{}),
		bufferedWriter: bufio.NewWriter(envLocal.GetUIFile()),
		pool: pool.Make[sku.Transacted](
			nil,
			nil,
		),
	}

	envLocal.After(printer.Close)

	go func() {
		for object := range printer.chObjects {
			ui.TodoP4("handle write errors")
			printer.bufferedWriter.WriteString(object.GetObjectId().String())
			printer.bufferedWriter.WriteByte('\t')

			g := object.GetObjectId().GetGenre()
			printer.bufferedWriter.WriteString(g.String())

			tipe := object.GetType().String()

			if tipe != "" {
				printer.bufferedWriter.WriteString(": ")
				printer.bufferedWriter.WriteString(object.GetType().String())
			}

			description := object.GetMetadataMutable().Description.String()

			if description != "" {
				printer.bufferedWriter.WriteString(" ")
				printer.bufferedWriter.WriteString(
					object.GetMetadataMutable().Description.String(),
				)
			}

			printer.bufferedWriter.WriteString("\n")
			printer.pool.Put(object)
		}

		printer.chDone <- struct{}{}
	}()

	return printer
}

func (printer *PrinterComplete) PrintOne(
	src *sku.Transacted,
) (err error) {
	if src.GetObjectId().String() == "/" {
		err = errors.New("empty sku")
		return err
	}

	dst := printer.pool.Get()
	sku.Resetter.ResetWith(dst, src)

	select {
	case <-printer.chDone:
		err = errors.MakeErrStopIteration()

	case printer.chObjects <- dst:
	}

	return err
}

func (printer *PrinterComplete) Close(
	context interfaces.ActiveContext,
) (err error) {
	close(printer.chObjects)
	<-printer.chDone

	if err = context.Cause(); err != nil {
		err = nil
		return err
	}

	if err = printer.bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
