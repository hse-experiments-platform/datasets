package models

type ColumnType uint8

const (
	ColumnTypeString ColumnType = iota
	ColumnTypeInteger
	ColumnTypeFloat
	ColumnTypeCategorial
	ColumnTypeDropped
)

var TypeToString = map[ColumnType]string{
	ColumnTypeString:     "string",
	ColumnTypeInteger:    "int",
	ColumnTypeFloat:      "float",
	ColumnTypeCategorial: "categorial",
	ColumnTypeDropped:    "dropped",
}

type Schema []Column

type Column struct {
	Name string
	Type ColumnType
}
