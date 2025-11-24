package genomes

import (
	"math/rand/v2"
	"sort"
	"strings"
	"time"
)

type Grammar struct {
	Rules   []Rule
	ruleMap map[string]*Rule
}

type Rule struct {
	Left        string
	Productions []Production
}

type Production struct {
	Elements []string
}

func isNonTerminal(element string) bool {
	return len(element) > 0 && element[0] == '<' && element[len(element)-1] == '>'
}

func ValidateGrammar(g Grammar) bool {
	stems := map[string]bool{}
	nonTerminalElements := map[string]bool{}

	for _, r := range g.Rules {
		if stems[r.Left] {
			return false
		}

		if !isNonTerminal(r.Left) {
			return false
		}

		stems[r.Left] = true

		for _, p := range r.Productions {
			for _, e := range p.Elements {
				if isNonTerminal(e) {
					nonTerminalElements[e] = true
				}
			}
		}
	}

	for e := range nonTerminalElements {
		if !stems[e] {
			return false
		}
	}

	return true
}

func (gr *Grammar) BuildRuleMap() {
	gr.ruleMap = make(map[string]*Rule, len(gr.Rules))
	for i := range gr.Rules {
		gr.ruleMap[gr.Rules[i].Left] = &gr.Rules[i]
	}
}

func (gr Grammar) getRule(token string) *Rule {
	if gr.ruleMap != nil {
		return gr.ruleMap[token]
	}
	for i := range gr.Rules {
		if gr.Rules[i].Left == token {
			return &gr.Rules[i]
		}
	}
	return nil
}

type GrammarNode struct {
	token    string
	children []*GrammarNode
}

func (node GrammarNode) String() string {
	if node.children == nil {
		return node.token
	}

	capacity := len(node.children) * 10
	var sb strings.Builder
	sb.Grow(capacity)

	for i, child := range node.children {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(child.String())
	}
	return sb.String()
}

type Genotype struct {
	Genes []uint8
}

func expand(gr Grammar, g Genotype, token string, offset *int, maxGenes int) *GrammarNode {
	if *offset >= maxGenes {
		return &GrammarNode{
			token:    token,
			children: nil,
		}
	}

	rule := gr.getRule(token)
	if rule == nil {
		return &GrammarNode{
			token:    token,
			children: nil,
		}
	}

	*offset++

	production := rule.Productions[g.Genes[*offset%len(g.Genes)]%uint8(len(rule.Productions))]

	children := make([]*GrammarNode, len(production.Elements))
	for i, e := range production.Elements {
		children[i] = expand(gr, g, e, offset, maxGenes)
	}

	return &GrammarNode{
		token:    rule.Left,
		children: children,
	}
}

func (g Genotype) MapToGrammar(gr Grammar, maxGenes int) GrammarNode {
	offset := -1
	root := expand(gr, g, gr.Rules[0].Left, &offset, maxGenes)
	return *root
}

func cloneG(g Genotype) Genotype {
	newGenes := make([]uint8, len(g.Genes))
	copy(newGenes, g.Genes)
	return Genotype{Genes: newGenes}
}

func NewCrossoverGenotype(rng *rand.Rand) func(g1, g2 Genotype) (Genotype, Genotype) {
	return func(g1, g2 Genotype) (Genotype, Genotype) {
		return g1.CrossoverGenotype(g2, rng)
	}
}

func (g Genotype) CrossoverGenotype(g2 Genotype, rng *rand.Rand) (Genotype, Genotype) {
	if len(g.Genes) == 0 || len(g2.Genes) == 0 {
		return g, g2
	}

	clone1 := cloneG(g)
	clone2 := cloneG(g2)

	crossPoint1 := rng.IntN(len(clone1.Genes))
	crossPoint2 := rng.IntN(len(clone2.Genes))

	tail1 := make([]uint8, len(clone1.Genes[crossPoint1:]))
	copy(tail1, clone1.Genes[crossPoint1:])

	tail2 := make([]uint8, len(clone2.Genes[crossPoint2:]))
	copy(tail2, clone2.Genes[crossPoint2:])

	clone1.Genes = append(clone1.Genes[:crossPoint1], tail2...)
	clone2.Genes = append(clone2.Genes[:crossPoint2], tail1...)

	return clone1, clone2
}

func NewMutateGenotype(mutationRate float64) func(g Genotype) Genotype {
	return func(g Genotype) Genotype {
		rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
		clone := cloneG(g)
		for i := 0; i < len(clone.Genes); i++ {
			if rng.Float64() < mutationRate {
				clone.Genes[i] = uint8(rng.IntN(256))
			}
		}
		return clone
	}
}

func ExtractInputVariables(gr Grammar) []string {
	var inputs []string

	for _, rule := range gr.Rules {
		if rule.Left == "<input>" {
			for _, prod := range rule.Productions {
				for _, elem := range prod.Elements {
					// terminal: not <nonterminal>
					if len(elem) > 0 && elem[0] != '<' {
						inputs = append(inputs, elem)
					}
				}
			}
		}
	}
	return inputs
}

func BuildVarMapFromGrammar(gr Grammar) map[string]int {
	vars := ExtractInputVariables(gr)

	sort.Strings(vars)
	m := make(map[string]int, len(vars))
	for i, name := range vars {
		m[name] = i
	}
	return m
}

func NewCreateGenotype(length int, rng *rand.Rand) func() Genotype {
	return func() Genotype {
		genes := make([]uint8, length)
		for i := 0; i < length; i++ {
			genes[i] = uint8(rng.IntN(256))
		}
		return Genotype{Genes: genes}
	}
}
