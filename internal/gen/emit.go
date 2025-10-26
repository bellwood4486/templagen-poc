package gen

import (
	"fmt"
	"go/format"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/bellwood4486/tmpltype/internal/scan"
	"github.com/bellwood4486/tmpltype/internal/typing"
	"github.com/bellwood4486/tmpltype/internal/util"
)

// Unit は単一のテンプレート処理単位
type Unit struct {
	Pkg           string // 出力パッケージ名
	SourcePath    string // 埋め込むテンプレファイルのパス（go:embedディレクティブで使用）
	SourceLiteral string // テンプレ本文
}

// tmpl は単一テンプレートのコード生成に必要な情報
type tmpl struct {
	name       string              // テンプレート名
	groupName  string              // グループ名（空ならフラット）
	typeName   string              // 生成する型名
	sourcePath string              // テンプレートファイルパス
	varName    string              // embed変数名
	typed      *typing.TypedSchema // 型情報
}

// tmplGroup はテンプレートグループのコード生成に必要な情報
type tmplGroup struct {
	name      string // グループ名
	typeName  string // 生成する型名
	templates []tmpl // グループ内のテンプレート
}

// emitPrepared は解析・準備が完了したコード生成のための情報
type emitPrepared struct {
	pkg           string
	imports       map[string]struct{}
	groups        []tmplGroup // グループ
	flatTemplates []tmpl      // フラットなテンプレート
}

// allTemplates はフラットとグループ内の全テンプレートを返す
func (p *emitPrepared) allTemplates() []tmpl {
	all := make([]tmpl, 0, len(p.flatTemplates)+len(p.groups)*2)
	all = append(all, p.flatTemplates...)
	for _, g := range p.groups {
		all = append(all, g.templates...)
	}
	return all
}

// prepare はテンプレートをスキャンし、型を解決して、コード生成に必要なデータを準備する
func prepare(units []Unit, basedir string) (*emitPrepared, error) {
	if len(units) == 0 {
		return nil, fmt.Errorf("no units provided")
	}

	templates := make([]tmpl, 0, len(units))
	allImports := make(map[string]struct{})

	// デフォルトのimport
	allImports["io"] = struct{}{}
	allImports["text/template"] = struct{}{}
	allImports["embed"] = struct{}{}
	allImports["fmt"] = struct{}{}

	// 各テンプレートを処理
	for _, unit := range units {
		// テンプレート名を抽出 (例: "mail_invite/title" または "footer")
		templateName, err := extractTemplateName(unit.SourcePath, basedir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract template name from %s: %w", unit.SourcePath, err)
		}

		// グループ名を抽出 (スラッシュが含まれていればグループ)
		var groupName string
		var localName string
		if strings.Contains(templateName, "/") {
			parts := strings.Split(templateName, "/")
			groupName = parts[0]
			localName = parts[1]
		} else {
			localName = templateName
		}

		// 型名を生成 (例: "MailInviteTitle" または "Footer")
		var typeName string
		if groupName != "" {
			typeName = util.Export(groupName) + util.Export(localName)
		} else {
			typeName = util.Export(localName)
		}

		// embed変数名を生成 (スラッシュをアンダースコアに変換)
		varName := strings.ReplaceAll(templateName, "/", "_") + "TplSource"

		// テンプレートをスキャン
		sch, err := scan.ScanTemplate(unit.SourceLiteral)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template %s: %w", unit.SourcePath, err)
		}

		// 型解決
		typed, err := typing.Resolve(sch, unit.SourceLiteral)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve types for %s: %w", unit.SourcePath, err)
		}

		// テンプレートデータを追加
		templates = append(templates, tmpl{
			name:       templateName,
			groupName:  groupName,
			typeName:   typeName,
			sourcePath: unit.SourcePath,
			varName:    varName,
			typed:      typed,
		})
	}

	// テンプレート名でソート（出力を安定させるため）
	slices.SortFunc(templates, func(a, b tmpl) int {
		return strings.Compare(a.name, b.name)
	})

	// グループ情報を整理
	groups, flatTemplates := organizeGroups(templates)

	return &emitPrepared{
		pkg:           units[0].Pkg, // すべて同じパッケージ名のはず
		imports:       allImports,
		groups:        groups,
		flatTemplates: flatTemplates,
	}, nil
}

// organizeGroups はテンプレートをグループとフラットに分類する
func organizeGroups(templates []tmpl) ([]tmplGroup, []tmpl) {
	groupMap := make(map[string][]tmpl)
	var flatTemplates []tmpl

	for _, t := range templates {
		if t.groupName != "" {
			groupMap[t.groupName] = append(groupMap[t.groupName], t)
		} else {
			flatTemplates = append(flatTemplates, t)
		}
	}

	// グループ情報を構築
	var groups []tmplGroup
	for groupName, groupTemplates := range groupMap {
		groups = append(groups, tmplGroup{
			name:      groupName,
			typeName:  util.Export(groupName),
			templates: groupTemplates,
		})
	}

	// グループ名でソート
	slices.SortFunc(groups, func(a, b tmplGroup) int {
		return strings.Compare(a.name, b.name)
	})

	return groups, flatTemplates
}

// Emit は複数のテンプレートから1つの統合Goファイルを生成する
// 単一テンプレートの場合も同じフォーマットで生成される
func Emit(units []Unit, basedir string) (string, error) {
	// Phase 1: データ収集と準備
	prepared, err := prepare(units, basedir)
	if err != nil {
		return "", err
	}

	// Phase 2: コード生成
	var b strings.Builder
	generateHeader(&b, prepared.pkg)
	generateImports(&b, prepared.imports)
	generateTemplateNamespace(&b, prepared)
	generateEmbedDeclarations(&b, prepared.allTemplates())
	generateTemplateInitialization(&b, prepared)
	generateTemplatesFunction(&b)
	generateGenericRenderFunction(&b)
	generateTemplateBlocks(&b, prepared.allTemplates())

	// Phase 3: フォーマット
	return formatCode(b.String())
}

// generateHeader はパッケージ宣言とコメントを生成する
func generateHeader(b *strings.Builder, pkg string) {
	write(b, "// Code generated by tmpltype; DO NOT EDIT.\n")
	write(b, "package %s\n\n", pkg)
}

// generateImports はimportセクションを生成する
func generateImports(b *strings.Builder, imports map[string]struct{}) {
	write(b, "import (\n")
	keys := slices.Sorted(maps.Keys(imports))
	for _, k := range keys {
		if k == "embed" {
			write(b, "\t_ %q\n", k)
		} else {
			write(b, "\t%q\n", k)
		}
	}
	write(b, ")\n\n")
}

// generateTemplateNamespace はTemplateName型と名前空間を生成する
func generateTemplateNamespace(b *strings.Builder, p *emitPrepared) {
	write(b, "// TemplateName is a type-safe template name\n")
	write(b, "type TemplateName string\n\n")
	write(b, "// Template provides type-safe access to template names\n")
	write(b, "var Template = struct {\n")

	// フラットなテンプレート
	for _, t := range p.flatTemplates {
		write(b, "\t%s TemplateName\n", t.typeName)
	}

	// グループ
	for _, g := range p.groups {
		write(b, "\t%s struct {\n", g.typeName)
		for _, t := range g.templates {
			// ローカル名（グループプレフィックスなし）を取得
			localName := strings.TrimPrefix(t.typeName, g.typeName)
			write(b, "\t\t%s TemplateName\n", localName)
		}
		write(b, "\t}\n")
	}

	write(b, "}{\n")

	// フラットなテンプレートの初期化
	for _, t := range p.flatTemplates {
		write(b, "\t%s: %q,\n", t.typeName, t.name)
	}

	// グループの初期化
	for _, g := range p.groups {
		write(b, "\t%s: struct {\n", g.typeName)
		for _, t := range g.templates {
			localName := strings.TrimPrefix(t.typeName, g.typeName)
			write(b, "\t\t%s TemplateName\n", localName)
		}
		write(b, "\t}{\n")
		for _, t := range g.templates {
			localName := strings.TrimPrefix(t.typeName, g.typeName)
			write(b, "\t\t%s: %q,\n", localName, t.name)
		}
		write(b, "\t},\n")
	}

	write(b, "}\n\n")
}

// generateEmbedDeclarations は各テンプレートのembed宣言を生成する
func generateEmbedDeclarations(b *strings.Builder, templates []tmpl) {
	for _, t := range templates {
		write(b, "//go:embed %s\n", t.sourcePath)
		write(b, "var %s string\n\n", t.varName)
	}
}

// generateTemplateInitialization はテンプレート初期化のためのヘルパー関数とマップを生成する
func generateTemplateInitialization(b *strings.Builder, p *emitPrepared) {
	// Helper function for template initialization
	write(b, "func newTemplate(name TemplateName, source string) *template.Template {\n")
	write(b, "\treturn template.Must(template.New(string(name)).Option(%q).Parse(source))\n", "missingkey=error")
	write(b, "}\n\n")

	// Templates map - initialized once at package initialization
	write(b, "var templates = map[TemplateName]*template.Template{\n")

	// フラットなテンプレート
	for _, t := range p.flatTemplates {
		fieldRef := "Template." + t.typeName
		write(b, "\t%s: newTemplate(%s, %s),\n",
			fieldRef, fieldRef, t.varName)
	}

	// グループ内のテンプレート
	for _, g := range p.groups {
		for _, t := range g.templates {
			localName := strings.TrimPrefix(t.typeName, g.typeName)
			fieldRef := "Template." + g.typeName + "." + localName
			write(b, "\t%s: newTemplate(%s, %s),\n",
				fieldRef, fieldRef, t.varName)
		}
	}

	write(b, "}\n\n")
}

// generateTemplatesFunction はTemplates()関数を生成する
func generateTemplatesFunction(b *strings.Builder) {
	write(b, "// Templates returns a map of all templates\n")
	write(b, "func Templates() map[TemplateName]*template.Template {\n")
	write(b, "\treturn templates\n")
	write(b, "}\n\n")
}

// generateGenericRenderFunction は汎用Render関数を生成する
func generateGenericRenderFunction(b *strings.Builder) {
	write(b, "// Render renders a template by name with the given data\n")
	write(b, "func Render(w io.Writer, name TemplateName, data any) error {\n")
	write(b, "\ttmpl, ok := templates[name]\n")
	write(b, "\tif !ok {\n")
	write(b, "\t\treturn fmt.Errorf(\"template %%q not found\", name)\n")
	write(b, "\t}\n")
	write(b, "\treturn tmpl.Execute(w, data)\n")
	write(b, "}\n\n")
}

// generateTemplateBlocks は各テンプレートごとの型定義とRender関数を生成する
func generateTemplateBlocks(b *strings.Builder, templates []tmpl) {
	generatedTypes := make(map[string]bool)

	for _, t := range templates {
		// テンプレートブロックのセパレータ
		write(b, "// ============================================================\n")
		write(b, "// %s template\n", t.name)
		write(b, "// ============================================================\n\n")

		generateNamedTypes(b, t, generatedTypes)
		generateParamType(b, t)
		generateRenderFunction(b, t)
	}
}

// generateNamedTypes は名前付き型を生成する
func generateNamedTypes(b *strings.Builder, t tmpl, generatedTypes map[string]bool) {
	for _, namedType := range t.typed.NamedTypes {
		// 型名の衝突を避けるため、プレフィックスを付ける
		typeName := t.typeName + namedType.Name
		if generatedTypes[typeName] {
			continue // すでに生成済み
		}
		generatedTypes[typeName] = true

		write(b, "type %s struct {\n", typeName)
		// フィールドをソートして順序を安定化
		fieldNames := slices.Sorted(maps.Keys(namedType.Fields))
		for _, fieldName := range fieldNames {
			field := namedType.Fields[fieldName]
			// フィールドの型名も調整が必要な場合がある
			goType := adjustTypeForTemplate(field.GoType, t.typeName)
			write(b, "\t%s %s\n", field.Name, goType)
		}
		write(b, "}\n\n")
	}
}

// generateParamType はメインのパラメータ型を生成する
func generateParamType(b *strings.Builder, t tmpl) {
	write(b, "// %s represents parameters for %s template\n", t.typeName, t.name)
	write(b, "type %s struct {\n", t.typeName)
	// トップレベルフィールドをソートして順序を安定化
	topFieldNames := slices.Sorted(maps.Keys(t.typed.Fields))
	for _, fieldName := range topFieldNames {
		field := t.typed.Fields[fieldName]
		// フィールドの型名も調整が必要な場合がある
		goType := adjustTypeForTemplate(field.GoType, t.typeName)
		write(b, "\t%s %s\n", field.Name, goType)
	}
	write(b, "}\n\n")
}

// generateRenderFunction は型安全なRender関数を生成する
func generateRenderFunction(b *strings.Builder, t tmpl) {
	funcName := "Render" + t.typeName

	// フィールド参照を構築 (グループ対応)
	var fieldRef string
	if t.groupName != "" {
		groupTypeName := util.Export(t.groupName)
		localName := strings.TrimPrefix(t.typeName, groupTypeName)
		fieldRef = "Template." + groupTypeName + "." + localName
	} else {
		fieldRef = "Template." + t.typeName
	}

	write(b, "// %s renders the %s template\n", funcName, t.name)
	write(b, "func %s(w io.Writer, p %s) error {\n", funcName, t.typeName)
	write(b, "\ttmpl, ok := templates[%s]\n", fieldRef)
	write(b, "\tif !ok {\n")
	write(b, "\t\treturn fmt.Errorf(\"template %%q not found\", %s)\n", fieldRef)
	write(b, "\t}\n")
	write(b, "\treturn tmpl.Execute(w, p)\n")
	write(b, "}\n\n")
}

// formatCode はgo/formatでコードをフォーマットする
func formatCode(code string) (string, error) {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return "", fmt.Errorf("failed to format generated code: %w", err)
	}
	return string(formatted), nil
}

// extractTemplateName はファイルパスからテンプレート名を抽出する
// basedir からの相対パスでグループ判定を行う
// 例: basedir="templates", path="templates/footer.tmpl" -> "footer" (フラット)
// 例: basedir="templates", path="templates/email/welcome.tmpl" -> "email/welcome" (グループ)
func extractTemplateName(path string, basedir string) (string, error) {
	// basedir からの相対パスを取得
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absBasedir, err := filepath.Abs(basedir)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(absBasedir, absPath)
	if err != nil {
		return "", fmt.Errorf("path %s is not under basedir %s", path, basedir)
	}

	// 拡張子を削除
	pathWithoutExt := strings.TrimSuffix(relPath, filepath.Ext(relPath))

	// ディレクトリ区切りで分割
	parts := strings.Split(filepath.ToSlash(pathWithoutExt), "/")

	// 各パーツから数字プレフィックスを削除してクリーンアップ
	for i, part := range parts {
		parts[i] = cleanName(part)
	}

	// 階層チェック（フラット=1パーツ、グループ=2パーツ、それ以上はエラー）
	if len(parts) > 2 {
		return "", fmt.Errorf("template nesting too deep: %s (max 1 level of grouping)", relPath)
	}

	// パスとして結合
	return strings.Join(parts, "/"), nil
}

// cleanName は名前から数字プレフィックスを削除し、ハイフンをアンダースコアに変換する
func cleanName(name string) string {
	// 数字プレフィックスを削除（例: "01_header" -> "header", "1-mail" -> "mail"）
	re := regexp.MustCompile(`^\d+[-_]`)
	name = re.ReplaceAllString(name, "")

	// ハイフンをアンダースコアに変換
	name = strings.ReplaceAll(name, "-", "_")

	return name
}

// adjustTypeForTemplate は型名をテンプレート固有に調整する
func adjustTypeForTemplate(goType string, templatePrefix string) string {
	// 名前付き型への参照を調整
	// 例: "[]ItemsItem" -> "[]UserItemsItem" (Userテンプレートの場合)
	// これは簡略化された実装。実際にはより複雑な型の処理が必要

	// スライスの場合
	if strings.HasPrefix(goType, "[]") {
		elemType := goType[2:]
		if !isBuiltinType(elemType) && !strings.Contains(elemType, ".") {
			// カスタム型の場合、プレフィックスを付ける
			return "[]" + templatePrefix + elemType
		}
	}

	// マップの場合
	if strings.HasPrefix(goType, "map[string]") {
		elemType := goType[11:] // "map[string]" の後の部分
		if !isBuiltinType(elemType) && !strings.Contains(elemType, ".") {
			return "map[string]" + templatePrefix + elemType
		}
	}

	// 単純な名前付き型の場合
	if !isBuiltinType(goType) && !strings.Contains(goType, ".") &&
		!strings.Contains(goType, "[") && !strings.HasPrefix(goType, "*") {
		return templatePrefix + goType
	}

	return goType
}

func isBuiltinType(typeName string) bool {
	builtins := []string{
		"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "any",
		"time.Time", "error",
	}
	for _, b := range builtins {
		if typeName == b {
			return true
		}
	}
	return false
}

// write は strings.Builder への書き込みヘルパー
// 万が一失敗した場合は panic する
func write(b *strings.Builder, format string, args ...any) {
	_, err := fmt.Fprintf(b, format, args...)
	if err != nil {
		panic(err)
	}
}
