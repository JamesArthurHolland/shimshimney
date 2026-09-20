package k8s

type IntOrString struct {
	Type    Type
	IntVal  int
	StrVal  string
}

type Type int

const (
	Int Type = iota
	String
)

func (i IntOrString) String() string {
	if i.Type == String {
		return i.StrVal
	}
	return ""
}
