package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style

	newFormatter bool = true
)

// CodeBlockInfo represents info about code snippet
type CodeBlockInfo struct {
	Lang          string
	GitHubURI     string
	PlaygroundURI string
}

func init() {
	htmlFormatter = html.New(html.WithClasses(), html.TabWidth(2))
	panicIf(htmlFormatter == nil, "couldn't create html formatter")
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
	panicIf(highlightStyle == nil, "didn't find style '%s'", styleName)
}

// gross hack: we need to change html generated by chroma
func fixupHTMLCodeBlock(htmlCode string, info *CodeBlockInfo) string {
	classLang := ""
	if info.Lang != "" {
		classLang = " lang-" + info.Lang
	}

	if info.GitHubURI == "" && info.PlaygroundURI == "" {
		html := fmt.Sprintf(`
<div class="code-box%s">
	<div>
		%s
	</div>
</div>`, classLang, htmlCode)
		return html
	}

	playgroundPart := ""
	if info.PlaygroundURI != "" {
		playgroundPart = fmt.Sprintf(`
<div class="code-box-playground">
	<a href="%s" target="_blank">try online</a>
</div>
`, info.PlaygroundURI)
	}

	gitHubPart := ""
	if info.GitHubURI != "" {
		// gitHubLoc is sth. like github.com/essentialbooks/books/books/go/main.go
		fileName := path.Base(info.GitHubURI)
		gitHubPart = fmt.Sprintf(`
<div class="code-box-github">
	<a href="%s" target="_blank">%s</a>
</div>`, info.GitHubURI, fileName)
	}

	html := fmt.Sprintf(`
<div class="code-box%s">
	<div>
	%s
	</div>
	<div class="code-box-nav">
		%s
		%s
	</div>
</div>`, classLang, htmlCode, playgroundPart, gitHubPart)
	return html
}

// based on https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	reportOvertime := func() {
		fmt.Printf("Too long processing lang: %s, defaultLang: %s, source:\n%s\n\n", lang, defaultLang, source)
		ioutil.WriteFile("hili_test_case.txt", []byte(source), 0644)
		panic("failed to hilight")
	}
	time.AfterFunc(time.Second*15, reportOvertime)

	if newFormatter {
		htmlFormatter = html.New(html.WithClasses(), html.TabWidth(2))
		panicIf(htmlFormatter == nil, "couldn't create html formatter")
		styleName := "monokailight"
		highlightStyle = styles.Get(styleName)
		panicIf(highlightStyle == nil, "didn't find style '%s'", styleName)
	}

	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func testHang() {
	s := `package main

import (
	"fmt"
	"math"
)

const s string = "constant"

func main() {
	fmt.Println(s) // constant

	// A const statement can appear anywhere a var statement can.
	const n = 10
	fmt.Println(n)                           // 10
	fmt.Printf("n=%d is of type %T\n", n, n) // n=10 is of type int

	const m float64 = 4.3
	fmt.Println(m) // 4.3

	// An untyped constant takes the type needed by its context.
	// For example, here math.Sin expects a float64.
	const x = 10
	fmt.Println(math.Sin(x)) // -0.5440211108893699
}
`
	for i := 0; i < 1024*32; i++ {
		var buf bytes.Buffer
		htmlHighlight(&buf, s, "Go", "")
	}
}
