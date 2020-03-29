// Package admin implements an HTTP server providing useful information & tools about this server.
package admin

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/trace"
	"gopkg.in/op/go-logging.v1"
)

var log Logger = logging.MustGetLogger("admin")

// Top level groups for the Sidebar nav
var (
	ProcessInfoGroup = "Process Info"
	PerfProfileGroup = "Performance Profile"
	UtilitiesGroup   = "Utilities"
	MetricsGroup     = "Metrics"
)

// Well known paths
var (
	RootPath    = "/"
	AdminPath   = "/admin"
	ClientsPath = AdminPath + "/clients/"
	ServersPath = AdminPath + "/servers/"
	FilesPath   = AdminPath + "/files/"
)

// Route represents a path and a handler, making it possible to display a menu of routes in the UI.
type Route struct {
	path           string
	prefix         bool
	handler        http.Handler
	alias          string
	group          string
	includeInIndex bool
	method         string
}

// Opts is all flags associated with the admin HTTP server.
type Opts struct {
	Disabled bool       `long:"disabled" description:"If true, the admin server will never start." env:"ADMIN_DISABLE_HTTP"`
	Port     int        `long:"port" default:"9990" description:"The port to listen on."`
	Logger   Logger     `no-flag:"true"`
	LogInfo  LoggerInfo `no-flag:"true"`
}

// DefaultAdminHTTPServer is the global admin http server.
var DefaultAdminHTTPServer = &HTTPServer{}

var routes = []Route{
	{
		path:           RootPath,
		handler:        RedirectHandler(AdminPath, http.StatusTemporaryRedirect),
		alias:          "Admin Redirect",
		includeInIndex: false,
	},
	{
		path:           AdminPath,
		handler:        http.HandlerFunc(SummaryHandler),
		alias:          "Summary",
		includeInIndex: true,
	},
	{
		path:           AdminPath + "/",
		handler:        RedirectHandler(AdminPath, http.StatusTemporaryRedirect),
		alias:          "Admin Redirect",
		includeInIndex: false,
	},
	{
		path:           "/admin/ping",
		handler:        http.HandlerFunc(PingHandler),
		alias:          "Ping",
		includeInIndex: true,
		group:          UtilitiesGroup,
	},
	{
		path:           "/admin/gc",
		handler:        http.HandlerFunc(gcHandler),
		alias:          "Garbage Collect",
		includeInIndex: true,
		group:          UtilitiesGroup,
	},
	{
		path:           "/admin/logging",
		handler:        http.HandlerFunc(LoggingHandler),
		alias:          "Logging",
		group:          UtilitiesGroup,
		includeInIndex: true,
	},
	{
		path:           "/admin/logging",
		handler:        http.HandlerFunc(UpdateLoggingHandler),
		method:         http.MethodPost,
		includeInIndex: false,
	},
	{
		path:           "/admin/metrics",
		handler:        http.HandlerFunc(MetricQueryHandler),
		alias:          "Metrics",
		includeInIndex: true,
		group:          MetricsGroup,
	},
	{
		path:           "/debug/pprof/",
		handler:        http.HandlerFunc(pprof.Index),
		alias:          "PProf",
		includeInIndex: true,
		group:          PerfProfileGroup,
	},
	{
		path:           "/debug/pprof/cmdline",
		handler:        http.HandlerFunc(pprof.Cmdline),
		alias:          "CmdLine",
		includeInIndex: true,
		group:          PerfProfileGroup,
	},
	{
		path:           "/debug/pprof/profile",
		handler:        http.HandlerFunc(pprof.Profile),
		alias:          "Profile",
		includeInIndex: true,
		group:          PerfProfileGroup,
	},
	{
		path:           "/debug/pprof/symbol",
		handler:        http.HandlerFunc(pprof.Symbol),
		alias:          "Symbol",
		includeInIndex: true,
		group:          PerfProfileGroup,
	},
	{
		path:           "/debug/pprof/trace",
		handler:        http.HandlerFunc(pprof.Trace),
		alias:          "Trace",
		includeInIndex: true,
		group:          PerfProfileGroup,
	},
	{
		path:           "/debug/pprof/",
		prefix:         true,
		handler:        http.HandlerFunc(pprof.Index),
		includeInIndex: false,
	},
	{
		path:           "/debug/events",
		handler:        http.HandlerFunc(trace.Events),
		alias:          "Event Traces",
		includeInIndex: true,
		group:          UtilitiesGroup,
	},
	{
		path:           "/debug/requests",
		handler:        http.HandlerFunc(trace.Traces),
		alias:          "Request Traces",
		includeInIndex: true,
		group:          UtilitiesGroup,
	},
	{
		path:           "/debug/vars",
		handler:        expvar.Handler(),
		alias:          "Vars",
		includeInIndex: true,
		group:          ProcessInfoGroup,
	},
	{
		path: "/metrics",
		handler: promhttp.HandlerFor(
			prometheus.Gatherers{Gatherer},
			promhttp.HandlerOpts{}),
		includeInIndex: false,
		alias:          "Metrics",
	},
	{
		path:           "/favicon.ico",
		handler:        ResourceHandler("/", "img"),
		alias:          "Favicon",
		includeInIndex: false,
	},
	{
		path:           FilesPath,
		prefix:         true,
		handler:        ResourceHandler(FilesPath, "admin"),
		includeInIndex: false,
		alias:          "Files",
	},
}

// PingHandler serves a very simple response that can be used as a health check.
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(path.Base(os.Args[0])))
}

// RedirectHandler avoids the default behaviour of writing random html into the response.
func RedirectHandler(url string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeContentType(w, "text/plain;charset=UTF-8")
		http.Redirect(w, r, url, code)
	})
}

// ResourceHandler returns the asset, relative to the given paths.
func ResourceHandler(baseRequestPath, baseResourcePath string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path := strings.TrimPrefix(request.URL.Path, baseRequestPath)
		writeContentType(writer, detectTypeFromExtension(path))
		contents, err := Asset(baseResourcePath + "/" + path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.Write(contents)
		}
	})
}

func gcHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Forcing GC...")
	runtime.GC()
	log.Infof("GC completed")
	w.Write([]byte("GC complete"))
}

// HTTPServer is a holder for the router and server involved in serving the admin UI.
type HTTPServer struct {
	adminHTTPMuxer  *mux.Router
	allRoutes []Route
}

func (a *HTTPServer) addAdminRoutes(newRoutes ...Route) {
	for _, r := range newRoutes {
		method := r.method
		if method == "" {
			r.method = http.MethodGet
		}
		a.allRoutes = append(a.allRoutes, r)
	}

	// Some libraries like to register into DefaultServeMux, we serve all of them from the admin endpoint, so check if we missed any.
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.DefaultServeMux)
	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if s, ok := route.GetHandler().(*http.ServeMux); ok {
			muxEntries := reflect.ValueOf(s).Elem().FieldByName("m")
			for _, path := range muxEntries.MapKeys() {
				p := path.String()

				var found bool
				for _, r := range a.allRoutes {
					if p == r.path {
						found = true
					}
				}
				if !found {
					log.Warningf("Found entry %s in DefaultServeMux that isn't handled in AdminHTTPServer!", p)
				}
			}
		}

		return nil
	})

	a.updateMuxer()
}

func (a *HTTPServer) updateMuxer() {
	r := mux.NewRouter()

	for _, route := range a.allRoutes {
		handler := &indexView{
			title: route.alias,
			next:  route.handler,
			entries: func() entrySlice {
				return a.indexEntries()
			},
		}

		if route.prefix {
			r.PathPrefix(route.path).Handler(handler).Methods(route.method).Name(route.alias)
		} else {
			r.Path(route.path).Handler(handler).Methods(route.method).Name(route.alias)
		}
	}

	endpoints := make(sort.StringSlice, 0, len(a.allRoutes))
	for _, route := range a.allRoutes {
		endpoints = append(endpoints, fmt.Sprintf("\t%s => %s", route.path, getFunctionName(route.handler)))
	}
	sort.Sort(endpoints)

	log.Debugf("AdminHttpServer Muxer endpoints:\n%s", strings.Join(endpoints, "\n"))

	a.adminHTTPMuxer = r
}

func (a *HTTPServer) indexEntries() []entry {
	entries := make([]entry, 0)
	// TODO(mike): Add clients, servers here
	entries = append(entries, a.localRoutes()...)

	return entries
}

func (a *HTTPServer) localRoutes() []entry {
	routes := make([]Route, 0)
	for _, r := range a.allRoutes {
		if r.includeInIndex {
			routes = append(routes, r)
		}
	}

	results := make([]entry, 0, len(routes))

	grouped := groupRoutesByGroup(routes)

	for g, routes := range grouped {
		links := make(entrySlice, 0, len(routes))
		for _, r := range routes {
			links = append(links, link{ID: r.alias, HRef: r.path, Method: r.method})
		}
		if g == "" {
			results = append(results, links...)
		} else {
			results = append(results, group{
				Name:  g,
				Links: links,
			})
		}
	}

	return results
}

func groupRoutesByGroup(routes []Route) map[string][]Route {
	results := make(map[string][]Route)

	for _, r := range routes {
		existing, ok := results[r.group]
		if !ok {
			results[r.group] = []Route{r}
		} else {
			results[r.group] = append(existing, r)
		}
	}

	return results
}

func (a *HTTPServer) startServer(opts Opts) {
	if opts.Logger != nil {
		log = opts.Logger
	}
	if opts.LogInfo != nil {
		loggerInfo = opts.LogInfo
	}
	if opts.Disabled {
		log.Infof("Not starting admin http")
		return
	}

	log.Infof("Serving admin http on :%d", opts.Port)
	log.Errorf("Failed to serve admin HTTP: %s", http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), a.adminHTTPMuxer))
}

func getFunctionName(i interface{}) string {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	default:
		return runtime.FuncForPC(v.Pointer()).Name()
	}
}

var once sync.Once

// Serve starts the HTTPServer.
func Serve(opts Opts) {
	DefaultAdminHTTPServer.addAdminRoutes(routes...)
	DefaultAdminHTTPServer.startServer(opts)
}

// ServeOnce starts the HTTPServer, but only once.
func ServeOnce(opts Opts) {
	once.Do(func() {
		Serve(opts)
	})
}
