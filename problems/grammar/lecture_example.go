package grammar

import (
	"github.com/danielkennedy1/sieve/genomes"
)

func NewLectureExampleGrammar() genomes.Grammar {
	return genomes.Grammar{
		Rules: []genomes.Rule{
			{
				Left: "<expr>",
				Productions: []genomes.Production{
					{Elements: []string{"<expr>", "<op>", "<expr>"}},
					{Elements: []string{"<var>"}},
				},
			},
			{
				Left: "<op>",
				Productions: []genomes.Production{
					{Elements: []string{"+"}},
					{Elements: []string{"-"}},
					{Elements: []string{"*"}},
					{Elements: []string{"/"}},
				},
			},
			{
				Left: "<var>",
				Productions: []genomes.Production{
					{Elements: []string{"<prc>"}},
					{Elements: []string{"<input>"}},
				},
			},
			{
				Left: "<input>",
				Productions: []genomes.Production{
					{Elements: []string{"a"}},
					{Elements: []string{"b"}},
				},
			},
			{
				Left: "<prc>",
				Productions: []genomes.Production{
					{Elements: []string{"0.0"}},
					{Elements: []string{"0.1"}},
					{Elements: []string{"0.2"}},
					{Elements: []string{"0.3"}},
					{Elements: []string{"0.4"}},
					{Elements: []string{"0.5"}},
				},
			},
		},
	}
}
