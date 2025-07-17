package sku_fmt

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
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
		pool: pool.MakePool[sku.Transacted](
			nil,
			nil,
		),
	}

	envLocal.After(printer.Close)

	go func(s *PrinterComplete) {
		for sk := range s.chObjects {
			ui.TodoP4("handle write errors")
			s.bufferedWriter.WriteString(sk.GetObjectId().String())
			s.bufferedWriter.WriteByte('\t')

			g := sk.GetObjectId().GetGenre()
			s.bufferedWriter.WriteString(g.String())

			tipe := sk.GetType().String()

			if tipe != "" {
				s.bufferedWriter.WriteString(": ")
				s.bufferedWriter.WriteString(sk.GetType().String())
			}

			description := sk.GetMetadata().Description.String()

			if description != "" {
				s.bufferedWriter.WriteString(" ")
				s.bufferedWriter.WriteString(sk.GetMetadata().Description.String())
			}

			s.bufferedWriter.WriteString("\n")
			printer.pool.Put(sk)
		}

		s.chDone <- struct{}{}
	}(printer)

	return printer
}

func (printer *PrinterComplete) PrintOne(
	src *sku.Transacted,
) (err error) {
	if src.GetObjectId().String() == "/" {
		err = errors.New("empty sku")
		return
	}

	dst := printer.pool.Get()
	sku.Resetter.ResetWith(dst, src)

	select {
	case <-printer.chDone:
		err = collections.MakeErrStopIteration()

	case printer.chObjects <- dst:
	}

	return
}

func (printer *PrinterComplete) Close(context interfaces.Context) (err error) {
	close(printer.chObjects)
	<-printer.chDone

	if err = context.Cause(); err != nil {
		err = nil
		return
	}

	if err = printer.bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
