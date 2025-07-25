package local_working_copy

import (
	"fmt"
	"maps"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	FormatFuncConstructorEntry struct {
		description string
		FormatFuncConstructor
	}

	FormatFuncConstructor func(
		*Repo,
		interfaces.WriterAndStringWriter,
	) interfaces.FuncIter[*sku.Transacted]

	FormatFlag struct {
		*Repo

		value       string
		description string
		constructor FormatFuncConstructor
	}
)

func (formatFlag *FormatFlag) String() string {
	if formatFlag == nil || formatFlag.constructor == nil {
		return fmt.Sprintf(
			"%q",
			slices.Collect(maps.Keys(formatters)),
		)
	} else if formatFlag.description != "" {
		return fmt.Sprintf("%s: %s", formatFlag.value, formatFlag.description)
	} else {
		return formatFlag.value
	}
}

var formatterCompletions = func() map[string]string {
	completion := make(map[string]string, len(formatters))

	for name, entry := range formatters {
		if entry.description != "" {
			completion[name] = name
		} else {
			completion[name] = entry.description
		}
	}

	return completion
}()

func (formatFlag *FormatFlag) GetCLICompletion() map[string]string {
	return formatterCompletions
}

func (formatFlag *FormatFlag) Set(value string) (err error) {
	var ok bool
	var entry FormatFuncConstructorEntry

	if entry, ok = formatters[value]; !ok {
		err = errors.BadRequestf(
			"unsupported format: %q. Available formats: %q",
			value,
			slices.Collect(maps.Keys(formatters)),
		)

		return
	}

	formatFlag.value = value
	formatFlag.description = entry.description
	formatFlag.constructor = entry.FormatFuncConstructor

	return
}

func (formatFlag *FormatFlag) MakeFormatFunc(
	repo *Repo,
	writer interfaces.WriterAndStringWriter,
) interfaces.FuncIter[*sku.Transacted] {
	return formatFlag.constructor(repo, writer)
}

var formatters = map[string]FormatFuncConstructorEntry{
	"tags-path": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.GetObjectId(),
					&object.Metadata.Cache.TagPaths,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	// TODO convert rest of the formatters from
	// src/november/local_working_copy/formatter_value.go in MakeFormatFunc to
	// this strategy. In the cases where the formatter construct should return an
	// error, instead call repo.Cancel(err) and then return
}
