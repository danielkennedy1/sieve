package grammar_test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

func TestParser(t *testing.T) {
	want := genomes.NewTestLectureExampleGrammar()

	s := 
`<expr> ::= <expr> <op> <expr> | <var>
<op> ::= + | - | * | /
<var> ::= <prc> | <input>
<input> ::= a | b
<prc> ::= 0.0 | 0.1 | 0.2 | 0.3 | 0.4 | 0.5`
	scanner := bufio.NewScanner(strings.NewReader(s))

	got := grammar.Parse(*scanner)

	fmt.Println(want)
	fmt.Println(got)
}
