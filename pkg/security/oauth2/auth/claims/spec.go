package claims

import "context"

type ClaimSpec interface {
	Calculate(ctx context.Context, opt *FactoryOption) (v interface{}, err error)
	Required(ctx context.Context, opt *FactoryOption) bool
}

type claimSpec struct {
	Func    ClaimFactoryFunc
	ReqFunc ClaimRequirementFunc
}

func (c claimSpec) Calculate(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if c.Func == nil {
		return nil, errorInvalidSpec
	}
	return c.Func(ctx, opt)
}

func (c claimSpec) Required(ctx context.Context, opt *FactoryOption) bool {
	if c.ReqFunc == nil {
		return false
	}
	return c.ReqFunc(ctx, opt)
}

func Required(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: requiredFunc,
	}
}

func Optional(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: optionalFunc,
	}
}

func RequiredIfParamsExists(fn ClaimFactoryFunc, requestParams ...string) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: func(ctx context.Context, opt *FactoryOption) bool {
			if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().Parameters() == nil {
				return false
			}
			req := opt.Source.OAuth2Request()
			for _, param := range requestParams {
				if _, ok := req.Parameters()[param]; ok {
					return true
				}
			}
			return false
		},
	}
}

func RequiredIfImplicitFlow(fn ClaimFactoryFunc) ClaimSpec {
	return &claimSpec{
		Func:    fn,
		ReqFunc: func(ctx context.Context, opt *FactoryOption) bool {
			if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().ResponseTypes() == nil {
				return false
			}
			return opt.Source.OAuth2Request().ResponseTypes().Has("token")
		},
	}
}

func Unsupported() ClaimSpec {
	return &claimSpec{
		Func: func(_ context.Context, _ *FactoryOption) (v interface{}, err error) {
			return nil, errorMissingDetails
		},
		ReqFunc: optionalFunc,
	}
}

func requiredFunc(_ context.Context, _ *FactoryOption) bool {
	return true
}

func optionalFunc(_ context.Context, _ *FactoryOption) bool {
	return false
}