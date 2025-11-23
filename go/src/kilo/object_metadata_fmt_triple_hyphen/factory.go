package object_metadata_fmt_triple_hyphen

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
)

type Factory struct {
	EnvDir        env_dir.Env
	BlobStore     interfaces.BlobStore
	BlobFormatter script_config.RemoteScript
}

func (factory Factory) getBlobDigestType() interfaces.FormatHash {
	hashType := factory.BlobStore.GetDefaultHashType()

	if hashType == nil {
		panic("no hash type set")
	}

	return hashType
}

func (factory Factory) MakeFormatterFamily() FormatterFamily {
	return FormatterFamily{
		BlobPath:     factory.makeFormatterMetadataBlobPath(),
		InlineBlob:   factory.makeFormatterMetadataInlineBlob(),
		MetadataOnly: factory.makeFormatterMetadataOnly(),
		BlobOnly:     factory.makeFormatterExcludeMetadata(),
	}
}

func (factory Factory) MakeTextParser() Parser {
	if factory.BlobStore == nil {
		panic("nil BlobWriterFactory")
	}

	return textParser{
		hashType:      factory.getBlobDigestType(),
		blobWriter:    factory.BlobStore,
		blobFormatter: factory.BlobFormatter,
	}
}

func (factory Factory) Make() Format {
	return Format{
		Parser:          factory.MakeTextParser(),
		FormatterFamily: factory.MakeFormatterFamily(),
	}
}

func (factory Factory) makeFormatterMetadataBlobPath() formatter {
	return formatter{
		factory.writeBoundary,
		factory.writeCommonMetadataFormat,
		factory.writeBlobPath,
		factory.writeTypeAndSigIfNecessary,
		factory.writeComments,
		factory.writeBoundary,
	}
}

func (factory Factory) makeFormatterMetadataOnly() formatter {
	return formatter{
		factory.writeBoundary,
		factory.writeCommonMetadataFormat,
		factory.writeBlobDigest,
		factory.writeTypeAndSigIfNecessary,
		factory.writeComments,
		factory.writeBoundary,
	}
}

func (factory Factory) makeFormatterMetadataInlineBlob() formatter {
	return formatter{
		factory.writeBoundary,
		factory.writeCommonMetadataFormat,
		factory.writeTypeAndSigIfNecessary,
		factory.writeComments,
		factory.writeBoundary,
		factory.writeNewLine,
		factory.writeBlob,
	}
}
