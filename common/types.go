package common

type Action struct {
	Name string
	Cmd  string
	Args string
}

type Rule struct {
	Name    string
	Path    string
	Limit   int
	Actions []Action
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
