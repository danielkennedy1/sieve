package genomes

type Expression interface {
	GetValue() float64
}

type NonTerminal struct {
	Operator Operator
	Left Expression
	Right Expression
}

type Operator int

const (
    Add Operator = iota
    Subtract
    Multiply
    Divide
)

type Primitive struct {
	Value float64
}

type Variable struct {
	Variables *[]float64
	Index int
}

// TODO: error handling
func (nt NonTerminal) GetValue() float64 {
    switch nt.Operator {
		case Add:
			return nt.Left.GetValue() + nt.Right.GetValue()
		case Subtract:
			return nt.Left.GetValue() - nt.Right.GetValue()
		case Multiply:
			return nt.Left.GetValue() * nt.Right.GetValue()
		case Divide:
			return nt.Left.GetValue() / nt.Right.GetValue()
		default:
			panic("invalid operator")
	}
}

func (p Primitive) GetValue() float64 {
	return p.Value
}

func (v Variable) GetValue() float64 {
	return (*v.Variables)[v.Index]
}
