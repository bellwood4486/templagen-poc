package magic

import (
	"fmt"
	"strings"
	"unicode"
)

// parseType は型文字列をTypeExprにパースする
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

	// スライスのチェック
	if strings.HasPrefix(p.remaining(), "[]") {
		p.pos += 2
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindSlice, Elem: &elem}, nil
	}

	// マップのチェック
	if strings.HasPrefix(p.remaining(), "map[string]") {
		p.pos += 11
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindMap, Elem: &elem}, nil
	}

	// ポインタのチェック
	if strings.HasPrefix(p.remaining(), "*") {
		p.pos++
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindPointer, Elem: &elem}, nil
	}

	// 構造体のチェック
	if strings.HasPrefix(p.remaining(), "struct{") {
		return p.parseStruct()
	}

	// 基本型のパース
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
	// "struct{" をスキップ
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

		// フィールド名のパース
		name := p.parseIdentifier()
		if name == "" {
			return TypeExpr{}, fmt.Errorf("expected field name at position %d", p.pos)
		}

		p.skipWhitespace()

		// フィールド型のパース
		fieldType, err := p.parseType()
		if err != nil {
			return TypeExpr{}, fmt.Errorf("invalid field type for %s: %w", name, err)
		}

		fields = append(fields, FieldDef{Name: name, Type: fieldType})

		p.skipWhitespace()

		// セパレータまたは終端のチェック
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
