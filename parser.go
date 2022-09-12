package vhs

// Parser is the structure that manages the parsing of tokens.
type Parser struct {
	l      *Lexer
	errors []ParserError
	cur    Token
	peek   Token
}

// NewParser returns a new Parser.
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, errors: []ParserError{}}

	// Read two tokens, so cur and peek are both set.
	p.nextToken()
	p.nextToken()

	return p
}

// Parse takes an input string provided by the lexer and parses it into a
// list of commands.
func (p *Parser) Parse() []Command {
	cmds := []Command{}

	for p.cur.Type != EOF {
		cmds = append(cmds, p.parseCommand())
		p.nextToken()
	}

	return cmds
}

// parseCommand parses a command.
func (p *Parser) parseCommand() Command {
	switch p.cur.Type {
	case SPACE, BACKSPACE, ENTER, ESCAPE, TAB, DOWN, LEFT, RIGHT, UP:
		return p.parseKeypress(p.cur.Type)
	case SET:
		return p.parseSet()
	case SLEEP:
		return p.parseSleep()
	case TYPE:
		return p.parseType()
	case CTRL:
		return p.parseCtrl()
	case START:
		return p.parseStart()
	case STOP:
		return p.parseStop()
	default:
		p.errors = append(p.errors, NewError(p.cur, "Invalid command: "+p.cur.Literal))
		return Command{Type: ILLEGAL}
	}
}

// parseSpeed parses a typing speed indication.
//
// i.e. @<time>
//
// This is optional (defaults to 100ms), thus skips (rather than error-ing)
// if the typing speed is not specified.
func (p *Parser) parseSpeed() string {
	if p.peek.Type == AT {
		p.nextToken()
		return p.parseTime()
	}

	return "100ms"
}

// parseRepeat parses an optional repeat count for a command.
//
// i.e. Backspace [count]
//
// This is optional (defaults to 1), thus skips (rather than error-ing)
// if the repeat count is not specified.
func (p *Parser) parseRepeat() string {
	if p.peek.Type == NUMBER {
		count := p.peek.Literal
		p.nextToken()
		return count
	}

	return "1"
}

// parseTime parses a time argument.
//
// i.e. <number>(s|ms)
//
func (p *Parser) parseTime() string {
	var t string

	if p.peek.Type == NUMBER {
		t = p.peek.Literal
		p.nextToken()
	} else {
		p.errors = append(p.errors, NewError(p.cur, "Expected time after "+p.cur.Literal))
	}

	if p.peek.Type == SECONDS || p.peek.Type == MILLISECONDS {
		t += p.peek.Literal
		p.nextToken()
	} else {
		t += "ms"
	}

	return t
}

// parseCtrl parses a control command.
// A control command takes a character to type while the modifier is held down.
//
// Ctrl+<character>
//
func (p *Parser) parseCtrl() Command {
	if p.peek.Type == PLUS {
		p.nextToken()
		if p.peek.Type == STRING {
			c := p.peek.Literal
			p.nextToken()
			return Command{Type: CTRL, Args: c}
		}
	}

	p.errors = append(p.errors, NewError(p.cur, "Expected control character, got "+p.cur.Literal))
	return Command{Type: CTRL}
}

// parseKeypress parses a repeatable and time adjustable keypress command.
// A keypress command takes an optional typing speed and optional count.
//
// Key[@<time>] [count]
//
func (p *Parser) parseKeypress(ct TokenType) Command {
	cmd := Command{Type: CommandType(ct)}
	cmd.Options = p.parseSpeed()
	cmd.Args = p.parseRepeat()
	return cmd
}

// parseSet parses a set command.
// A set command takes a setting name and a value.
//
// Set <setting> <value>
//
func (p *Parser) parseSet() Command {
	cmd := Command{Type: SET}

	if p.peek.Type == SETTING {
		cmd.Options = p.peek.Literal
	} else {
		p.errors = append(p.errors, NewError(p.peek, "Unknown setting: "+p.peek.Literal))
	}
	p.nextToken()

	cmd.Args = p.peek.Literal
	p.nextToken()

	// Allow Padding to have bare units (e.g. 10px, 5em, 10%)
	//
	// Set Padding 5em
	//
	if p.peek.Type == EM || p.peek.Type == PX || p.peek.Type == PERCENT {
		cmd.Args += p.peek.Literal
		p.nextToken()
	}

	return cmd
}

// parseSleep parses a sleep command.
// A sleep command takes a time for how long to sleep.
//
// Sleep <time>
//
func (p *Parser) parseSleep() Command {
	cmd := Command{Type: SLEEP}
	cmd.Args = p.parseTime()
	return cmd
}

// parseStart parses a start command.
// A start command takes no arguments.
//
// Start
//
func (p *Parser) parseStart() Command {
	cmd := Command{Type: START}
	return cmd
}

// parseStop parses a stop command.
// A start command takes no arguments.
//
// Stop
//
func (p *Parser) parseStop() Command {
	cmd := Command{Type: STOP}
	return cmd
}

// parseType parses a type command.
// A type command takes a string to type.
//
// Type "string"
//
func (p *Parser) parseType() Command {
	cmd := Command{Type: TYPE}

	cmd.Options = p.parseSpeed()

	if p.peek.Type != STRING {
		p.errors = append(p.errors, NewError(p.peek, p.cur.Literal+" expects string"))
	}

	for p.peek.Type == STRING {
		p.nextToken()
		cmd.Args += p.cur.Literal

		// If the next token is a string, add a space between them.
		// Since tokens must be separated by a whitespace, this is most likely
		// what the user intended.
		//
		// Although it is possible that there may be multiple spaces / tabs between
		// the tokens, however if the user was intending to type multiple spaces
		// they would need to use a string literal.

		if p.peek.Type == STRING {
			cmd.Args += " "
		}
	}

	return cmd
}

// Errors returns any errors that occurred during parsing.
func (p *Parser) Errors() []ParserError {
	return p.errors
}

// nextToken gets the next token from the lexer
// and updates the parser tokens accordingly.
func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}
