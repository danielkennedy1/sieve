package genomes

import (
	"fmt"
	"math"
	"math/rand/v2"
)

type Expression interface {
	GetValue() float64
	String() string
	Compare(other Expression) bool
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

func (nt NonTerminal) String() string {
	return fmt.Sprintf("(%s %s %s)", nt.Left.String(), nt.Operator.String(), nt.Right.String())
}

func (nt NonTerminal) Compare(other Expression) bool {
	nt2, ok := other.(NonTerminal)
	if ok {
		return nt.Operator == nt2.Operator && nt.Left.Compare(nt2.Left) && nt.Right.Compare(nt2.Right)
	}
	return false
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

func (p Primitive) String() string {
	return fmt.Sprintf("%.2f", p.Value)
}

func (p Primitive) Compare(other Expression) bool {
	p2, ok := other.(Primitive)
	if ok {
		return p.Value == p2.Value
	}
	return false
}

func (v Variable) GetValue() float64 {
	return (*v.Variables)[v.Index]
}

func (v Variable) String() string {
	return fmt.Sprintf("x%d", v.Index)
}

func depth(e Expression) int {
	switch node := e.(type) {
	case Primitive, Variable:
		return 1
	case NonTerminal:
		depthl := depth(node.Left)
		depthr := depth(node.Right)
		if depthl > depthr {
			return depthl + 1
		}
		return depthr + 1
	default:
		panic("unknown node")
	}
}

func countNodes(e Expression) int {
	switch node := e.(type) {
	case Primitive, Variable:
		return 1
	case NonTerminal:
		return 1 + countNodes(node.Left) + countNodes(node.Right)
	default:
		panic("unknown node")
	}
}

func clone(e Expression) Expression {
	switch node := e.(type) {
	case Primitive:
		return Primitive{Value: node.Value}
	case Variable:
		return Variable{Variables: node.Variables, Index: node.Index}
	case NonTerminal:
		return NonTerminal{
			Operator: node.Operator,
			Left:     clone(node.Left),
			Right:    clone(node.Right),
		}
	default:
		panic("unknown node")
	}
}

type Path []int

func getAt(e Expression, path Path) Expression {
	if len(path) == 0 {
		return e
	}
	nt, ok := e.(NonTerminal)
	if !ok {
		return e
	}
	if path[0] == 0 {
		return getAt(nt.Left, path[1:])
	}
	return getAt(nt.Right, path[1:])
}

func setAt(e Expression, path Path, repl Expression) Expression {
	if len(path) == 0 {
		return clone(repl)
	}
	nt, ok := e.(NonTerminal)
	if !ok {
		return clone(repl)
	}
	if path[0] == 0 {
		return NonTerminal{
			Operator: nt.Operator,
			Left:     setAt(nt.Left, path[1:], repl),
			Right:    clone(nt.Right),
		}
	}
	return NonTerminal{
		Operator: nt.Operator,
		Left:     clone(nt.Left),
		Right:    setAt(nt.Right, path[1:], repl),
	}
}

func pickPath(e Expression, idx int) Path {
	switch node := e.(type) {
	case Primitive, Variable:
		return Path{}
	case NonTerminal:
		if idx == 0 {
			return Path{}
		}
		idx--
		leftCount := countNodes(node.Left)
		if idx < leftCount {
			return append(Path{0}, pickPath(node.Left, idx)...)
		}
		return append(Path{1}, pickPath(node.Right, idx-leftCount)...)
	default:
		panic("unknown node")
	}
}

func NewCrossoverExpression(rng *rand.Rand, maxDepth int) func(Expression, Expression) (Expression, Expression) {
	return func(p1, p2 Expression) (Expression, Expression) {
		return Crossover(p1, p2, rng, maxDepth)
	}
}

func Crossover(p1, p2 Expression, rng *rand.Rand, maxDepth int) (Expression, Expression) {
	clone1 := clone(p1)
	clone2 := clone(p2)

	nodes1 := countNodes(clone1)
	nodes2 := countNodes(clone2)
	if nodes1 == 0 || nodes2 == 0 {
		return clone1, clone2
	}

	point1 := rng.IntN(int(math.Ceil(float64(nodes1) / 3.0)))
	point2 := rng.IntN(int(math.Ceil(float64(nodes2) / 3.0)))
	// fmt.Println("Crossover points:", point1, point2)
	path1 := pickPath(clone1, point1)
	path2 := pickPath(clone2, point2)

	sub1 := getAt(clone1, path1)
	sub2 := getAt(clone2, path2)
	// fmt.Println(sub1.String(), "<->", sub2.String())

	child1 := setAt(clone1, path1, sub2)
	child2 := setAt(clone2, path2, sub1)

	if maxDepth > 0 {
		if depth(child1) > maxDepth {
			child1 = clone1
		}
		if depth(child2) > maxDepth {
			child2 = clone2
		}
	}

	return child1, child2
}

func (v Variable) Compare(other Expression) bool {
	v2, ok := other.(Variable)
	if ok {
		return v.Index == v2.Index
	}
	return false
}

func NewMutateExpression(constants []float64, rng *rand.Rand) func(e Expression) Expression {
	var MutateExpression func(e Expression) Expression

	MutateExpression = func(e Expression) Expression {
		switch x := e.(type) {
		case Primitive:
			x.Value = constants[rand.IntN(len(constants))]
			return x
		case Variable:
			x.Index = rand.IntN(len(*x.Variables))
			return x
		case NonTerminal:
			random := rand.Float64()
			if random < 0.1 { // FIXME: Shouldn't be hardcoded
				x.Operator = Operator(rand.IntN(int(numOperators)))
				return x
			} else if random < 0.55 {
				x.Left = MutateExpression(x.Left)
				return x
			} else {
				x.Right = MutateExpression(x.Right)
				return x
			}
		default:
			panic("Unexpected expression type in MutateExpression")
		}
	}
	return MutateExpression
}
