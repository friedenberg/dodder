package object_metadata_fmt_triple_hyphen

type TextFormat struct {
	TextFormatterFamily
	TextParser
}

func MakeTextFormat(
	common Dependencies,
) TextFormat {
	return TextFormat{
		TextParser:          MakeTextParser(common),
		TextFormatterFamily: MakeTextFormatterFamily(common),
	}
}
