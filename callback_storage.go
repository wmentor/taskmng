package taskmng

import (
	"context"
)

type CallbackFunc func(context.Context, string)

type callbackStorage map[interface{}]CallbackFunc

func (cs callbackStorage) RegCallback(key interface{}, fn CallbackFunc) {
	cs[key] = fn
}

func (cs callbackStorage) PublishEvent(ctx context.Context, eventKey interface{}, msg string) {
	if fn, has := cs[eventKey]; has {
		fn(ctx, msg)
	}
}
