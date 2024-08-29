package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/to6ka/ledger/pkg/api/apierrors"
	"github.com/to6ka/ledger/pkg/contextlogger"
	"github.com/to6ka/ledger/pkg/ledger"
	"github.com/to6ka/ledger/pkg/opentelemetry"
)

type LedgerMiddleware struct {
	resolver *ledger.Resolver
}

func NewLedgerMiddleware(resolver *ledger.Resolver) LedgerMiddleware {
	return LedgerMiddleware{
		resolver: resolver,
	}
}

func (m *LedgerMiddleware) LedgerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("ledger")
		if name == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		span := opentelemetry.WrapGinContext(c, "Ledger access")
		defer span.End()

		contextlogger.WrapGinRequest(c)

		l, err := m.resolver.GetLedger(c.Request.Context(), name)
		if err != nil {
			apierrors.ResponseError(c, err)
			return
		}
		defer l.Close(context.Background())

		c.Set("ledger", l)
		c.Next()
	}
}
