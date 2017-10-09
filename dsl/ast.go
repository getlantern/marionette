package dsl

type Document struct {
	Model        *Model
	ActionBlocks []*ActionBlock
}

type Model struct {
	Connection   Pos
	Lparen       Pos
	Transport    string
	TransportPos Pos
	Comma        Pos
	Port         string
	PortPos      Pos
	Rparen       Pos
	Colon        Pos
	Transitions  []*Transition
}

type Transition struct {
	Source            string
	SourcePos         Pos
	Destination       string
	DestinationPos    Pos
	ActionBlock       string
	ActionBlockPos    Pos
	Probability       float64
	ProbabilityPos    Pos
	IsErrorTransition bool
}

type ActionBlock struct {
	Action  Pos
	Name    string
	NamePos Pos
	Colon   Pos
	Actions []*Action
}

type Action struct {
	Party     string
	PartyPos  Pos
	Name      string
	NamePos   Pos
	Dot       Pos
	Method    string
	MethodPos Pos
	Lparen    Pos
	Args      []*Arg
	Rparen    Pos
	If        Pos

	RegexMatchIncoming       Pos
	RegexMatchIncomingLparen Pos
	Regex                    string
	RegexPos                 Pos
	RegexMatchIncomingRparen Pos
}

type Arg struct {
	Value  interface{}
	Pos    Pos
	EndPos Pos
}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}
