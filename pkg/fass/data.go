package fass

// Course represents a course with its exercise sheets and registered users.
type Course struct {
	Identifier string `json:"-"`
	Name       string
	URL        string
	Exercises  []Exercise `json:"-"`
	Users      []Token
}

// Exercise represents an excerise sheet.
type Exercise struct {
	Identifier string `json:"-"`
}
