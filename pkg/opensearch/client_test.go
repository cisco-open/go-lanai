package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go"
	"testing"
)

func TestOpenClientImpl_AddBeforeHook(t *testing.T) {
	type fields struct {
		client     *opensearch.Client
		beforeHook []BeforeHook
		afterHook  []AfterHook
	}
	type args struct {
		hook BeforeHook
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedLength int
	}{
		{
			name: "test that beforeHook grows by 1",
			fields: fields{
				client: &opensearch.Client{}, // doesn't really matter
				beforeHook: []BeforeHook{
					BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
						return ctx
					}),
				},
				afterHook: nil,
			},
			args: args{hook: BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
				return ctx
			})},
			expectedLength: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &OpenClientImpl{
				client:     tt.fields.client,
				beforeHook: tt.fields.beforeHook,
				afterHook:  tt.fields.afterHook,
			}
			c.AddBeforeHook(tt.args.hook)
			if tt.expectedLength != len(c.beforeHook) {
				t.Errorf("expected length of :%v, got :%v", tt.expectedLength, len(c.beforeHook))
			}
		})
	}
}

type SpecificHook struct {
	something string
}

func (s SpecificHook) Before(ctx context.Context, before BeforeContext) context.Context {
	return ctx
}

func TestOpenClientImpl_RemoveBeforeHook(t *testing.T) {
	specificHook := &SpecificHook{something: "hello"}
	type fields struct {
		client     *opensearch.Client
		beforeHook []BeforeHook
		afterHook  []AfterHook
	}
	type args struct {
		hook BeforeHook
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedLength int
	}{
		{
			name: "remove specific hook",
			fields: fields{
				client: &opensearch.Client{},
				beforeHook: []BeforeHook{
					BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
						return ctx
					}),
					specificHook,
				},
				afterHook: nil,
			},
			args:           args{specificHook},
			expectedLength: 1,
		},
		{
			name: "shouldn't remove any hook",
			fields: fields{
				client: &opensearch.Client{},
				beforeHook: []BeforeHook{
					BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
						return ctx
					}),
					specificHook,
				},
				afterHook: nil,
			},
			args:           args{&SpecificHook{something: "something else"}},
			expectedLength: 2,
		},
		{
			name: "removes hook skeleton",
			fields: fields{
				client: &opensearch.Client{},
				beforeHook: []BeforeHook{
					BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
						return ctx
					}),
					BeforeHookBase{
						Identifier: "abc",
						F: func(ctx context.Context, after BeforeContext) context.Context {
							return ctx
						},
					},
				},
				afterHook: nil,
			},
			args: args{
				BeforeHookBase{
					Identifier: "abc",
					F: func(ctx context.Context, after BeforeContext) context.Context {
						return ctx
					},
				},
			},
			expectedLength: 1,
		},
		{
			name: "should not remove hook skeleton",
			fields: fields{
				client: &opensearch.Client{},
				beforeHook: []BeforeHook{
					BeforeHookFunc(func(ctx context.Context, before BeforeContext) context.Context {
						return ctx
					}),
					BeforeHookBase{
						Identifier: "abc",
						F: func(ctx context.Context, after BeforeContext) context.Context {
							return ctx
						},
					},
				},
				afterHook: nil,
			},
			args: args{
				BeforeHookBase{
					Identifier: "abcd", // different identifier
					F: func(ctx context.Context, after BeforeContext) context.Context {
						return ctx
					},
				},
			},
			expectedLength: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &OpenClientImpl{
				client:     tt.fields.client,
				beforeHook: tt.fields.beforeHook,
				afterHook:  tt.fields.afterHook,
			}
			c.RemoveBeforeHook(tt.args.hook)
			if tt.expectedLength != len(c.beforeHook) {
				t.Errorf("expected length of :%v, got :%v", tt.expectedLength, len(c.beforeHook))
			}
		})
	}
}
