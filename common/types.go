package common

import "time"

type Action struct {
	Description string
	Cmd         string
	Args        []string
}

type Request struct {
	Method string
	Path   string
}

type Rule struct {
	ID          string
	Description string
	Request     Request
	Evaluator   string
	Delay       time.Duration
	Actions     []Action
}

type Server struct {
	Description string
	Protocol    string
	Host        string
	Port        int
	User        string
	Password    string
	Rules       []Rule
}
