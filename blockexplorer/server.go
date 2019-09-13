package blockexplorer

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opencensus.io/plugin/ochttp"
)

func newBlockExplorerServer(l *logrus.Logger, c *LedgerClient, conf *ConfigFile) (*http.Server, error) {
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

	return &http.Server{
		Addr: conf.HTTPAddr,
		Handler: &ochttp.Handler{
			Handler: r,
		},
	}, nil
}
