package genomes

import (
	"maps"
	"math/rand/v2"
	"sort"
	"strings"
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
	Genes      []uint8
	Attributes map[string]any
}

func (gr Grammar) getTerminatingProductionIndex(rule *Rule) int {
	bestIndex := 0
	minRecursiveRefs := 99999
	minTotalNonTerminals := 99999

	for i, prod := range rule.Productions {
		recursiveRefs := 0
		totalNonTerminals := 0

		for _, e := range prod.Elements {
			if isNonTerminal(e) {
				totalNonTerminals++

				if e == rule.Left {
					recursiveRefs++
				}
			}
		}

		if recursiveRefs < minRecursiveRefs {
			minRecursiveRefs = recursiveRefs
			minTotalNonTerminals = totalNonTerminals
			bestIndex = i
		} else if recursiveRefs == minRecursiveRefs {
			if totalNonTerminals < minTotalNonTerminals {
				minTotalNonTerminals = totalNonTerminals
				bestIndex = i
			}
		}

	}

	return bestIndex
}

func expand(gr Grammar, g Genotype, token string, offset *int, maxGenes int) *GrammarNode {
	rule := gr.getRule(token)
	if rule == nil {
		return &GrammarNode{
			token:    token,
			children: nil,
		}
	}

	var production Production

	if *offset >= maxGenes {
		bestIdx := gr.getTerminatingProductionIndex(rule)
		production = rule.Productions[bestIdx]
	} else {
		*offset += 1

		geneIdx := (*offset) % len(g.Genes)
		prodIdx := int(g.Genes[geneIdx]) % len(rule.Productions)

		production = rule.Productions[prodIdx]
	}

	var children []*GrammarNode
	for _, e := range production.Elements {
		children = append(children, expand(gr, g, e, offset, maxGenes))
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
	clone1 := cloneG(g)
	clone2 := cloneG(g2)

	if len(clone1.Genes) == 0 || len(clone2.Genes) == 0 {
		return clone1, clone2
	}

	minLen := len(clone1.Genes)
	if len(clone2.Genes) < minLen {
		minLen = len(clone2.Genes)
	}

	// Single crossover point that works for both genotypes
	crossPoint := rng.IntN(minLen)

	// Swap the tails after the crossover point
	for i := crossPoint; i < minLen; i++ {
		clone1.Genes[i], clone2.Genes[i] = clone2.Genes[i], clone1.Genes[i]
	}

	return clone1, clone2
}

func NewMutateGenotype(rng *rand.Rand, perGeneMutationRate float64) func(g Genotype) Genotype {
    return func(g Genotype) Genotype {
        clone := cloneG(g)
        for i := range clone.Genes {
            if rng.Float64() < perGeneMutationRate {
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

func NewCreateGenotype(length int, rng *rand.Rand, options ...map[string]any) func() Genotype {

	var universalAttributes map[string]any
	if len(options) > 0 {
		universalAttributes = options[0]
	}

	id := 0
	return func() Genotype {
		attrs := make(map[string]any, len(universalAttributes))

		maps.Copy(attrs, universalAttributes)
		attrs["id"] = int(id)
		id++
		genes := make([]uint8, length)
		for i := range length {
			genes[i] = uint8(rng.IntN(256))
		}
		return Genotype{Genes: genes, Attributes: attrs}
	}
}
