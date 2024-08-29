package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/to6ka/ledger/pkg/api/apierrors"
	"github.com/to6ka/ledger/pkg/core"
	"github.com/to6ka/ledger/pkg/ledger"
	"github.com/to6ka/ledger/pkg/storage/sqlstorage"
)

type BalanceController struct{}

func NewBalanceController() BalanceController {
	return BalanceController{}
}

func (ctl *BalanceController) GetBalancesAggregated(c *gin.Context) {
	l, _ := c.Get("ledger")

	balancesQuery := ledger.NewAggregatedBalancesQuery().
		WithAddressFilter(c.Query("address"))
	balances, err := l.(*ledger.Ledger).GetBalancesAggregated(
		c.Request.Context(), *balancesQuery)
	if err != nil {
		apierrors.ResponseError(c, err)
		return
	}

	respondWithData[core.AssetsBalances](c, http.StatusOK, balances)
}

func (ctl *BalanceController) GetBalances(c *gin.Context) {
	l, _ := c.Get("ledger")

	balancesQuery := ledger.NewBalancesQuery()

	if c.Query(QueryKeyCursor) != "" {
		if c.Query("after") != "" ||
			c.Query("address") != "" ||
			c.Query(QueryKeyPageSize) != "" ||
			c.Query(QueryKeyPageSizeDeprecated) != "" {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("no other query params can be set with '%s'", QueryKeyCursor)))
			return
		}

		res, err := base64.RawURLEncoding.DecodeString(c.Query(QueryKeyCursor))
		if err != nil {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("invalid '%s' query param", QueryKeyCursor)))
			return
		}

		token := sqlstorage.BalancesPaginationToken{}
		if err := json.Unmarshal(res, &token); err != nil {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("invalid '%s' query param", QueryKeyCursor)))
			return
		}

		balancesQuery = balancesQuery.
			WithOffset(token.Offset).
			WithAfterAddress(token.AfterAddress).
			WithAddressFilter(token.AddressRegexpFilter...).
			WithPageSize(token.PageSize)

	} else if c.Query(QueryKeyCursorDeprecated) != "" {
		if c.Query("after") != "" ||
			c.Query("address") != "" ||
			c.Query(QueryKeyPageSize) != "" ||
			c.Query(QueryKeyPageSizeDeprecated) != "" {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("no other query params can be set with '%s'", QueryKeyCursorDeprecated)))
			return
		}

		res, err := base64.RawURLEncoding.DecodeString(c.Query(QueryKeyCursorDeprecated))
		if err != nil {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("invalid '%s' query param", QueryKeyCursorDeprecated)))
			return
		}

		token := sqlstorage.BalancesPaginationToken{}
		if err := json.Unmarshal(res, &token); err != nil {
			apierrors.ResponseError(c, ledger.NewValidationError(
				fmt.Sprintf("invalid '%s' query param", QueryKeyCursorDeprecated)))
			return
		}

		balancesQuery = balancesQuery.
			WithOffset(token.Offset).
			WithAfterAddress(token.AfterAddress).
			WithAddressFilter(token.AddressRegexpFilter...).
			WithPageSize(token.PageSize)

	} else {
		pageSize, err := getPageSize(c)
		if err != nil {
			apierrors.ResponseError(c, err)
			return
		}

		addresses := c.QueryArray("address")
		allAddresses := make([]string, 0)
		for _, address := range addresses {
			allAddresses = append(allAddresses, strings.Split(address, ",")...)
		}

		balancesQuery = balancesQuery.
			WithAfterAddress(c.Query("after")).
			WithAddressFilter(allAddresses...).
			WithPageSize(pageSize)
	}

	cursor, err := l.(*ledger.Ledger).GetBalances(c.Request.Context(), *balancesQuery)
	if err != nil {
		apierrors.ResponseError(c, err)
		return
	}

	respondWithCursor[core.AccountsBalances](c, http.StatusOK, cursor)
}
