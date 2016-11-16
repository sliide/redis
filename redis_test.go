package redis_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/sliide/redis"

	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

// TODO: move it somewhere so we don't copy this everywhere
func RandSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type RedisTestSuite struct {
	client redis.Client
}

func (s *RedisTestSuite) TestIncrBy(c *C) {
	key := RandSeq(32)

	incrVal, err := s.client.IncrBy(key, 10)
	c.Assert(err, IsNil)
	c.Assert(incrVal, Equals, int64(10))

	c.Assert(s.client.Set(key, 1), IsNil)
	newVal, err := s.client.IncrBy(key, 10)

	c.Assert(err, IsNil)

	c.Assert(newVal, Equals, int64(11))

	val, err := s.client.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, strconv.Itoa(11))
}

func (s *RedisTestSuite) TestExpire(c *C) {
	key := RandSeq(32)

	c.Assert(s.client.Set(key, "1"), IsNil)

	val, err := s.client.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	set, err := s.client.Expire(key, 1)
	c.Assert(set, Equals, true)
	c.Assert(err, IsNil)
	time.Sleep(2 * time.Second)

	val, err = s.client.Get(key)
	c.Assert(err, Not(IsNil))
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestExpireNotExistingKey(c *C) {
	key := RandSeq(32)

	set, err := s.client.Expire(key, 1)
	c.Assert(set, Equals, false)
	c.Assert(err, IsNil)
}

func (s *RedisTestSuite) TestSetEx(c *C) {
	key := RandSeq(32)

	c.Assert(s.client.SetEx(key, 2, "1"), IsNil)

	val, err := s.client.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	time.Sleep(3 * time.Second)

	val, err = s.client.Get(key)
	c.Assert(err, Not(IsNil))
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestRPush(c *C) {
	key := RandSeq(32)

	for i := 0; i < 2; i++ {
		count, err := s.client.RPush(key, strconv.Itoa(i))
		c.Assert(count, Equals, int64(i + 1))
		c.Assert(err, IsNil)
	}

	vals, err := s.client.LRange(key)
	c.Assert(err, IsNil)
	c.Assert(vals, DeepEquals, []string{"0", "1"})
}

func (s *RedisTestSuite) TestRedis(c *C) {

	key := RandSeq(32)
	pop := RandSeq(32)
	val := RandSeq(32)
	val2 := RandSeq(32)
	val3 := RandSeq(32)

	key2 := RandSeq(32)

	v, err := s.client.Get(key)
	c.Assert(err, Not(IsNil))

	err = s.client.Set(key, val)
	c.Assert(err, IsNil)

	v, err = s.client.Get(key)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	v, err = s.client.LPop(pop)
	c.Assert(err, Not(IsNil))

	count, err := s.client.LPush(pop, val)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.LPush(pop, val2)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(2))

	count, err = s.client.LPush(pop, val3)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(3))

	v, err = s.client.LPop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val3)

	v, err = s.client.LPop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val2)

	v, err = s.client.LPop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	err = s.client.Set(key2, "2")
	c.Assert(err, IsNil)

	value, err := s.client.Incr(key2)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, int64(3))

	value, err = s.client.Incr(key2)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, int64(4))

	value, err = s.client.Incr(key2)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, int64(5))

	v, err = s.client.Get(key2)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "5")
}

func (s *RedisTestSuite) TestGetNonExistentKey(c *C) {
	v, err := s.client.Get("NotExsting")
	c.Assert(err, Not(IsNil))
	c.Assert(v, Equals, "")
}

func (s *RedisTestSuite) TestMGet(c *C) {

	keys := []string{}
	for i := 0; i < 5; i++ {
		key := RandSeq(10)
		s.client.Set(key, fmt.Sprintf("%d", i))
		keys = append(keys, key)
	}

	values, err := s.client.MGet(keys...)
	c.Assert(err, IsNil)

	expectedValues := []string{}
	for _, key := range keys {
		val, err := s.client.Get(key)
		c.Assert(err, IsNil)
		expectedValues = append(expectedValues, val)
	}

	c.Assert(len(values), Equals, 5)
	for i := 0; i < 5; i++ {
		c.Assert(values[i], Equals, expectedValues[i])
	}
}

func (s *RedisTestSuite) TestMGetWithFailedKeys(c *C) {
	keyValMap := map[string]string{
		RandSeq(10): RandSeq(10),
		RandSeq(10): RandSeq(10),
	}
	keys := []string{}

	for key, val := range keyValMap {
		keys = append(keys, key)

		s.client.Set(key, val)
	}

	keys = append(keys, "THISDOESNOTEXIST")

	values, err := s.client.MGet(keys...)

	c.Assert(err, IsNil)

	c.Assert(len(values), Equals, 3)
	c.Assert(values[0], Not(Equals), "")
	c.Assert(values[1], Not(Equals), "")
	c.Assert(values[2], Equals, "")
}

func (s *RedisTestSuite) TestZAdd(c *C) {
	key := RandSeq(32)
	count, err := s.client.ZAdd(key, 0.0, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.ZCount(key, "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.ZAdd(key, 0.0, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(0))
}

func (s *RedisTestSuite) TestZCount(c *C) {
	key := RandSeq(32)
	_, err := s.client.ZAdd(key, -1.0, "a")
	c.Assert(err, IsNil)
	_, err = s.client.ZAdd(key, 1.0, "b")
	c.Assert(err, IsNil)

	count, err := s.client.ZCount(key, 0, "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.ZCount(key, "-inf", 0)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.ZCount(key, -10, 10)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(2))
}

func (s *RedisTestSuite) TestZCountNotExistentKey(c *C) {
	count, err := s.client.ZCount("NotExisting", "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(0))
}
