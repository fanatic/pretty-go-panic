package api

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maruel/panicparse/stack"
)

func CreateRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
		})
	})

	r.Static("/public", "public")
	r.LoadHTMLGlob("api/templates/*")
	r.GET("/", index)
	r.POST("/", upload)
	return r
}

func index(c *gin.Context) {
	render(c, gin.H{
		"title": "Main website",
	})
}

var defaultPalette = stack.Palette{
	EOLReset:               "</span>",
	RoutineFirst:           `<span class="has-text-info has-text-weight-bold">`,
	CreatedBy:              `<span class="has-text-grey">`,
	Package:                `<span class="has-text-weight-bold")`,
	SourceFile:             "</span>",
	FunctionStdLib:         `<span class="has-text-success">`,
	FunctionStdLibExported: `<span class="has-text-success has-text-weight-bold">`,
	FunctionMain:           `<span class="has-text-primary has-text-weight-bold">`,
	FunctionOther:          `<span class="has-text-danger">`,
	FunctionOtherExported:  `<span class="has-text-danger has-text-weight-bold">`,
	Arguments:              "</span>",
}

func upload(c *gin.Context) {
	rawPanic := c.PostForm("text")
	if rawPanic == "" {
		rawPanic = `Panic: Want stack trace

goroutine 1 [running]:
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
				/Users/fanatic/go/main.go:9 +0x64
main.main()
				/Users/fanatic/go/main.go:5 +0x85

goroutine 2 [runnable]:
runtime.forcegchelper()
				/Users/fanatic/go/src/runtime/proc.go:90
runtime.goexit()
				/Users/fanatic/go/src/runtime/asm_amd64.s:2232 +0x1

goroutine 3 [runnable]:
runtime.bgsweep()
				/Users/fanatic/go/src/runtime/mgc0.go:82
runtime.goexit()
				/Users/fanatic/go/src/runtime/asm_amd64.s:2232 +0x1`
	}
	fullPath := false
	p := defaultPalette
	s := stack.AnyValue // stack.AnyPointer
	in := strings.NewReader(rawPanic)
	out := ioutil.Discard
	h := gin.H{
		"rawPanic": rawPanic,
	}
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		h["error"] = "Internal error: " + err.Error()
		render(c, h)
		return
	}

	h["showGotracebackTip"] = len(goroutines) == 1

	buckets := stack.SortBuckets(stack.Bucketize(goroutines, s))
	srcLen, pkgLen := stack.CalcLengths(buckets, fullPath)
	htmlBuckets := []string{}
	for _, bucket := range buckets {
		htmlBuckets = append(htmlBuckets, p.BucketHeader(&bucket, fullPath, len(buckets) > 1))
		htmlBuckets = append(htmlBuckets, p.StackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	h["buckets"] = template.HTML(strings.Join(htmlBuckets, ""))

	render(c, h)
}

func render(c *gin.Context, h gin.H) {
	c.HTML(http.StatusOK, "index.tmpl", h)
}
