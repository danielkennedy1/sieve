package genomes

func NewTestLectureExampleGrammar() Grammar {
	return Grammar{
		Rules: []Rule{
			{
				Left: "<expr>",
				Productions: []Production{
					{Elements: []string{"<expr>", "<op>", "<expr>"}},
					{Elements: []string{"<var>"}},
				},
			},
			{
				Left: "<op>",
				Productions: []Production{
					{Elements: []string{"+"}},
					{Elements: []string{"-"}},
					{Elements: []string{"*"}},
					{Elements: []string{"/"}},
				},
			},
			{
				Left: "<var>",
				Productions: []Production{
					{Elements: []string{"<prc>"}},
					{Elements: []string{"<input>"}},
				},
			},
			{
				Left: "<input>",
				Productions: []Production{
					{Elements: []string{"a"}},
					{Elements: []string{"b"}},
				},
			},
			{
				Left: "<prc>",
				Productions: []Production{
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
