package redis

type DummyClient struct {}

func (dc DummyClient) Close() {
}

func (dc DummyClient) Get(key string) (val string, err error) {
	return "", nil
}

func (dc DummyClient) Set(key string, value interface{}) (err error) {
	return nil
}

func (dc DummyClient) SetEx(key string, expire int, value interface{}) (err error) {
	return nil
}

func (dc DummyClient) LPush(key string, value string) (err error) {
	return nil
}

func (dc DummyClient) RPush(key string, value string) (err error) {
	return nil
}

func (dc DummyClient) LRange(key string) (vals []string, err error) {
	return []string{}, nil
}

func (dc DummyClient) LPop(key string) (val string, err error) {
	return "", nil
}

func (dc DummyClient) Incr(key string) (err error) {
	return nil
}

func (dc DummyClient) IncrBy(key string, inc interface{}) (val interface{}, err error) {
	return nil, nil
}

func (dc DummyClient) Expire(key string, seconds int) (err error) {
	return nil
}

func (dc DummyClient) Del(key string) (err error) {
	return nil
}

func (dc DummyClient) MGet(keys []string) ([]string, error) {
	return []string{}, nil
}

func (dc DummyClient) ZAdd(key string, score float64, value interface{})  (int, error) {
	return 0, nil
}

func (dc DummyClient) ZCount(key string, min interface{}, max interface{}) (int, error) {
	return 0, nil
}
