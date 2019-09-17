package blockexplorer

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/ledgerservice"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	proto "github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opencensus.io/plugin/ochttp"
)

func newBlockExplorerServer(l *logrus.Logger, c *ledgerservice.LedgerClient, conf *ConfigFile) (*http.Server, error) {
	r := gin.New()
	r.Use(ginrus.Ginrus(l, time.RFC3339, true), gin.Recovery())
	if gin.Mode() != gin.TestMode {
		// Disabled in testing because it incorrectly handles duplicate metric
		// registrations:
		// time="2019-09-16T16:21:33+12:00" level=error msg="requests_total could not be registered in Prometheus" error="duplicate metrics collector registration attempted"
		p := ginprometheus.NewPrometheus("gin")
		// Keep url cardinality low in metrics:
		// TODO: for unrecognised routes (e.g. random paths collapse them, so that
		// attackers do not control our cardinality).
		p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
			url := c.Request.URL.Path
			for _, p := range c.Params {
				// TODO: test coverage: need to introspect the metric generated here
				if p.Key == "id" {
					url = strings.Replace(url, p.Value, ":id", 1)
					break
				}
			}
			return url
		}
		r.Use(p.HandlerFunc())
	}

	root, err := url.Parse(conf.Root)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse API root")
	}
	if !root.IsAbs() {
		return nil, errors.New("API root must be an absolute URL")
	}
	t, err := loadTemplates(l)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load templates")
	}
	r.SetHTMLTemplate(t)

	// Add routes here
	r.GET("/", func(c *gin.Context) {
		format := c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML, "application/protobuf")
		if len(format) == 0 {
			c.HTML(http.StatusNotAcceptable, "error.gohtml", nil)
			return
		}
		doc := APIRoot{}
		doc.XLinks = &Links{}
		doc.XLinks.Links = make(map[string]*Link)
		doc.XLinks.Links["self"] = &Link{Href: makeHref(l, root, "/"), Name: "Self"}
		doc.XLinks.Links["escrows"] = &Link{Href: makeHref(l, root, "/escrows"), Name: "Escrows"}
		if format == gin.MIMEJSON {
			c.JSON(200, doc)
		} else if format == gin.MIMEHTML {
			data := struct {
				Doc *APIRoot
			}{&doc}
			c.HTML(http.StatusOK, "index.gohtml", data)
		} else {
			// c.ProtoBuf sets the wrong content header, switch to c.Data in future.
			bytes, err := proto.Marshal(&doc)
			if err != nil {
				// TODO capture err details
				c.HTML(http.StatusInternalServerError, "error.gohtml", err)
				return
			}
			c.Data(http.StatusOK, "application/protobuf", bytes)
		}
	})

	return &http.Server{
		Addr: conf.HTTPAddr,
		Handler: &ochttp.Handler{
			Handler: r,
		},
	}, nil
}

func loadTemplates(l *logrus.Logger) (*template.Template, error) {
	t := template.New("")
	box := packr.NewBox("./html")
	err := box.Walk(func(name string, f packr.File) error {
		if !strings.HasSuffix(name, ".gohtml") {
			return nil
		}
		s := f.String()
		if len(s) == 0 {
			return errors.Errorf("zero length template %s", name)
		}
		_, err := t.New(name).Parse(s)
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// makeHref combines a references with the root
// reference must be a valid url fragment or this will panic
// (as this is only used with internal constant strings a panic
// interface is most convenient)
func makeHref(l *logrus.Logger, root *url.URL, reference string) string {
	ref_url, err := url.Parse(reference)
	if err != nil {
		l.WithError(err).Error("Bad url reference")
		panic("Bad url reference")
	}
	return root.ResolveReference(ref_url).String()
}
