package common

type Rule struct {
	Name    string
	Path    string
	Greater int
}

type Definition struct {
	Name     string
	Protocol string
	Host     string
	Port     int
	User     string
	Password string
	Rules    []Rule
}
