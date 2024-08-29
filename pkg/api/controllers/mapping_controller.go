package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/to6ka/ledger/pkg/api/apierrors"
	"github.com/to6ka/ledger/pkg/core"
	"github.com/to6ka/ledger/pkg/ledger"
)

type MappingController struct{}

func NewMappingController() MappingController {
	return MappingController{}
}

func (ctl *MappingController) PutMapping(c *gin.Context) {
	l, _ := c.Get("ledger")

	mapping := &core.Mapping{}
	if err := c.ShouldBind(mapping); err != nil {
		apierrors.ResponseError(c, err)
		return
	}

	if err := l.(*ledger.Ledger).SaveMapping(c.Request.Context(), *mapping); err != nil {
		apierrors.ResponseError(c, err)
		return
	}

	respondWithData[*core.Mapping](c, http.StatusOK, mapping)
}

func (ctl *MappingController) GetMapping(c *gin.Context) {
	l, _ := c.Get("ledger")

	mapping, err := l.(*ledger.Ledger).LoadMapping(c.Request.Context())
	if err != nil {
		apierrors.ResponseError(c, err)
		return
	}

	respondWithData[*core.Mapping](c, http.StatusOK, mapping)
}
