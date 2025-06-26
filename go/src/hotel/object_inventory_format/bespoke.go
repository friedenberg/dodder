package object_inventory_format

type bespoke struct {
	Formatter
	Parser
}

func MakeBespoke(f Formatter, p Parser) Format {
	return bespoke{
		Formatter: f,
		Parser:    p,
	}
}
