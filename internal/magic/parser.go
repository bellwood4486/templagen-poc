package magic

import (
	"fmt"
	"strings"
	"unicode"
)

// parseType parses a type string into TypeExpr
func parseType(s string) (TypeExpr, error) {
	p := &typeParser{input: s, pos: 0}
	return p.parseType()
}

type typeParser struct {
	input string
	pos   int
}

func (p *typeParser) parseType() (TypeExpr, error) {
	p.skipWhitespace()

	// Check for slice
	if strings.HasPrefix(p.remaining(), "[]") {
		p.pos += 2
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindSlice, Elem: &elem}, nil
	}

	// Check for map
	if strings.HasPrefix(p.remaining(), "map[string]") {
		p.pos += 11
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindMap, Elem: &elem}, nil
	}

	// Check for pointer
	if strings.HasPrefix(p.remaining(), "*") {
		p.pos++
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindPointer, Elem: &elem}, nil
	}

	// Check for struct
	if strings.HasPrefix(p.remaining(), "struct{") {
		return p.parseStruct()
	}

	// Parse base type
	return p.parseBaseType()
}

func (p *typeParser) parseBaseType() (TypeExpr, error) {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) && ch != '.' && ch != '_' {
			break
		}
		p.pos++
	}

	if start == p.pos {
		return TypeExpr{}, fmt.Errorf("expected type at position %d", p.pos)
	}

	baseType := p.input[start:p.pos]
	return TypeExpr{Kind: TypeKindBase, BaseType: baseType}, nil
}

func (p *typeParser) parseStruct() (TypeExpr, error) {
	// Skip "struct{"
	p.pos += 7

	var fields []FieldDef
	for {
		p.skipWhitespace()

		if p.pos >= len(p.input) {
			return TypeExpr{}, fmt.Errorf("unexpected end of struct")
		}

		if p.input[p.pos] == '}' {
			p.pos++
			break
		}

		// Parse field name
		name := p.parseIdentifier()
		if name == "" {
			return TypeExpr{}, fmt.Errorf("expected field name at position %d", p.pos)
		}

		p.skipWhitespace()

		// Parse field type
		fieldType, err := p.parseType()
		if err != nil {
			return TypeExpr{}, fmt.Errorf("invalid field type for %s: %w", name, err)
		}

		fields = append(fields, FieldDef{Name: name, Type: fieldType})

		p.skipWhitespace()

		// Check for separator or end
		if p.pos < len(p.input) && p.input[p.pos] == ';' {
			p.pos++
			continue
		}

		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
			break
		}
	}

	return TypeExpr{Kind: TypeKindStruct, Fields: fields}, nil
}

func (p *typeParser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) && ch != '_' {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *typeParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *typeParser) remaining() string {
	if p.pos >= len(p.input) {
		return ""
	}
	return p.input[p.pos:]
}
