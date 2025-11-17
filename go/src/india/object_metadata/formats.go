package object_metadata

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
