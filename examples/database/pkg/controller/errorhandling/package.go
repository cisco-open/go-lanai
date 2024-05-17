package errorhandling

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/cisco-open/go-lanai/pkg/web/weberror"
	"go.uber.org/fx"
)

// error pkg register

const (
	_ = order.Lowest - 0xffff + iota
	OrderDataIntegrityError
)

func Use() {
	bootstrap.AddOptions(
		fx.Invoke(registerErrorTranslators),
	)
}

func registerErrorTranslators(r *web.Registrar) {
	// data integrity errors, covers all APIs
	r.MustRegister(weberror.New("data integrity").
		Order(OrderDataIntegrityError).
		ApplyTo(matcher.RouteWithPattern("/api/**")).
		Use(TranslateDataIntegrityError).
		Build(),
	)
}
