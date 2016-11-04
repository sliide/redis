package redis_test

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/sliide/redis"

	"time"

	. "gopkg.in/check.v1"
)

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

func Test(t *testing.T) { TestingT(t) }

type RedisTestSuite struct{}

var _ = Suite(
	&RedisTestSuite{},
)

var pc redis.Client

func (s *RedisTestSuite) SetUpSuite(c *C) {
	pc = redis.NewClient("localhost:6379")
}

func (s *RedisTestSuite) TearDownSuite(c *C) {
	pc.Close()
}

func (s *RedisTestSuite) TestIncrBy(c *C) {
	key := RandSeq(32)

	incrVal, err := pc.IncrBy(key, 10)
	c.Assert(err, IsNil)
	c.Assert(incrVal, Equals, int64(10))

	c.Assert(pc.Set(key, 1), IsNil)
	newVal, err := pc.IncrBy(key, 10)

	c.Assert(err, IsNil)

	c.Assert(newVal, Equals, int64(11))

	val, err := pc.Get(key)

	if err != nil {
		log.Println(err)
		c.Fail()
	}

	c.Assert(val, Equals, strconv.Itoa(11))
}

func (s *RedisTestSuite) TestExpire(c *C) {
	key := RandSeq(32)

	c.Assert(pc.Set(key, "1"), IsNil)

	val, err := pc.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	c.Assert(pc.Expire(key, 1), IsNil)
	time.Sleep(2 * time.Second)

	val, err = pc.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestSetEx(c *C) {
	key := RandSeq(32)

	c.Assert(pc.SetEx(key, 2, "1"), IsNil)

	val, err := pc.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "1")

	time.Sleep(3 * time.Second)

	val, err = pc.Get(key)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "")
}

func (s *RedisTestSuite) TestRPush(c *C) {
	key := RandSeq(32)

	for i := 0; i < 2; i++ {
		err := pc.RPush(key, strconv.Itoa(i))
		c.Assert(err, IsNil)
	}

	vals, err := pc.LRange(key)
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

	v, err := pc.Get(key)
	c.Assert(err, IsNil)

	err = pc.Set(key, val)
	c.Assert(err, IsNil)

	v, err = pc.Get(key)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	v, err = pc.Pop(pop)
	c.Assert(err, Not(IsNil))

	err = pc.LPush(pop, val)
	c.Assert(err, IsNil)

	err = pc.LPush(pop, val2)
	c.Assert(err, IsNil)

	err = pc.LPush(pop, val3)
	c.Assert(err, IsNil)

	v, err = pc.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val3)

	v, err = pc.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val2)

	v, err = pc.Pop(pop)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, val)

	err = pc.Set(key2, "2")
	c.Assert(err, IsNil)

	err = pc.Incr(key2)
	c.Assert(err, IsNil)

	err = pc.Incr(key2)
	c.Assert(err, IsNil)

	err = pc.Incr(key2)
	c.Assert(err, IsNil)

	v, err = pc.Get(key2)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "5")
}

func (s *RedisTestSuite) TestGetNonExistentKey(c *C) {
	v, err := pc.Get("NotExsting")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "")
}

func (s *RedisTestSuite) TestMGet(c *C) {

	keys := []string{}
	for i := 0; i < 5; i++ {
		key := RandSeq(10)
		pc.Set(key, fmt.Sprintf("%d", i))
		keys = append(keys, key)
	}

	values, err := pc.MGet(keys)
	c.Assert(err, IsNil)

	expectedValues := []string{}
	for _, key := range keys {
		val, err := pc.Get(key)
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
		RandSeq(10): RandSeq(10),
		RandSeq(10): RandSeq(10),
	}
	keys := []string{}

	for key, val := range keyValMap {
		keys = append(keys, key)

		pc.Set(key, val)
	}

	keys = append(keys, "THISDOESNOTEXIST")

	values, err := pc.MGet(keys)

	c.Assert(err, IsNil)

	c.Assert(len(values), Equals, 3)
	c.Assert(values[0], Not(Equals), "")
	c.Assert(values[1], Not(Equals), "")
	c.Assert(values[2], Equals, "")
}

func (s *RedisTestSuite) TestZAdd(c *C) {
	key := RandSeq(32)
	count, err := pc.ZAdd(key, 0.0, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = pc.ZCount(key, "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *RedisTestSuite) TestZCount(c *C) {
	key := RandSeq(32)
	_, err := pc.ZAdd(key, -1.0, "a")
	c.Assert(err, IsNil)
	_, err = pc.ZAdd(key, 1.0, "b")
	c.Assert(err, IsNil)

	count, err := pc.ZCount(key, 0, "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = pc.ZCount(key, "-inf", 0)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)

	count, err = pc.ZCount(key, -10, 10)
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *RedisTestSuite) TestZCountNotExistentKey(c *C) {
	count, err := pc.ZCount("NotExisting", "-inf", "+inf")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 0)
}