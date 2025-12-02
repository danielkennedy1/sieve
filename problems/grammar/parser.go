package grammar

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"github.com/danielkennedy1/sieve/genomes"
)

func Parse(scanner bufio.Scanner) genomes.Grammar {
	var grammar genomes.Grammar

	numberRange := regexp.MustCompile(`^(\d+)\.\.(\d+)$`) 

	for scanner.Scan() {
		sides := strings.Split(scanner.Text(), "::=") 
		left := strings.Trim(sides[0], " ")
		right := strings.Split(sides[1], "|")
		var productions []genomes.Production
		for _, s := range right {
			s = strings.Trim(s, " ")

			matches := numberRange.FindStringSubmatch(s)

			if matches != nil {

				start, _ := strconv.Atoi(matches[1])
				end, _ := strconv.Atoi(matches[2])

				for i := start; i < end; i++ {
					productions = append(productions, genomes.Production{Elements: []string{strconv.Itoa(i)}})
				}
				continue
			}

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
