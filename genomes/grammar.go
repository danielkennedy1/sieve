package genomes

import (
	"strings"
)

type Grammar struct {
	Rules []Rule
}

// NOTE: I'm calling Rule Lefts "stem" in some places
type Rule struct {
	Left string
	Productions []Production
}

type Production struct {
	Elements []string
}

func isNonTerminal(element string) bool {
	if element[0] == '<' && element[len(element) - 1] == '>' {
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
	token string
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

func expand(gr Grammar, g Genotype, token string, offset *int) *GrammarNode {
	rule := gr.getRule(token)

	if rule == nil {
		return &GrammarNode{
			token: token,
			children: nil,
		}
	}

	*offset += 1

	var children []*GrammarNode

	production := rule.Productions[g.Genes[(*offset) % len(g.Genes)] % uint8(len(rule.Productions))]

	for _, e := range production.Elements {
		children = append(children, expand(gr, g, e, offset))
	}

	return &GrammarNode{
		token: rule.Left,
		children: children,
	}
}

func (g Genotype) MapToGrammar(gr Grammar) GrammarNode {
	offset := -1
	root := expand(gr, g, gr.Rules[0].Left, &offset)
	return *root
}
