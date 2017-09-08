package marionette

import (
	"regexp"
)

type Action struct {
	name_                 string
	party_                string
	module_               string
	method_               string
	args_                 []interface{}
	regex_match_incoming_ *regexp.Regexp
}

func NewAction(name, party, module, method string, args []interface{}, regex *regexp.Regexp) *Action {
	return &Action{
		name_:   name,
		party_:  party,
		module_: module,
		method_: method,
		args_:   args,
		regex_match_incoming_: regex,
	}
}

func (a *Action) execute(party, name string) bool {
	return a.party_ == party && a.name_ == name
}
