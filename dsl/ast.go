package dsl

// Node represents a node within the AST.
type Node interface {
	node()
}

func (*Document) node()    {}
func (*Model) node()       {}
func (*Transition) node()  {}
func (*ActionBlock) node() {}
func (*Action) node()      {}
func (*Arg) node()         {}
func (*Pos) node()         {}

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

// Walk traverses an AST in depth-first order.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk children.
	switch node := node.(type) {
	case *Document:
		Walk(v, node.Model)
		for _, blk := range node.ActionBlocks {
			Walk(v, blk)
		}

	case *Model:
		for _, transition := range node.Transitions {
			Walk(v, transition)
		}

	case *ActionBlock:
		for _, action := range node.Actions {
			Walk(v, action)
		}

	case *Action:
		for _, arg := range node.Args {
			Walk(v, arg)
		}
	}

	v.Visit(nil)
}

// Visitor represents an object for iterating over nodes using Walk().
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// VisitorFunc implements a type to use a function as a Visitor.
type VisitorFunc func(node Node)

func (fn VisitorFunc) Visit(node Node) Visitor {
	fn(node)
	return fn
}
