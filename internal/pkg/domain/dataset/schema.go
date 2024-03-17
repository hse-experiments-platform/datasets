package dataset

type CSVType uint8

const (
	CSVTypeString CSVType = iota
	CSVTypeInteger
	CSVTypeFloat
	CSVTypeBool
	CSVTypeAny
)

type Schema []Column

type Column struct {
	Name string
	Type CSVType
}
