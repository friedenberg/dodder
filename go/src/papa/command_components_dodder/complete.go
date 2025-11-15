package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/queries"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type Complete struct {
	ObjectMetadata
	Query
}

func (cmd Complete) GetFlagValueMetadataTags(
	metadata *object_metadata.Metadata,
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
	metadata *object_metadata.Metadata,
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
	proto.SetFlagSetDescription(
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
