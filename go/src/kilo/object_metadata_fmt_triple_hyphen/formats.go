package object_metadata_fmt_triple_hyphen

type Format struct {
	FormatterFamily
	Parser
}

func Make(
	common Dependencies,
) Format {
	return Format{
		Parser:          MakeTextParser(common),
		FormatterFamily: MakeFormatterFamily(common),
	}
}
