package object_metadata_fmt_triple_hyphen

type FormatterFamily struct {
	BlobPath     Formatter
	InlineBlob   Formatter
	MetadataOnly Formatter
	BlobOnly     Formatter
}

func MakeFormatterFamily(
	common Dependencies,
) FormatterFamily {
	return FormatterFamily{
		BlobPath:     MakeFormatterMetadataBlobPath(common),
		InlineBlob:   MakeFormatterMetadataInlineBlob(common),
		MetadataOnly: MakeFormatterMetadataOnly(common),
		BlobOnly:     MakeFormatterExcludeMetadata(common),
	}
}
