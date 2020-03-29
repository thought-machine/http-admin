package admin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type entrySlice []entry

func (e entrySlice) Len() int {
	return len(e)
}

func (e entrySlice) Less(i, j int) bool {
	a := e[i]
	b := e[j]
	aL, aLok := a.(link)
	aG, aGok := a.(group)
	bL, bLok := b.(link)
	bG, bGok := b.(group)

	if aLok && bGok {
		return true
	}
	if aGok && bLok {
		return false
	}
	if aLok && bLok {
		return aL.ID < bL.ID
	}
	return aG.Name < bG.Name
}

func (e entrySlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type entry interface {
	isEntry()
}

type group struct {
	Name  string
	Links entrySlice
}

func (g group) isEntry() {}

type link struct {
	ID     string
	HRef   string
	Method string
}

func (l link) isEntry() {}

func renderNav(nav entrySlice, uri string) string {
	builder := strings.Builder{}
	if len(nav) == 0 {
		return builder.String()
	}

	for _, entry := range nav {
		switch v := entry.(type) {
		case link:
			if v.Method == http.MethodGet {
				formattedID := strings.Replace(v.ID, " ", "-", -1)
				selected := ""
				if v.HRef == uri {
					selected = "selected"
				}
				builder.WriteString(fmt.Sprintf(`
<a class="nav-link" href="%s">
	<li id="%s" class="selectable %s">
		%s
	</li>
</a>`, v.HRef, formattedID, selected, v.ID))
			}
		case group:
			isActive := false
			for _, l := range v.Links {
				switch v2 := l.(type) {
				case link:
					if !strings.Contains(strings.TrimPrefix(v2.HRef, uri), "/") {
						isActive = true
					}
				default:
				}
			}
			active := ""
			if isActive {
				active = "active"
			}
			collapse := ""
			if isActive {
				collapse = "fa-caret-square-up"
			}

			builder.WriteString(fmt.Sprintf(`
<li class="subnav %s">
	<div class="subnav-title selectable">
		<span class="fas fa-caret-square-right %s"></span>
		<span>%s</span>
	</div>
	<ul>%s</ul>
</li>
`, active, collapse, v.Name, renderNav(v.Links, uri)))
		}
	}

	return builder.String()
}

func render(title, uri string, nav entrySlice, contents io.Reader) io.Reader {
	return io.MultiReader(strings.NewReader(fmt.Sprintf(`
<!doctype html>
<html lang="en">
	<head>
		<title>%s &middot; TM Server Admin</title>
		<!-- css -->
		<link type="text/css" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet"/>
		<link type="text/css" href="/admin/files/css/index.css" rel="stylesheet"/>
		<link type="text/css" href="/admin/files/css/client-registry.css" rel="stylesheet"/>
		<link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.8.1/css/all.css" integrity="sha384-50oBUHEmvpQ+1lW4y57PTFmhCaXp0ML5d60M1M7uH2+nqUivzIebhndOJK28anvf" crossorigin="anonymous">
		<!-- js -->
		<script type="application/javascript" src="//www.google.com/jsapi"></script>
		<script type="application/javascript" src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.0/jquery.min.js"></script>
		<script type="application/javascript" src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js"></script>
		<script type="application/javascript" src="/admin/files/js/index.js"></script>
		<script type="application/javascript" src="/admin/files/js/utils.js"></script>
	</head>
	<body>
		<div class="container-fluid" id="wrapper">
			<nav class="nav" id="sidebar">
				<ul>%s</ul>
			</nav>
			<div id="toggle"><span class="fas fa-angle-left"></span></div>
			<div id="contents">
				<div class="row">
					<div class="col-md-12">`, title, renderNav(nav, uri))),
		contents,
		strings.NewReader(`</div>
				</div>
			</div>
		</div>
	</body>
</html>`))
}

type indexView struct {
	title   string
	next    http.Handler
	entries func() entrySlice
}

type cachingResponseWriter struct {
	w          http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (c *cachingResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *cachingResponseWriter) Write(b []byte) (int, error) {
	return c.buffer.Write(b)
}

func (c *cachingResponseWriter) WriteHeader(code int) {
	c.statusCode = code
}

func mkCachingResponseWriter(underlying http.ResponseWriter) *cachingResponseWriter {
	return &cachingResponseWriter{
		w:          underlying,
		buffer:     new(bytes.Buffer),
		statusCode: 200,
	}
}

func (i *indexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !expectsHTML(r) {
		i.next.ServeHTTP(w, r)
	} else {
		cw := mkCachingResponseWriter(w)
		i.next.ServeHTTP(cw, r)

		entries := i.entries()
		sort.Stable(entries)
		contentType := cw.Header().Get("Content-Type")
		content := cw.buffer.String()
		if !isFragment(contentType, content) {
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(cw.statusCode)
			io.Copy(w, cw.buffer)
		} else {
			w.Header().Set("Content-Type", "text/html;charset=UTF-8")
			io.Copy(w, render(i.title, r.URL.Path, entries, cw.buffer))
		}
	}
}

func isFragment(contentType, content string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html") && !strings.Contains(content, "<html>")
}

func accepts(r *http.Request, contentType string) bool {
	for _, h := range r.Header["Accept"] {
		if strings.Contains(h, contentType) {
			return true
		}
	}
	return false
}

func expectsHTML(r *http.Request) bool {
	return strings.HasSuffix(r.URL.Path, ".html") || accepts(r, "text/html")
}

// SummaryHandler renders the front-page content.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	procInfo := []string{"process_uptime", "go_goroutines", "go_memstats_alloc_bytes", "go_gc_duration_seconds"}

	var b bytes.Buffer
	b.WriteString(`<script type="application/javascript" src="/admin/files/js/summary.js"></script>
      <link type="text/css" href="/admin/files/css/summary.css" rel="stylesheet">
      <div id="lint-warnings" data-refresh-uri="/admin/failedlint"></div>
      <div id="process-info" class="text-center card" data-refresh-uri="/admin/metrics">
        <ul class="list-inline">
          <li class="list-inline-item"><span class="fas fa-info-circle"/></li>`)

	for _, key := range procInfo {
		b.WriteString(fmt.Sprintf(`<li class="list-inline-item" data-key="%s">
                    <div>
                      <a href="/admin/metrics#%s">%s:</a>
                      <span id="%s">...</span>
                      &middot;
                    </div>
                  </li>`, key, key, key, strings.Replace(key, "/", "-", -1)) + "\n")
	}
	b.WriteString(`<br />
        </ul>
      </div>
      <div id="server-info" data-refresh-uri="/admin/servers/index.txt"></div>
      <div id="client-info" data-refresh-uri="/admin/clients/index.txt"></div>
`)
	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	w.Write(b.Bytes())
}
