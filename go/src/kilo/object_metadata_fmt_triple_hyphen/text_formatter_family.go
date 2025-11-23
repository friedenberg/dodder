package object_metadata_fmt_triple_hyphen

type TextFormatterFamily struct {
	BlobPath     TextFormatter
	InlineBlob   TextFormatter
	MetadataOnly TextFormatter
	BlobOnly     TextFormatter
}

type FormatterDependencies struct{}

func MakeTextFormatterFamily(
	common Dependencies,
) TextFormatterFamily {
	return TextFormatterFamily{
		BlobPath:     MakeTextFormatterMetadataBlobPath(common),
		InlineBlob:   MakeTextFormatterMetadataInlineBlob(common),
		MetadataOnly: MakeTextFormatterMetadataOnly(common),
		BlobOnly:     MakeTextFormatterExcludeMetadata(common),
	}
}
