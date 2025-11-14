package genomes

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
