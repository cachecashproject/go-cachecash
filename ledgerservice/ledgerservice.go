package ledgerservice

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
)

type LedgerService struct {
	// The LedgerService knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l  *logrus.Logger
	db *sql.DB
	kp *keypair.KeyPair

	newTxChan *chan struct{}
}

func NewLedgerService(l *logrus.Logger, db *sql.DB, kp *keypair.KeyPair, newTxChan *chan struct{}) (*LedgerService, error) {
	s := &LedgerService{
		l:  l,
		db: db,
		kp: kp,

		newTxChan: newTxChan,
	}

	return s, nil
}

func (s *LedgerService) PostTransaction(ctx context.Context, req *ccmsg.PostTransactionRequest) (*ccmsg.PostTransactionResponse, error) {
	s.l.WithFields(logrus.Fields{"tx": req.Tx}).Info("PostTransaction")

	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin database transaction")
	}
	// defer dbTx.Close() ?

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tx into bytes")
	}

	txid, err := req.Tx.TXID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get txid")
	}
	txidBytes := types.BytesArray{0: txid[:]}

	mpTx := models.MempoolTransaction{
		Txid: txidBytes,
		Raw:  txBytes,
	}
	if err := mpTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert mempool transaction")
	}

	taTx := models.TransactionAuditlog{
		Txid:   txidBytes,
		Raw:    txBytes,
		Status: models.TransactionStatusPending,
	}
	if err := taTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction into audit log")
	}

	if err := dbTx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit database transaction")
	}

	if s.newTxChan != nil {
		s.l.Debug("Notifying mining go routine")
		*s.newTxChan <- struct{}{}
	}

	s.l.Info("PostTransaction - success")

	return &ccmsg.PostTransactionResponse{}, nil
}

// This is handed out in an opaque field for pagination: vary at will.
type pageToken struct {
	Forward bool           `json:"forward"`
	Height  int64          `json:"height"`
	BlockID ledger.BlockID `json:"blockID"`
}

type blockquery struct {
	qm []qm.QueryMod
	// does the result need to be reversed before sending (because reverse
	// pagination is in use)
	flipped bool
	// token that was used to generate the query - used for generating tokens
	// at the start and end of the view
	token *pageToken
}

// requestToQuery compiles a query for GetBlocks
// The query is suitable for stable pagination on BlockID with only one corner
// case - the addition of blocks within the page will push results out on
// refreshes.
// The structure on forward pagination is that higher rows are further through
// the result set.
// For reverse pagination, higher rows are earlier in the result set.
func requestToQuery(l *logrus.Logger, req *ccmsg.GetBlocksRequest) (*blockquery, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 50
	} else if req.Limit > 100 {
		return nil, errors.New("limit is too high")
	}

	query := []qm.QueryMod{qm.Limit(int(limit))}

	token := &pageToken{}
	if len(req.PageToken) > 0 {
		err := json.Unmarshal(req.PageToken, &token)
		if err != nil {
			return nil, errors.Wrap(err, "bad page token")
		}
		l.WithField("BlockID", token.BlockID).WithField("Forward", token.Forward).WithField(
			"Height", token.Height).Trace("pagination token")
	} else {
		token = nil
	}

	// we generate queries like so:
	// height < 61 or (height = 60 and block_id < abcdefg)
	flipped := false
	more_blocks := "(height = ? AND block_id > ?)"
	less_blocks := "(height = ? AND block_id < ?)"
	higher := "height > ?"
	lower := "height < ?"
	forward := "height ASC, block_id ASC"
	backwards := "height DESC, block_id DESC"
	if token != nil {
		if !token.Forward {
			forward, backwards = backwards, forward
			flipped = true
		}
	}
	if req.StartDepth < -1 {
		return nil, errors.New("invalid StartDepth")
	} else if req.StartDepth == -1 {
		more_blocks, less_blocks = less_blocks, more_blocks
		higher, lower = lower, higher
		forward = backwards
		// flipped = !flipped
	} else if token == nil && req.StartDepth > 0 {
		query = append(query, qm.Where("height >= ?", req.StartDepth))
	}

	l.WithField("forward", forward).WithField("blocks", more_blocks).WithField("higher", higher).WithField(
		"flipped", flipped).Trace("block query")

	if token != nil {
		if token.Forward {
			query = append(query, qm.Where(higher, token.Height), qm.Or(more_blocks, token.Height, token.BlockID[:]))
		} else {
			query = append(query, qm.Where(lower, token.Height), qm.Or(less_blocks, token.Height, token.BlockID[:]))
		}
	}

	query = append(query, qm.OrderBy(forward))

	return &blockquery{qm: query, flipped: flipped, token: token}, nil
}

func (s *LedgerService) GetBlocks(ctx context.Context, req *ccmsg.GetBlocksRequest) (*ccmsg.GetBlocksResponse, error) {
	s.l.Info("GetBlocks")

	query, err := requestToQuery(s.l, req)
	if err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	dbBlocks, err := models.Blocks(query.qm...).All(ctx, s.db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get blocks")
	}

	blocks := make([]*ledger.Block, 0, len(dbBlocks))
	for _, dbBlock := range dbBlocks {
		block := &ledger.Block{}
		err = block.Unmarshal(dbBlock.Raw)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse stored block")
		}
		blocks = append(blocks, block)
	}

	if query.flipped {
		for i := len(blocks)/2 - 1; i >= 0; i-- {
			opp := len(blocks) - 1 - i
			blocks[i], blocks[opp] = blocks[opp], blocks[i]
			dbBlocks[i], dbBlocks[opp] = dbBlocks[opp], dbBlocks[i]
		}
	}

	var nextToken []byte
	var prevToken []byte
	if len(blocks) != 0 {
		token := &pageToken{Forward: true, BlockID: blocks[len(blocks)-1].BlockID(), Height: int64(dbBlocks[len(dbBlocks)-1].Height)}
		nextToken, err = json.Marshal(&token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialise page token")
		}
		token = &pageToken{Forward: false, BlockID: blocks[0].BlockID(), Height: int64(dbBlocks[0].Height)}
		// NOTE: one may consider that we should not emit a previous token under
		// some circumstances, but as results can be inserted before the first
		// page began, the tokens are really cursors in a non-repeatable-read
		// database, and so there is no such thing as the 'first page', merely a
		// point at which scrolling in a given direction, at a point in time,
		// doesn't give new results.
		prevToken, err = json.Marshal(&token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialise page token")
		}
	} else if query.token != nil {
		if query.token.Forward {
			// return the original token for the same direction.
			nextToken = req.PageToken
			var token *pageToken
			var blockid_bytes [ledger.BlockIDSize]byte
			block_id := ledger.BlockID(blockid_bytes)
			if req.StartDepth == -1 {
				// lowest possible token
				token = &pageToken{Forward: false, BlockID: block_id, Height: int64(0)}
			} else {
				// take the current tokens height +1, guaranteed to include all
				// current results; and block_id becomes irrelevant
				token = &pageToken{Forward: false, BlockID: block_id, Height: query.token.Height + 1}
			}
			prevToken, err = json.Marshal(&token)
			if err != nil {
				return nil, errors.Wrap(err, "failed to serialise page token")
			}
		} else {
			// The first page of this query, whatever it is, is the default, so
			// an empty forward token
			nextToken = []byte("")
			prevToken = req.PageToken
		}
	} else {
		nextToken = []byte("")
		prevToken = []byte("")
	}

	s.l.WithFields(logrus.Fields{
		"blocks":     len(blocks),
		"startDepth": req.StartDepth,
		"limit":      req.Limit,
	}).Debug("sending block reply")

	return &ccmsg.GetBlocksResponse{
		Blocks:        blocks,
		NextPageToken: nextToken,
		PrevPageToken: prevToken,
	}, nil
}
