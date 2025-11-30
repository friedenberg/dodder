package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/env_local"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/november/sku_fmt"
	pkg_query "code.linenisgreat.com/dodder/go/src/oscar/queries"
	"code.linenisgreat.com/dodder/go/src/victor/local_working_copy"
)

type Complete struct {
	ObjectMetadata
	Query
}

func (cmd Complete) GetFlagValueMetadataTags(
	metadata object_metadata.MetadataMutable,
) interfaces.FlagValue {
	return command.FlagValueCompleter{
		FlagValue: cmd.ObjectMetadata.GetFlagValueMetadataTags(metadata),
		FuncCompleter: func(
			req command.Request,
			envLocal env_local.Env,
			commandLine command.CommandLine,
		) {
			local := LocalWorkingCopy{}.MakeLocalWorkingCopy(req)

			cmd.CompleteObjects(
				req,
				local,
				pkg_query.BuilderOptionDefaultGenres(genres.Tag),
			)
		},
	}
}

func (cmd Complete) GetFlagValueStringTags(
	value *values.String,
) interfaces.FlagValue {
	return command.FlagValueCompleter{
		FlagValue: value,
		FuncCompleter: func(
			req command.Request,
			envLocal env_local.Env,
			commandLine command.CommandLine,
		) {
			local := LocalWorkingCopy{}.MakeLocalWorkingCopy(req)

			cmd.CompleteObjects(
				req,
				local,
				pkg_query.BuilderOptionDefaultGenres(genres.Tag),
			)
		},
	}
}

func (cmd Complete) GetFlagValueMetadataType(
	metadata object_metadata.MetadataMutable,
) interfaces.FlagValue {
	return command.FlagValueCompleter{
		FlagValue: cmd.ObjectMetadata.GetFlagValueMetadataType(metadata),
		FuncCompleter: func(
			req command.Request,
			envLocal env_local.Env,
			commandLine command.CommandLine,
		) {
			local := LocalWorkingCopy{}.MakeLocalWorkingCopy(req)

			cmd.CompleteObjects(
				req,
				local,
				pkg_query.BuilderOptionDefaultGenres(genres.Type),
			)
		},
	}
}

func (cmd Complete) SetFlagsProto(
	proto *sku.Proto,
	flagSet interfaces.CLIFlagDefinitions, descriptionUsage string,
	tagUsage string,
	typeUsage string,
) {
	proto.Metadata.SetFlagSetDescription(
		flagSet,
		descriptionUsage,
	)

	flagSet.Var(
		cmd.GetFlagValueMetadataTags(&proto.Metadata),
		"tags",
		tagUsage,
	)

	flagSet.Var(
		cmd.GetFlagValueMetadataType(&proto.Metadata),
		"type",
		typeUsage,
	)
}

func (cmd Complete) CompleteObjectsIncludingWorkspace(
	req command.Request,
	local *local_working_copy.Repo,
	queryBuilderOptions pkg_query.BuilderOption,
	args ...string,
) {
	sku_fmt.OutputCliCompletions(
		local.GetEnvRepo().Env,
		local.GetAbbr().GetSeenIds(),
	)
	// printerCompletions := sku_fmt.MakePrinterComplete(local)

	// query := cmd.MakeQueryIncludingWorkspace(
	// 	req,
	// 	queryBuilderOptions,
	// 	local,
	// 	args,
	// )

	// if err := local.GetStore().QueryTransacted(
	// 	query,
	// 	printerCompletions.PrintOne,
	// ); err != nil {
	// 	local.Cancel(err)
	// }
}

func (cmd Complete) CompleteObjects(
	req command.Request,
	local *local_working_copy.Repo,
	queryBuilderOptions pkg_query.BuilderOption,
	args ...string,
) {
	sku_fmt.OutputCliCompletions(
		local.GetEnvRepo().Env,
		local.GetAbbr().GetSeenIds(),
	)
	// printerCompletions := sku_fmt.MakePrinterComplete(local)

	// query := cmd.MakeQuery(
	// 	req,
	// 	queryBuilderOptions,
	// 	local,
	// 	args,
	// )

	// if err := local.GetStore().QueryTransacted(
	// 	query,
	// 	printerCompletions.PrintOne,
	// ); err != nil {
	// 	local.Cancel(err)
	// }
}
