package scan

import (
	"fmt"
	tplparse "text/template/parse"

	"github.com/bellwood4486/templagen-poc/internal/util"
)

// Kind は推論されたフィールド種別
type Kind int

const (
	KindString Kind = iota
	KindStruct
	KindSlice
	KindMap
)

// Fileld はスキーマ木のノード
type Field struct {
	Name     string
	Kind     Kind
	Elem     *Field            // Slice/Map の要素
	Children map[string]*Field // Struct の子
}

// Schema はトップレベル（Params直下）のフィールド集合
type Schema struct {
	Fields map[string]*Field
}

func ScanTemplate(src string) (Schema, error) {
	trees, err := tplparse.Parse("tpl", src, "", "")
	if err != nil {
		return Schema{}, fmt.Errorf("failed to parse template: %w", err)
	}
	t := trees["tpl"]
	if t == nil {
		return Schema{}, fmt.Errorf("template not found: %s", "tpl")
	}

	s := Schema{Fields: map[string]*Field{}}
	walk(t.Root, &s, ctx{})

	return s, nil
}

type ctx struct {
	dot []string // 現在の . のパス（with/range で増える）
}

func (c ctx) with(prefix []string) ctx {
	dup := make([]string, len(c.dot))
	copy(dup, c.dot)
	return ctx{dot: append(dup, prefix...)}
}

func walk(n tplparse.Node, s *Schema, c ctx) {
	switch x := n.(type) {
	case *tplparse.ListNode:
		for _, nn := range x.Nodes {
			walk(nn, s, c)
		}
	case *tplparse.ActionNode:
		collectFromPipe(x.Pipe, s, c)
	case *tplparse.IfNode:
		collectFromPipe(x.Pipe, s, c)
		if x.List != nil {
			walk(x.List, s, c)
		}
		if x.ElseList != nil {
			walk(x.ElseList, s, c)
		}
	case *tplparse.WithNode:
		base := firstFieldFromPipe(x.Pipe)
		if len(base) > 0 {
			ensurePath(s, append(c.dot, base...), true)
		}
		nc := c
		if len(base) > 0 {
			nc = c.with(base)
		}
		if x.List != nil {
			walk(x.List, s, nc)
		}
		if x.ElseList != nil {
			walk(x.ElseList, s, c)
		}
	case *tplparse.RangeNode:
		base := firstFieldFromPipe(x.Pipe)
		if len(base) > 0 {
			markSliceStruct(s, append(c.dot, base...))
		}
		nc := c
		if len(base) > 0 {
			nc = c.with(base)
		}
		if x.List != nil {
			walk(x.List, s, nc)
		}
		if x.ElseList != nil {
			walk(x.ElseList, s, c)
		}
	}
}

func collectFromPipe(p *tplparse.PipeNode, s *Schema, c ctx) {
	if p == nil {
		return
	}

	for _, cmd := range p.Cmds {
		// index .Meta "key" → Meta は map[string]string
		if len(cmd.Args) >= 2 {
			if id, ok := cmd.Args[0].(*tplparse.IdentifierNode); ok && id.Ident == "index" {
				if fn, ok := cmd.Args[1].(*tplparse.FieldNode); ok {
					markMapString(s, append(c.dot, fn.Ident...))
				}
			}
		}
		for _, a := range cmd.Args {
			if f, ok := a.(*tplparse.FieldNode); ok {
				ensurePath(s, append(c.dot, f.Ident...), true)
			}
		}
	}
}

func firstFieldFromPipe(p *tplparse.PipeNode) []string {
	if p == nil {
		return nil
	}

	for _, cmd := range p.Cmds {
		for _, a := range cmd.Args {
			if f, ok := a.(*tplparse.FieldNode); ok && len(f.Ident) > 0 {
				return f.Ident
			}
		}
	}

	return nil
}

func ensurePath(s *Schema, parts []string, leafAsString bool) {
	if len(parts) == 0 {
		return
	}

	// 単一セグメントは葉(string)として扱う
	if len(parts) == 1 {
		name := parts[0]
		// 既に別の形で確定しているなら壊さない
		if cur, ok := s.Fields[name]; ok && cur != nil {
			// 子を持つstruct / slice / map は触らない
			if cur.Kind == KindStruct && len(cur.Children) == 0 {
				*cur = Field{
					Name: util.Export(name),
					Kind: KindString,
				}
			}
			// それ以外（Slice/Map/Struct(子あり)はそのまま
			return
		}
		if s.Fields == nil {
			s.Fields = map[string]*Field{}
		}
		s.Fields[name] = &Field{
			Name: util.Export(name),
			Kind: KindString,
		}
		return
	}

	// 2要素以上は中間をstructで掘っていく
	if s.Fields == nil {
		s.Fields = map[string]*Field{}
	}
	cur := ensureStruct(s.Fields, parts[0])
	for i := 1; i < len(parts); i++ {
		// slice の場合は要素へ降りる
		if cur.Kind == KindSlice {
			if cur.Elem == nil {
				cur.Elem = &Field{
					Name:     cur.Name + "Item",
					Kind:     KindStruct,
					Children: map[string]*Field{},
				}
			}
			cur = cur.Elem
		}
		if cur.Children == nil {
			cur.Children = map[string]*Field{}
		}

		name := parts[i]
		if i == len(parts)-1 {
			// 葉は string として作る（既存が確定していれば尊重）
			if ch, ok := cur.Children[name]; ok && ch != nil {
				if ch.Kind == KindStruct && len(ch.Children) == 0 {
					*ch = Field{
						Name: util.Export(name),
						Kind: KindString,
					}
				}
				return
			}
			cur.Children[name] = &Field{
				Name: util.Export(name),
				Kind: KindString,
			}
			return
		}
		cur = ensureStruct(cur.Children, name)
	}
}

func ensureStruct(m map[string]*Field, name string) *Field {
	if m[name] == nil {
		m[name] = &Field{
			Name:     util.Export(name),
			Kind:     KindStruct,
			Children: map[string]*Field{},
		}
	}

	return m[name]
}

func markSliceStruct(s *Schema, parts []string) {
	if len(parts) == 0 {
		return
	}

	if s.Fields == nil {
		s.Fields = map[string]*Field{}
	}
	cur := ensureStruct(s.Fields, parts[0])
	for i := 1; i < len(parts); i++ {
		if cur.Children == nil {
			cur.Children = map[string]*Field{}
		}
		cur = ensureStruct(cur.Children, parts[i])
	}
	cur.Kind = KindSlice
	if cur.Elem == nil {
		cur.Elem = &Field{
			Name:     cur.Name + "Item",
			Kind:     KindStruct,
			Children: map[string]*Field{},
		}
	}
}

func markMapString(s *Schema, parts []string) {
	if len(parts) == 0 {
		return
	}

	if s.Fields == nil {
		s.Fields = map[string]*Field{}
	}
	cur := ensureStruct(s.Fields, parts[0])
	for i := 1; i < len(parts); i++ {
		if cur.Children == nil {
			cur.Children = map[string]*Field{}
		}
		cur = ensureStruct(cur.Children, parts[i])
	}
	cur.Kind = KindMap
	cur.Elem = &Field{
		Name: cur.Name + "Value",
		Kind: KindString, // string を既定
	}
}
