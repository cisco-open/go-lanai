package web

import (
	"context"
	"github.com/gin-gonic/gin"
)

// RecoveryCustomizer implements Customizer
type RecoveryCustomizer struct {

}

func NewRecoveryCustomizer() *RecoveryCustomizer {
	return &RecoveryCustomizer{}
}

func (c RecoveryCustomizer) Customize(ctx context.Context, r *Registrar) error {
	return r.AddGlobalMiddlewares(gin.Recovery())
}
