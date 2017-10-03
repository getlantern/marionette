package marionette

import (
	"regexp"
)

type Action struct {
	name_                 string
	party_                string
	module_               string
	method_               string
	args_                 []string
	regex_match_incoming_ *regexp.Regexp
}

func NewAction(name, party, module, method string, args []string, regex *regexp.Regexp) *Action {
	return &Action{
		name_:   name,
		party_:  party,
		module_: module,
		method_: method,
		args_:   args,
		regex_match_incoming_: regex,
	}
}

func (a *Action) Arg(i int) string {
	if i >= len(a.args_) {
		return ""
	}
	return a.args_[i]
}

func (a *Action) execute(party, name string) bool {
	return a.party_ == party && a.name_ == name
}
