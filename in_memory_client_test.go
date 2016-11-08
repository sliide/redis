package redis_test

import (
	"fmt"
	"strconv"

	"github.com/sliide/redis"

	"time"

	. "gopkg.in/check.v1"
)

type InMemoryRedisTestSuite struct{}

var _ = Suite(
	&InMemoryRedisTestSuite{},
)

var inMemoryRedis redis.Client

func (s *InMemoryRedisTestSuite) SetUpSuite(c *C) {
	inMemoryRedis = redis.NewInMemoryClient()
}

func (s *InMemoryRedisTestSuite) TearDownSuite(c *C) {
	inMemoryRedis.Close()
}

func (s *InMemoryRedisTestSuite) TestIncrBy(c *C) {
	key := RandSeq(32)

	incrVal, err := inMemoryRedis.IncrBy(key, 10)
	c.Assert(err, IsNil)
	c.Assert(incrVal, Equals, int64(10))

	c.Assert(inMemoryRedis.Set(key, 1), IsNil)
	newVal, err := inMemoryRedis.IncrBy(key, 10)

	c.Assert(err, IsNil)
	c.Assert(newVal, Equals, int64(11))

	val, err := inMemoryRedis.Get(key)
	c.Assert(err, IsNil)

	c.Assert(val, Equals, strconv.Itoa(11))
}

func (s *InMemoryRedisTestSuite) TestExpire(c *C) {
	key := RandSeq(32)

	c.Assert(inMemoryRedis.Set(key, "1"), IsNil)

	val, err := inMemoryRedis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	c.Assert(inMemoryRedis.Expire(key, 1), IsNil)
	time.Sleep(2 * time.Second)

	val, err = inMemoryRedis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "")
}

func (s *InMemoryRedisTestSuite) TestSetEx(c *C) {
	key := RandSeq(32)

	c.Assert(inMemoryRedis.SetEx(key, 2, "1"), IsNil)

	val, err := inMemoryRedis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	time.Sleep(3 * time.Second)

	val, err = inMemoryRedis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "")
}

func (s *InMemoryRedisTestSuite) TestRPush(c *C) {
	key := RandSeq(32)

	for i := 0; i < 2; i++ {
		err := inMemoryRedis.RPush(key, strconv.Itoa(i))
		c.Assert(err, IsNil)
	}

	vals, err := inMemoryRedis.LRange(key)
	c.Assert(err, IsNil)
	c.Assert(vals, DeepEquals, []string{"0", "1"})
}

func (s *InMemoryRedisTestSuite) TestRedis(c *C) {

	key := RandSeq(32)
	pop := RandSeq(32)
	val := RandSeq(32)
	val2 := RandSeq(32)
	val3 := RandSeq(32)

	key2 := RandSeq(32)

	v, err := inMemoryRedis.Get(key)
	c.Assert(err, Not(IsNil))

	err = inMemoryRedis.Set(key, val)
	c.Assert(err, IsNil)

	v, err = inMemoryRedis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	v, err = inMemoryRedis.Pop(pop)
	c.Assert(err, Not(IsNil))

	err = inMemoryRedis.LPush(pop, val)
	c.Assert(err, IsNil)

	err = inMemoryRedis.LPush(pop, val2)
	c.Assert(err, IsNil)

	err = inMemoryRedis.LPush(pop, val3)
	c.Assert(err, IsNil)

	v, err = inMemoryRedis.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val3)

	v, err = inMemoryRedis.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val2)

	v, err = inMemoryRedis.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	err = inMemoryRedis.Set(key2, "2")
	c.Assert(err, IsNil)

	err = inMemoryRedis.Incr(key2)
	c.Assert(err, IsNil)

	err = inMemoryRedis.Incr(key2)
	c.Assert(err, IsNil)

	err = inMemoryRedis.Incr(key2)
	c.Assert(err, IsNil)

	v, err = inMemoryRedis.Get(key2)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "5")
}

func (s *InMemoryRedisTestSuite) TestGetNonExistentKey(c *C) {
	v, err := inMemoryRedis.Get("NotExsting")
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
}

func (s *InMemoryRedisTestSuite) TestMGet(c *C) {

	keys := []string{}
	for i := 0; i < 5; i++ {
		key := RandSeq(10)
		inMemoryRedis.Set(key, fmt.Sprintf("%d", i))
		keys = append(keys, key)
	}

	values, err := inMemoryRedis.MGet(keys)
	c.Assert(err, IsNil)

	expectedValues := []string{}
	for _, key := range keys {
		val, err := inMemoryRedis.Get(key)
		c.Assert(err, IsNil)
		expectedValues = append(expectedValues, val)
	}

	c.Assert(len(values), Equals, 5)
	for i := 0; i < 5; i++ {
		c.Assert(values[i], Equals, expectedValues[i])
	}
}

func (s *InMemoryRedisTestSuite) TestMGetWIthFailedKeys(c *C) {
	keyValMap := map[string]string{
		RandSeq(10): RandSeq(10),
		RandSeq(10): RandSeq(10),
	}
	keys := []string{}

	for key, val := range keyValMap {
		keys = append(keys, key)

		inMemoryRedis.Set(key, val)
	}

	keys = append(keys, "THISDOESNOTEXIST")

	values, err := inMemoryRedis.MGet(keys)

	c.Assert(err, IsNil)

	c.Assert(len(values), Equals, 3)
	c.Assert(values[0], Not(Equals), "")
	c.Assert(values[1], Not(Equals), "")
	c.Assert(values[2], Equals, "")
}

func (s *InMemoryRedisTestSuite) TestZAdd(c *C) {
	key := RandSeq(32)
	count, err := inMemoryRedis.ZAdd(key, 0.0, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = inMemoryRedis.ZCount(key, "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *InMemoryRedisTestSuite) TestZCount(c *C) {
	key := RandSeq(32)
	_, err := inMemoryRedis.ZAdd(key, -1.0, "a")
	c.Assert(err, IsNil)
	_, err = inMemoryRedis.ZAdd(key, 1.0, "b")
	c.Assert(err, IsNil)

	count, err := inMemoryRedis.ZCount(key, 0, "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = inMemoryRedis.ZCount(key, "-inf", 0)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = inMemoryRedis.ZCount(key, -10, 10)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *InMemoryRedisTestSuite) TestZCountNotExistentKey(c *C) {
	count, err := inMemoryRedis.ZCount("NotExisting", "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 0)
}
