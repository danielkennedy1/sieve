package genomes

import (
	"fmt"
)

type Expression interface {
	GetValue() float64
	String() string
}

type NonTerminal struct {
	Operator Operator
	Left     Expression
	Right    Expression
}

type Operator int

const (
	Add Operator = iota
	Subtract
	Multiply
	Divide
	numOperators
)

type Primitive struct {
	Value float64
}

type Variable struct {
	Variables *[]float64
	Index     int
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
		if nt.Right.GetValue() == 0 {
			return 1
		}
		return nt.Left.GetValue() / nt.Right.GetValue()
	default:
		panic("invalid operator")
	}
}

func (op Operator) String() string {
	switch op {
	case Add:
		return "+"
	case Subtract:
		return "-"
	case Multiply:
		return "*"
	case Divide:
		return "/"
	default:
		return "?"
	}
}

func (p Primitive) GetValue() float64 {
	return p.Value
}

func (v Variable) GetValue() float64 {
	return (*v.Variables)[v.Index]
}

func (nt NonTerminal) String() string {
	return fmt.Sprintf("(%s %s %s)", nt.Left.String(), nt.Operator.String(), nt.Right.String())
}

func (p Primitive) String() string {
	return fmt.Sprintf("%.2f", p.Value)
}

func (v Variable) String() string {
	return fmt.Sprintf("x%d", v.Index)
}
