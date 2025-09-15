package scan

import (
	"fmt"
	"text/template"
	tplparse "text/template/parse"

	"github.com/bellwood4486/templagen-poc/internal/util"
)

// Kind は推論されたフィールド種別を表します。
type Kind int

const (
	KindString Kind = iota
	KindStruct
	KindSlice
	KindMap
)

// Fileld は推論スキーマ木のノードです。
type Field struct {
	Name     string
	Kind     Kind
	Elem     *Field            // Slice/Map の要素
	Children map[string]*Field // Struct の子
}

// Schema はトップレベル（Params直下）のフィールド集合です。
type Schema struct {
	Fields map[string]*Field
}

// ScanTemplate は Go テンプレートを AST 解析して、.(ドット）スコープを追跡して
// フィールド参照からスキーマ木を推論します。
// 既定では葉はすべて string として扱い、 range は []struct{}, index は map[string]string を推論します。
func ScanTemplate(src string) (Schema, error) {
	// Use text/template to ensure built-in funcs (e.g., index) are defined.
	tmpl, err := template.New("tpl").Parse(src)
	if err != nil {
		return Schema{}, fmt.Errorf("failed to parse template: %w", err)
	}
	if tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return Schema{}, fmt.Errorf("template not found: %s", "tpl")
	}

	s := Schema{Fields: map[string]*Field{}}
	walk(tmpl.Tree.Root, &s, ctx{})

	return s, nil
}

// ctx は現在の .(ドット)を表すパスを保持します。
// with/range でドットが移動したときはこのパスを延長します。
type ctx struct {
	dot []string
}

func (c ctx) with(prefix []string) ctx {
	dup := make([]string, len(c.dot))
	copy(dup, c.dot)
	return ctx{dot: append(dup, prefix...)}
}

// walk はテンプレ AST を DFS します。 with/range/inf での . の取り扱いをテンプレ仕様取りに行います。
func walk(n tplparse.Node, s *Schema, c ctx) {
	switch x := n.(type) {
	case *tplparse.ListNode:
		for _, nn := range x.Nodes {
			walk(nn, s, c)
		}
	case *tplparse.ActionNode:
		collectFromPipe(x.Pipe, s, c)
	case *tplparse.IfNode:
		// if のパイプに出る単独フィールドは存在チェック用途が多いので、
		// 基点フィールドは struct として確保しておくと後続の .Foo.Bar に親和的。
		base := baseFieldFromPipe(x.Pipe)
		if len(base) > 0 {
			ensureStructPath(s, append(c.dot, base...))
		}
		collectFromPipe(x.Pipe, s, c)
		if x.List != nil {
			walk(x.List, s, c)
		}
		if x.ElseList != nil {
			walk(x.ElseList, s, c)
		}
	case *tplparse.WithNode:
		// with 本体では . が基点に切り替わる。 esle 側は元の . に戻る。
		base := baseFieldFromPipe(x.Pipe)
		if len(base) > 0 {
			ensureStructPath(s, append(c.dot, base...))
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
		// range .Items → Items は []struct{] に
		base := baseFieldFromPipe(x.Pipe)
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

// collectFromPipe は {{ .Foo.Bar }} や {{ index .Meta "k" }} など、パイプ内のフィールド参照を収集します。
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
		// 通常のフィールド参照 .Foo.Bar を葉 string として確保
		for _, a := range cmd.Args {
			if f, ok := a.(*tplparse.FieldNode); ok {
				ensurePath(s, append(c.dot, f.Ident...), true)
			}
		}
	}
}

// baseFieldFromPipe はパイプ内で最初に現れるフィールドノード（.Foo.Bar など）の識別子スライスを返します。
func baseFieldFromPipe(p *tplparse.PipeNode) []string {
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

// ensureStructPath は与えられたパス（ドット起点）を「必ず struct の連結」として確保します。
// 途中で既に存在し Kind が string などでも、構造体に昇格させ、Children を確保します。
func ensureStructPath(s *Schema, parts []string) {
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
}

// ensurePath は（通常の）フィールド参照を処理します。
// 1セグメントのみなら葉 string、2セグメント以上なら中間を struct で掘り、葉を string で確保します。
// 既に Slice/Map/Struct(子あり) で確定しているノードは壊さず尊重します。
// 既存が Struct(子なし) の場合は必要に応じて String ⇔ Struct へ昇格/置換します。
func ensurePath(s *Schema, parts []string, leafAsString bool) {
	if len(parts) == 0 {
		return
	}

	// 単一セグメントは葉 string として扱う（既存が確定済みなら尊重）
	if len(parts) == 1 {
		name := parts[0]
		if s.Fields == nil {
			s.Fields = map[string]*Field{}
		}
		if cur, ok := s.Fields[name]; ok && cur != nil {
			switch cur.Kind {
			case KindSlice, KindMap:
				// 既にコンテナとして確定 → 触らない
				return
			case KindStruct:
				if len(cur.Children) == 0 {
					// 子なし struct → 文字列に置換
					*cur = Field{
						Name: util.Export(name),
						Kind: KindString,
					}
				}
				return
			default:
				// string 等 → そのまま
				return
			}
		}
		s.Fields[name] = &Field{
			Name: util.Export(name),
			Kind: KindString,
		}
		return
	}

	// 2要素以上は中間を struct で掘っていく。
	// 先頭セグメントが既に存在する場合は種別を尊重する（特に Slice/Map を上書きしない）。
	// 存在しない場合は struct として作成する。
	if s.Fields == nil {
		s.Fields = map[string]*Field{}
	}
	var cur *Field
	if existing := s.Fields[parts[0]]; existing != nil {
		// Slice/Map は保持し、String 等は struct に昇格させる
		switch existing.Kind {
		case KindSlice, KindMap:
			// そのまま尊重
		case KindStruct:
			if existing.Children == nil {
				existing.Children = map[string]*Field{}
			}
		default:
			existing.Kind = KindStruct
			if existing.Children == nil {
				existing.Children = map[string]*Field{}
			}
		}
		cur = existing
	} else {
		cur = ensureStruct(s.Fields, parts[0])
	}

	for i := 1; i < len(parts); i++ {
		// スライスは要素へ潜る
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
			// 葉は string として確保（既存確定は尊重）
			if ch, ok := cur.Children[name]; ok && ch != nil {
				switch ch.Kind {
				case KindSlice, KindMap:
					return
				case KindStruct:
					if len(ch.Children) == 0 {
						*ch = Field{
							Name: util.Export(name),
							Kind: KindString,
						}
					}
				default:
					return
				}
			}
			cur.Children[name] = &Field{
				Name: util.Export(name),
				Kind: KindString,
			}
			return
		}

		// 中間ノードの処理: 既存の Slice/Map を壊さず尊重し、必要なら昇格
		if ch := cur.Children[name]; ch != nil {
			switch ch.Kind {
			case KindSlice, KindMap:
				// コンテナはそのまま潜る
				cur = ch
			case KindStruct:
				if ch.Children == nil {
					ch.Children = map[string]*Field{}
				}
				cur = ch
			default:
				// String 等 → Struct に昇格
				ch.Kind = KindStruct
				if ch.Children == nil {
					ch.Children = map[string]*Field{}
				}
				cur = ch
			}
		} else {
			cur = ensureStruct(cur.Children, name)
		}
	}
}

// ensureStruct は name に対応するノードを必ず struct として返します。
// 既に存在して Kind が struct 以外でも、struct に「昇格」させ、Children を確保します。
func ensureStruct(m map[string]*Field, name string) *Field {
	if m[name] != nil {
		if m[name].Kind != KindStruct {
			m[name].Kind = KindStruct
		}
		if m[name].Children == nil {
			m[name].Children = map[string]*Field{}
		}
		return m[name]
	}

	m[name] = &Field{
		Name:     util.Export(name),
		Kind:     KindStruct,
		Children: map[string]*Field{},
	}

	return m[name]
}

// markSliceStruct は parts の最終セグメントをスライス（要素は struct）として確定します。
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

// markMapString は parts の最終セグメントを map[string]string として確定します。
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
