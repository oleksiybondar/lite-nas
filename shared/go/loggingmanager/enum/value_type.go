package enum

// ValueType defines typed occurrence value kinds.
type ValueType string

const (
	ValueTypeInt   ValueType = "int"
	ValueTypeFloat ValueType = "float"
	ValueTypeText  ValueType = "text"
	ValueTypeBool  ValueType = "bool"
)
