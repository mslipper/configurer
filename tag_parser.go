package configurer

import "errors"

type token struct {
	typ int
	val string
}

const (
	assignToken = iota
	sepToken
	charToken
	eofToken
)

func parseTag(in string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tokens := new(tagTokenizer).Tokenize(in)

	var state int
	var tuples [][2]string
	for _, tok := range tokens {
		switch state {
		case 0:
			if tok.typ != charToken {
				return nil, errors.New("expected a label")
			}
			tuples = append(tuples, [2]string{
				tok.val,
				"",
			})
			state = 1
		case 1:
			if tok.typ == sepToken {
				state = 0
				continue
			}
			if tok.typ == assignToken {
				state = 2
				continue
			}
			if tok.typ == eofToken {
				state = -1
				continue
			}
			return nil, errors.New("expected an assignment or a separator")
		case 2:
			if tok.typ == charToken {
				tuples[len(tuples)-1][1] = tok.val
				state = 3
				continue
			}
			return nil, errors.New("expected a label")
		case 3:
			if tok.typ == eofToken {
				state = -1
				continue
			}
			if tok.typ == sepToken {
				state = 0
				continue
			}
			return nil, errors.New("expected EOF or a separator")
		default:
			panic("unknown parse state")
		}
	}

	out := make(map[string]string)
	for _, tuple := range tuples {
		out[tuple[0]] = tuple[1]
	}
	return out, nil
}

type tagTokenizer struct {
	escaping bool
}

func (t *tagTokenizer) Tokenize(in string) []*token {
	var tokens []*token
	for i := 0; i < len(in); i++ {
		var tok *token
		chr := in[i]
		switch chr {
		case '=':
			tok = t.assign()
		case ',':
			tok = t.sep()
		case '\\':
			tok = t.escape()
		default:
			tok = t.char(chr)
		}

		if tok == nil {
			continue
		}

		if len(tokens) > 0 {
			lastToken := tokens[len(tokens)-1]
			if lastToken.typ == charToken && tok.typ == charToken {
				lastToken.val = lastToken.val + tok.val
				continue
			}
		}

		tokens = append(tokens, tok)
	}
	return append(tokens, t.end())
}

func (t *tagTokenizer) assign() *token {
	if t.escaping {
		return t.char('=')
	}

	return &token{
		typ: assignToken,
	}
}

func (t *tagTokenizer) sep() *token {
	if t.escaping {
		return t.char(',')
	}
	return &token{
		typ: sepToken,
	}
}

func (t *tagTokenizer) escape() *token {
	if t.escaping {
		return t.char('\\')
	}
	t.escaping = true
	return nil
}

func (t *tagTokenizer) char(chr byte) *token {
	t.escaping = false
	return &token{
		typ: charToken,
		val: string(chr),
	}
}

func (t *tagTokenizer) end() *token {
	return &token{
		typ: eofToken,
	}
}
