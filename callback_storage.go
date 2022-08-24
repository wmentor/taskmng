package taskmng

type callbackStorage map[interface{}]CallbackFunc

func (cs callbackStorage) RegCallback(key interface{}, fn CallbackFunc) {
	cs[key] = fn
}

func (cs callbackStorage) PublishEvent(key interface{}, msg string) {
	if fn, has := cs[key]; has {
		fn(msg)
	}
}
