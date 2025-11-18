package genomes

import (
	"math/rand/v2"
	"strings"
)

type Grammar struct {
	Rules []Rule
}

// NOTE: I'm calling Rule Lefts "stem" in some places
type Rule struct {
	Left        string
	Productions []Production
}

type Production struct {
	Elements []string
}

func isNonTerminal(element string) bool {
	if element[0] == '<' && element[len(element)-1] == '>' {
		return true
	}
	return false
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

func (gr Grammar) getRule(token string) *Rule {
	for _, r := range gr.Rules {
		if r.Left == token {
			return &r
		}
	}
	return nil
}

// FIXME: Currently relies on the assumption that a terminal statement only ever has one element in its productions

type GrammarNode struct {
	token    string
	children []*GrammarNode
}

func (node GrammarNode) String() string {
	if node.children == nil {
		return node.token
	}
	var childStrings []string
	for _, child := range node.children {
		childStrings = append(childStrings, child.String())
	}
	return strings.Join(childStrings, " ")
}

type Genotype struct {
	Genes []uint8
}

func expand(gr Grammar, g Genotype, token string, offset *int, depth int, maxDepth int, maxGenes int) *GrammarNode {

	if depth >= maxDepth || *offset >= maxGenes {
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

	*offset += 1

	var children []*GrammarNode

	production := rule.Productions[g.Genes[(*offset)%len(g.Genes)]%uint8(len(rule.Productions))]

	for _, e := range production.Elements {
		children = append(children, expand(gr, g, e, offset, depth+1, maxDepth, maxGenes))
	}

	return &GrammarNode{
		token:    rule.Left,
		children: children,
	}
}

func (g Genotype) MapToGrammar(gr Grammar, maxDepth int, maxGenes int) GrammarNode {
	offset := -1
	root := expand(gr, g, gr.Rules[0].Left, &offset, 0, maxDepth, maxGenes)
	return *root
}

func cloneG(g Genotype) Genotype {
	newGenes := make([]uint8, len(g.Genes))
	copy(newGenes, g.Genes)
	return Genotype{Genes: newGenes}
}

func (g Genotype) CrossoverGenotype(g1, g2 Genotype, rng *rand.Rand) (Genotype, Genotype) {

	clone1 := cloneG(g1)
	clone2 := cloneG(g2)

	if len(clone1.Genes) == 0 || len(clone2.Genes) == 0 {
		return g1, g2
	}

	crossPoint1 := rng.IntN(len(clone1.Genes))
	crossPoint2 := rng.IntN(len(clone2.Genes))

	clone1.Genes = append(clone1.Genes[:crossPoint1], clone2.Genes[crossPoint2:]...)
	clone2.Genes = append(clone2.Genes[:crossPoint2], clone1.Genes[crossPoint1:]...)

	return clone1, clone2
}

func (g Genotype) MutateGenotype(rng *rand.Rand, mutationRate float64) Genotype {
	clone := cloneG(g)

	for i := 0; i < len(clone.Genes); i++ {
		if rng.Float64() < mutationRate {
			clone.Genes[i] = uint8(rng.IntN(256))
		}
	}

	// fmt.Println(clone.Genes)

	return clone
}
