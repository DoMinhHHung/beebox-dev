package fielddefinition

type FieldKind string

const (
	FieldKindString  FieldKind = "STRING"
	FieldKindNumber  FieldKind = "NUMBER"
	FieldKindBoolean FieldKind = "BOOLEAN"
)

func (k FieldKind) Valid() bool {
	switch k {
	case FieldKindString, FieldKindNumber, FieldKindBoolean:
		return true
	default:
		return false
	}
}
