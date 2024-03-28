package models

type ColumnType uint8

const (
	ColumnTypeUndefined ColumnType = iota
	ColumnTypeString
	ColumnTypeInteger
	ColumnTypeFloat
	ColumnTypeCategorial
	ColumnTypeDropped
)

var TypeToString = map[ColumnType]string{
	ColumnTypeUndefined:  "undefined",
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
