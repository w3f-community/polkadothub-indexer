package account

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getDetailsHttpHandler)(nil)
)

type getDetailsHttpHandler struct {
	client *client.Client

	useCase *getDetailsUseCase

	accountEraSeqDb store.AccountEraSeq
	eventSeqDb      store.EventSeq
}

func NewGetDetailsHttpHandler(c *client.Client, accountEraSeqDb store.AccountEraSeq, eventSeqDb store.EventSeq) *getDetailsHttpHandler {
	return &getDetailsHttpHandler{
		client: c,

		accountEraSeqDb: accountEraSeqDb,
		eventSeqDb:      eventSeqDb,
	}
}

type GetDetailsRequest struct {
	StashAccount string `uri:"stash_account" binding:"required"`
}

func (h *getDetailsHttpHandler) Handle(c *gin.Context) {
	var req GetDetailsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid stash account"))
		return
	}

	ds, err := h.getUseCase().Execute(req.StashAccount)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, ds)
}

func (h *getDetailsHttpHandler) getUseCase() *getDetailsUseCase {
	if h.useCase == nil {
		return NewGetDetailsUseCase(h.client, h.accountEraSeqDb, h.eventSeqDb)
	}
	return h.useCase
}
