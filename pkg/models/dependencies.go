package models

type ReverseDependency struct {
	Id                       string `pg:",pk"`
	Atom                     string
	Type                     string
	ReverseDependencyAtom    string
	ReverseDependencyVersion string
	Condition                string
}
