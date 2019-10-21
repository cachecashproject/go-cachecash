package blockexplorer

import (
	"context"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/ledger"

	"github.com/cachecashproject/go-cachecash/ccmsg"

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

func newBlockExplorerServer(l *logrus.Logger, lc *ledgerservice.LedgerClient, conf *ConfigFile) (*http.Server, error) {
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
		format := format(c)
		if format == nil {
			return
		}
		doc := APIRoot{}
		doc.XLinks = &Links{}
		doc.XLinks.Links = make(map[string]*Link)
		doc.XLinks.Links["self"] = &Link{Href: makeHref(l, root, "/"), Name: "Self"}
		doc.XLinks.Links["blocks"] = &Link{Href: makeHref(l, root, "/blocks"), Name: "Blocks"}
		if *format == gin.MIMEJSON {
			c.JSON(200, doc)
		} else if *format == gin.MIMEHTML {
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
	r.GET("/blocks", func(c *gin.Context) {
		format := format(c)
		if format == nil {
			return
		}
		page, present := c.GetQuery("page")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		req := &ccmsg.GetBlocksRequest{StartDepth: -1}
		if present {
			pageToken, err := url.QueryUnescape(page)
			if err != nil {
				l.WithError(err).Error("bad token escaping")
				c.HTML(http.StatusInternalServerError, "error.gohtml", err)
				return
			}
			req.PageToken = []byte(pageToken)
		}
		rep, err := lc.GrpcClient.GetBlocks(ctx, req)
		if err != nil {
			// TODO capture err details
			l.WithError(err).Error("Failed to query ledgerd")
			c.HTML(http.StatusInternalServerError, "error.gohtml", err)
			return
		}
		doc := Blocks{}
		doc.XLinks = &Links{}
		doc.XLinks.Links = make(map[string]*Link)
		doc.XLinks.Links["self"] = &Link{Href: makeHref(l, root, "/blocks"), Name: "Self"}
		next := url.QueryEscape(string(rep.NextPageToken))
		doc.XLinks.Links["next"] = &Link{Href: makeHref(l, root, "/blocks?page="+next), Name: "Next"}
		prev := url.QueryEscape(string(rep.PrevPageToken))
		doc.XLinks.Links["prev"] = &Link{Href: makeHref(l, root, "/blocks?page="+prev), Name: "Prev"}
		// TODO: add a template link definition here for individual block lookup
		// doc.XLinks.Links["blocks"] = &Link{Href: makeHref(l, root, "/blocks/<id>"), Name: "Blocks"}
		doc.XEmbedded = &Blocks_Embedded{}
		doc.XEmbedded.Blocks = make([]*Block, 0, len(rep.Blocks))
		for _, ledger_block := range rep.Blocks {
			block := &Block{Data: ledger_block}
			block.XLinks = &Links{}
			block.XLinks.Links = make(map[string]*Link)
			block.XLinks.Links["self"] = &Link{
				Href: makeHref(l, root, "/blocks/"+ledger_block.BlockID().String()), Name: "Self"}
			doc.XEmbedded.Blocks = append(doc.XEmbedded.Blocks, block)
		}
		// TODO: factor this out for less duplication across types
		if *format == gin.MIMEJSON {
			c.JSON(200, doc)
		} else if *format == gin.MIMEHTML {
			c.HTML(http.StatusOK, "blocks.gohtml", &doc)
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

type flatTransactions struct {
	URL     string
	Meta    *ledger.Transaction
	Input   *ledger.TransactionInput
	Output  *ledger.TransactionOutput
	Witness *ledger.TransactionWitness
}

func flattenTransactions(transactions []*ledger.Transaction, block *Block) []*flatTransactions {
	result := make([]*flatTransactions, 0, len(transactions))
	for _, tx := range transactions {
		pos := len(result)
		row := &flatTransactions{}
		result = append(result, row)
		row.Meta = tx
		txid, err := tx.TXID()
		if err == nil {
			txidstr := txid.String()
			row.URL = block.XLinks.Links["self"].Href + "/tx/" + txidstr
		} else {
			row.URL = errors.Wrap(err, "Error getting transaction id").Error()
		}

		// accumulate inputs
		for idx, input := range tx.Inputs() {
			targetRow := pos + idx
			if targetRow == len(result) {
				row := &flatTransactions{}
				result = append(result, row)
			}
			tmp := input
			result[targetRow].Input = &tmp
		}
		// outputs
		for idx, output := range tx.Outputs() {
			targetRow := pos + idx
			if targetRow == len(result) {
				row := &flatTransactions{}
				result = append(result, row)
			}
			tmp := output
			result[targetRow].Output = &tmp
		}
		// witnesses
		for idx, witness := range tx.Witnesses() {
			targetRow := pos + idx
			// witnesses  always match length of inputs so this block doesn't
			// need an expansion case
			tmp := witness
			result[targetRow].Witness = &tmp
		}
	}
	return result
}

func loadTemplates(l *logrus.Logger) (*template.Template, error) {
	t := template.New("")
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"flattentransactions": flattenTransactions,
		"hex":                 hex.EncodeToString,
	}
	t.Funcs(funcMap)
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

// format calculates the format we should use
func format(c *gin.Context) *string {
	format := c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML, "application/protobuf")
	if len(format) == 0 {
		c.HTML(http.StatusNotAcceptable, "error.gohtml", nil)
		return nil
	}
	return &format
}
