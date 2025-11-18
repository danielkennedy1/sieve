package grammar

import (
	"bufio"
	"strings"

	"github.com/danielkennedy1/sieve/genomes"
)

func Parse(scanner bufio.Scanner) genomes.Grammar {
	var grammar genomes.Grammar
	for scanner.Scan() {
		sides := strings.Split(scanner.Text(), "::=") 
		left := strings.Trim(sides[0], " ")
		right := strings.Split(sides[1], "|")
		var productions []genomes.Production
		for _, s := range right {
			s = strings.Trim(s, " ")
			var elements []string
			for e := range strings.SplitSeq(s, " ") {
				elements = append(elements, e)
			}
			productions = append(productions, genomes.Production{Elements: elements})
		}
		grammar.Rules = append(grammar.Rules, genomes.Rule{Left: left, Productions: productions})
	}
	return grammar
}
