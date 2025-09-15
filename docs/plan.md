了解！“写経しながら”着実に組み上げる前提で、**小さい成功を積み重ねる段階的プラン**にします。各フェーズに「達成条件（DoD）」「作業ステップ」「動作確認」を付けました。CI の“再生成差分で落とす”運用も最終段で導入します。

# 要点（ロードマップ）

1. リポ初期化・最小の骨組み（Go module / ディレクトリ構成）
2. MVP：テンプレを埋め込んで `Template()` + `RenderAny()` だけ（型生成なし）
3. AST 解析導入：`text/template/parse` でフィールド経路を厳密収集
4. 型推論（デフォルト string）とスキーマ木の構築（range→slice, index→map）
5. `@param` 上書き（型だけ）適用
6. コード生成：ネスト struct/type・`Params`・`Render`・`RenderWith` 出力
7. CLI：`gen` サブコマンド（`go:generate` で使えるところまで）
8. 例プロジェクト `examples/` と **e2e（go\:generate → go build）** テスト
9. CI：“生成物コミット + 再生成で差分検出”の整合チェック
10. 仕上げ（`--root-type`/`--emit-params`/`--no-render-typed`、README/仕様書整備）

---

# 詳細プラン

## フェーズ0：準備

**DoD**

* Go 1.22+、git、make（任意）が使える。

**手順**

* 新規リポ作成（例: `github.com/you/templagen`）
* `go mod init github.com/you/templagen`

---

## フェーズ1：骨組み（最小実行）

**DoD**

* `cmd/templagen` バイナリがビルドできる。
* `templagen gen --help` 相当の最低限パースが動く（まだ生成はしないでもOK）。

**手順**

* ディレクトリ：

  ```
  cmd/templagen/
  internal/{magic,scan,gen}/
  examples/{mailtpl/,main.go}
  ```
* `cmd/templagen/main.go` にフラグ（`--in --pkg --out`）だけ実装。
* とりあえず `Template()` + `RenderAny()` を吐くコード生成（固定文字列）で出力→コンパイル通ることを目標。

**確認**

```bash
go build ./cmd/templagen
templagen gen --in examples/mailtpl/email.tmpl --pkg mailtpl --out examples/mailtpl/params_gen.go
go build ./examples
```

---

## フェーズ2：MVP（テンプレ埋め込み + RenderAny）

**DoD**

* 生成された `.go` に `//go:embed <tmpl>`、`Template()`、`RenderAny()` がある。
* `go run ./examples` でテンプレが実行される。

**手順**

* `gen.Emit` の最小版：preamble + imports + `//go:embed` + `Template()` + `RenderAny()`
* examples 側で `map[string]any` を渡して動作確認。

**確認**

* サンプルテンプレを簡単な `{{ .User.Name }}` 程度で確認。

---

## フェーズ3：AST 解析（厳密スコープ）

**DoD**

* `scan.ScanTemplate(src)` が `text/template/parse` を使って `.User.Name` / `range .Items` / `with .User` / `index .Meta "k"` を正しく検出し、**フィールド経路**を収集できる。

**手順**

* `internal/scan/scan_ast.go` を実装：

  * `List/Action/If/With/Range` を DFS。
  * `range .Items` → `Items` を配列扱い、ブロック内の `.Title` 等は要素側へ。
  * `index .Meta "k"` → `Meta` を `map[string]string` 扱い。
  * 結果はスキーマ木（struct/slice/map/leaf=string）に落とす。

**確認**

* 単体テスト or `fmt.Printf` デバッグで `Schema` の構造が期待通りか確認。

---

## フェーズ4：デフォルト string の型推論

**DoD**

* 参照された各フィールドが未指定なら `string`。
* `range` から `[]struct{...}`、`index` から `map[string]string` が入る。

**手順**

* フェーズ3のスキーマ木に `KindString/Struct/Slice/Map/Ptr` と `Children/Elem` を持たせる。
* 参照のたびに `ensureStruct/ensurePath/markSliceStruct/markMapString` で木を拡張。

**確認**

* スキーマ木のダンプで `User{Name string}`、`Items []{Title string}` 等が見える。

---

## フェーズ5：`@param` 上書き

**DoD**

* `{{/* @param Path Type */}}` を複数拾って **最終的な型**に上書きできる。
* サポート：`string/int/int64/float64/bool/time.Time`、`[]T`、`map[string]T`（keyは string 固定）、`*T`、`struct{...}`。

**手順**

* `internal/magic.ParseParams` で正規表現抽出（Path→Type）。
* 簡易型パーサでスキーマ木に上書き（`applyOverride`）。

**確認**

* `@param User.Age int`、`@param Items []struct{ ID int64; Title string }` などで木が変わること。

---

## フェーズ6：フルコード生成

**DoD**

* 生成物に以下が含まれる：

  * `type Params struct { ... }`（トップレベル）
  * ネスト struct/type（要素 struct は自動命名）
  * `Template()` / `Render(Params)` / `RenderAny()` / `RenderWith[T]()`
* `examples` を **型安全な `Params` で実行**できる。

**手順**

* スキーマ木から Go 型文字列を作る `goTypeOf`。
* 命名規則（エクスポート化、要素名 `Item`/`Value` など）を固定し、**安定した順序**で出力。
* `time.Time` を含む場合に `import "time"` を自動追加。

**確認**

* `go generate ./...` → `go build` が通る。
* `examples/main.go` で `mailtpl.Params{...}` を渡して動作。

---

## フェーズ7：CLI（`gen`）の実用化

**DoD**

* `//go:generate templagen gen -in ... -pkg ... -out ...` が動く。
* 生成物をコミットする運用に移行。

**手順**

* `cmd/templagen/main.go` で `--emit-params/--root-type/--no-render-typed` をフラグ化（後続で）。
* まずは `gen` の基本フロー固定：読み込み→解析→生成→書き出し。

**確認**

```bash
(cd examples/mailtpl && go generate ./...)
go build ./examples
```

---

## フェーズ8：examples + e2e テスト

**DoD**

* `go test ./examples -v` で「`go generate` → `go build`」が自動検証される。

**手順**

* `examples/mailtpl/gen.go` に `//go:generate` を記述。
* `examples/e2e_test.go` を作成：`exec.Command` で `go generate` と `go build` を呼ぶ。

**確認**

```bash
go test ./examples -v
```

---

## フェーズ9：CI（差分検出で落とす）

**DoD**

* PR / push 時に「生成物が最新か」を CI が判定し、差分があれば失敗。

**手順**

* スクリプト `scripts/verify_generated.sh`：

  1. ワークツリーがクリーン前提チェック（任意）
  2. `go generate ./...`
  3. `gofmt -s -w .` / `go mod tidy`
  4. `git diff --quiet` で差分検出 → 差分あれば失敗
* GitHub Actions などに組み込み（Go セットアップ → スクリプト実行）。

**確認**

* わざとテンプレを修正して CI が落ちることを確認。

---

## フェーズ10：仕上げ＆拡張

**DoD**

* 開発者ガイドと CLI マニュアル、仕様（`spec.md`/`impl.md`/`cli.md`/`examples.md`）がリポに揃う。
* 任意の追加機能：

  * `--root-type`（外部型エイリアス）
  * `--emit-params=false`（自前型前提・`RenderAny/With` のみ）
  * `--no-render-typed`
  * `templagen check`（必要なら。今回は CI 差分方式が主）

**手順**

* README に「思想／使い方／制限」を簡潔に集約。
* セマンティックバージョニング、タグ付け（`v0.x`）。

---

# 進め方のコツ（写経モード向け）

* **1コミット1小目標**：たとえば「AST で `.User.Name` だけ拾えるようにした」など、粒度を小さく。
* **ダンプを多用**：スキーマ木を `fmt.Printf("%#v\n", tree)` で可視化し、期待と比較。
* **順序の安定化**：生成時は**必ず sort**。import・フィールド・型定義の順序ブレは CI 地獄の元。
* **テンプレは段階的に複雑化**：プレーン→`with`→`range`→`index`→`@param`→ネスト…の順に難度を上げて確認。

---

# 完了の定義（Definition of Done）

* `go generate ./...` → `go build ./...` が問題なく通る。
* `go test ./examples -v` が緑。
* CI の生成差分チェックが動き、差分で確実に落ちる。
* `README` を読めば、非エンジニア／エンジニア双方が使い方に迷わない。

---

必要なら、このプランに沿って\*\*コミット単位のタスクリスト（チェックボックス）\*\*も用意します。
