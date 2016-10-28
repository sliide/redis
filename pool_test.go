package redis_test

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/sliide/redis"

	"math/rand"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type RedisTestSuite struct{}

var _ = Suite(
	&RedisTestSuite{},
)

func (s *RedisTestSuite) SetUpSuite(c *C) {
	redis.Init("localhost:6379")
}

func (s *RedisTestSuite) TearDownSuite(c *C) {
	redis.Close()
}

func (s *RedisTestSuite) TestIncrBy(c *C) {
	key := randSeq(32)

	incrVal, err := redis.IncrBy(key, 10)
	c.Assert(err, IsNil)
	c.Assert(incrVal, Equals, int64(10))

	c.Assert(redis.Set(key, 1), IsNil)
	newVal, err := redis.IncrBy(key, 10)

	c.Assert(err, IsNil)

	c.Assert(newVal, Equals, int64(11))

	val, err := redis.Get(key)

	if err != nil {
		log.Println(err)
		c.Fail()
	}

	c.Assert(val, Equals, strconv.Itoa(11))
}

func (s *RedisTestSuite) TestExpire(c *C) {
	key := randSeq(32)

	c.Assert(redis.Set(key, "1"), IsNil)

	val, err := redis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	c.Assert(redis.Expire(key, 1), IsNil)
	time.Sleep(2 * time.Second)

	val, err = redis.Get(key)
	c.Assert(err, Not(IsNil))
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestSetEx(c *C) {
	key := randSeq(32)

	c.Assert(redis.SetEx(key, 2, "1"), IsNil)

	val, err := redis.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	time.Sleep(3 * time.Second)

	val, err = redis.Get(key)
	c.Assert(err, Not(IsNil))
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestRPush(c *C) {
	key := randSeq(32)

	for i := 0; i < 2; i++ {
		err := redis.RPush(key, strconv.Itoa(i))
		c.Assert(err, Equals, nil)
	}

	vals, err := redis.LRange(key)
	c.Assert(err, IsNil)
	c.Assert(vals, DeepEquals, []string{"0", "1"})
}

func (s *RedisTestSuite) TestRedis(c *C) {

	key := randSeq(32)
	pop := randSeq(32)
	val := randSeq(32)
	val2 := randSeq(32)
	val3 := randSeq(32)

	key2 := randSeq(32)

	v, err := redis.Get(key)
	c.Assert(err, Not(Equals), nil)

	err = redis.Set(key, val)
	c.Assert(err, Equals, nil)

	v, err = redis.Get(key)
	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, val)

	v, err = redis.Pop(pop)
	c.Assert(err, Not(Equals), nil)

	err = redis.LPush(pop, val)
	c.Assert(err, Equals, nil)

	err = redis.LPush(pop, val2)
	c.Assert(err, Equals, nil)

	err = redis.LPush(pop, val3)
	c.Assert(err, Equals, nil)

	v, err = redis.Pop(pop)
	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, val3)

	v, err = redis.Pop(pop)
	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, val2)

	v, err = redis.Pop(pop)
	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, val)

	err = redis.Set(key2, "2")
	c.Assert(err, Equals, nil)

	err = redis.Incr(key2)
	c.Assert(err, Equals, nil)

	err = redis.Incr(key2)
	c.Assert(err, Equals, nil)

	err = redis.Incr(key2)
	c.Assert(err, Equals, nil)

	v, err = redis.Get(key2)
	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, "5")
}

func (s *RedisTestSuite) TestMGet(c *C) {

	keys := []string{}
	for i := 0; i < 5; i++ {
		key := randSeq(10)
		redis.Set(key, fmt.Sprintf("%d", i))
		keys = append(keys, key)
	}

	values, err := redis.MGet(keys)
	c.Assert(err, IsNil)

	expectedValues := []string{}
	for _, key := range keys {
		val, err := redis.Get(key)
		c.Assert(err, IsNil)
		expectedValues = append(expectedValues, val)
	}

	c.Assert(len(values), Equals, 5)
	for i := 0; i < 5; i++ {
		c.Assert(values[i], Equals, expectedValues[i])
	}
}

func (s *RedisTestSuite) TestMGetWIthFailedKeys(c *C) {
	keyValMap := map[string]string{
		randSeq(10): randSeq(10),
		randSeq(10): randSeq(10),
	}
	keys := []string{}

	for key, val := range keyValMap {
		keys = append(keys, key)

		redis.Set(key, val)
	}

	keys = append(keys, "THISDOESNOTEXIST")

	values, err := redis.MGet(keys)

	c.Assert(err, IsNil)

	c.Assert(len(values), Equals, 3)
	c.Assert(values[0], Not(Equals), "")
	c.Assert(values[1], Not(Equals), "")
	c.Assert(values[2], Equals, "")
}

func (s *RedisTestSuite) TestZAdd(c *C) {
	key := randSeq(32)
	count, err := redis.ZAdd(key, 0.0, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = redis.ZCount(key, "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *RedisTestSuite) TestZCount(c *C) {
	key := randSeq(32)
	_, err := redis.ZAdd(key, -1.0, "a")
	c.Assert(err, IsNil)
	_, err = redis.ZAdd(key, 1.0, "b")
	c.Assert(err, IsNil)

	count, err := redis.ZCount(key, 0, "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = redis.ZCount(key, "-inf", 0)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = redis.ZCount(key, -10, 10)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

// TODO: move it somewhere so we don't copy this everywhere
func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}