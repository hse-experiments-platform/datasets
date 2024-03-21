package dataset

type CSVType uint8

const (
	CSVTypeString CSVType = iota
	CSVTypeInteger
	CSVTypeFloat
	CSVTypeBool
	CSVTypeAny
)

var TypeToString = map[CSVType]string{
	CSVTypeString:  "string",
	CSVTypeInteger: "int",
	CSVTypeFloat:   "float",
	CSVTypeBool:    "bool",
	CSVTypeAny:     "string",
}

type Schema []Column

type Column struct {
	Name string
	Type CSVType
}
