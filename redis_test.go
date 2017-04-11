package redis

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
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
	client Client
}

func (s *RedisTestSuite) TestGetSetDelMGetConcurrentAccess(c *C) {

	loopSize := 100
	key := RandSeq(32)

	go func() {
		for i := 0; i < loopSize; i++ {
			go s.client.Set(key, 32)
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go s.client.Del(key)
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go s.client.Get(key)
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go s.client.MGet(key, key, key, key)
		}
	}()

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

func (s *RedisTestSuite) TestIncrByFloat(c *C) {
	key := RandSeq(32)

	incrVal, err := s.client.IncrByFloat(key, 10.5)
	c.Assert(err, IsNil)
	c.Assert(incrVal, Equals, float64(10.5))

	c.Assert(s.client.Set(key, 1.1), IsNil)
	newVal, err := s.client.IncrByFloat(key, 10.5)

	c.Assert(err, IsNil)

	c.Assert(newVal, Equals, float64(11.6))

	val, err := s.client.Get(key)
	c.Assert(err, IsNil)
	n, ok := NumberToFloat64(val)
	c.Assert(ok, Equals, true)
	c.Assert(n, Equals, 11.6)
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
		c.Assert(count, Equals, int64(i+1))
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

func (s *RedisTestSuite) TestSAdd(c *C) {
	key := RandSeq(32)
	count, err := s.client.SAdd(key, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.SAdd(key, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(0))

	count, err = s.client.SAdd(key, "a", "b")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))

	count, err = s.client.SAdd(key, "c", "d")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(2))

	err = s.client.Set(key, "string")
	c.Assert(err, IsNil)

	_, err = s.client.SAdd(key, "a")
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestSMembers(c *C) {
	key := RandSeq(32)

	count, err := s.client.SAdd(key, "a", "b", "c", "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(3))

	members, err := s.client.SMembers(key)
	c.Assert(err, IsNil)
	c.Assert(members, HasLen, 3)
	var hasA, hasB, hasC bool
	for _, v := range members {
		switch v {
		case "a":
			hasA = true
		case "b":
			hasB = true
		case "c":
			hasC = true
		}
	}
	c.Assert(hasA, Equals, true)
	c.Assert(hasB, Equals, true)
	c.Assert(hasC, Equals, true)

	err = s.client.Set(key, "string")
	c.Assert(err, IsNil)

	_, err = s.client.SMembers(key)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestSetResetsExpire(c *C) {
	key := RandSeq(32)
	s.client.SetEx(key, 1, 1)
	s.client.Set(key, 1)

	time.Sleep(2 * time.Second)
	v, err := s.client.Get(key)

	c.Assert(err, IsNil)
	c.Assert(v, Equals, "1")
}

func (s *RedisTestSuite) TestSetNXEX(c *C) {
	existingKey := RandSeq(32)
	s.client.Set(existingKey, 1)

	val, err := s.client.SetNxEx(existingKey, 1, 1)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, int64(0))

	nonExistingKey := RandSeq(32)

	val, err = s.client.SetNxEx(nonExistingKey, 1, 1)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, int64(1))

	time.Sleep(2 * time.Second)

	// Should have expire at this point
	val, err = s.client.SetNxEx(nonExistingKey, 1, 1)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, int64(1))
}

func (s *RedisTestSuite) TestHSet(c *C) {
	key := RandSeq(32)
	field := "test"
	value := "unit"

	// returns true when field or key does not exist
	doesNotExist, err := s.client.HSet(key, field, value)
	c.Assert(err, IsNil)
	c.Assert(doesNotExist, Equals, true)

	// stores value
	storedValue, err := s.client.HGet(key, field)
	c.Assert(err, IsNil)
	c.Assert(storedValue, Equals, value)

	// returns false when field does exist
	doesNotExist, err = s.client.HSet(key, field, value)
	c.Assert(err, IsNil)
	c.Assert(doesNotExist, Equals, false)

	// returns error when it is not a hash
	s.client.Set(key, value)
	doesNotExist, err = s.client.HSet(key, field, value)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHGet(c *C) {
	key := RandSeq(16)
	field := RandSeq(5)
	value := RandSeq(3)

	// key does not exist
	_, err := s.client.HGet(key, field)
	c.Assert(err, Equals, redis.ErrNil)

	// field does not exist on key
	s.client.HSet(key, "exists", 0)
	_, err = s.client.HGet(key, field)
	c.Assert(err, Equals, redis.ErrNil)

	// field exists on key
	s.client.HSet(key, field, value)
	v, err := s.client.HGet(key, field)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, value)

	// key is not a hash
	s.client.Set(key, value)
	_, err = s.client.HGet(key, field)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHMSet(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HSet(key, "original", "there")

	err := s.client.HMSet(key, values)
	c.Assert(err, IsNil)
	actualValues, err := s.client.HGetAll(key)

	c.Assert(actualValues, DeepEquals, map[string]string{
		"original": "there",
		"a":        "1",
		"b":        "x",
	})

	s.client.Set(key, "garbage")
	err = s.client.HMSet(key, values)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHMGet(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a":     1,
		"b":     "x",
		"extra": "true",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HMGet(key, "a", "b")
	c.Assert(err, IsNil)
	c.Assert(queriedValues, DeepEquals, map[string]string{
		"a": "1",
		"b": "x",
	})

	s.client.Set(key, "garbage")
	_, err = s.client.HMGet(key, "a", "b")
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHGetAll(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HGetAll(key)
	c.Assert(err, IsNil)
	c.Assert(queriedValues, DeepEquals, map[string]string{
		"a": "1",
		"b": "x",
	})

	s.client.Set(key, "garbage")
	_, err = s.client.HGetAll(key)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHLen(c *C) {
	key := RandSeq(16)
	l, err := s.client.HLen(key)
	c.Assert(err, IsNil)
	c.Assert(l, Equals, int64(0))

	s.client.HSet(key, "1", "length")
	l, err = s.client.HLen(key)
	c.Assert(err, IsNil)
	c.Assert(l, Equals, int64(1))

	s.client.HSet(key, "2", "length")
	l, err = s.client.HLen(key)
	c.Assert(err, IsNil)
	c.Assert(l, Equals, int64(2))

	s.client.Set(key, "garbage")
	_, err = s.client.HLen(key)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHKeys(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	keys, err := s.client.HKeys(key)
	sort.StringSlice(keys).Sort()
	c.Assert(err, IsNil)
	c.Assert(keys, DeepEquals, []string{"a", "b"})

	s.client.Set(key, "garbage")
	_, err = s.client.HKeys(key)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHVals(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HVals(key)
	sort.StringSlice(queriedValues).Sort()
	c.Assert(err, IsNil)
	c.Assert(queriedValues, DeepEquals, []string{"1", "x"})

	s.client.Set(key, "garbage")
	_, err = s.client.HVals(key)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHScan(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"aaa": 1,
		"aba": "x",
	}
	s.client.HMSet(key, values)

	hash, err := s.client.HScan(key, "aa*")
	c.Assert(err, IsNil)
	c.Assert(hash, HasLen, 1)
	c.Assert(hash["aaa"], Equals, "1")
}

func (s *RedisTestSuite) TestHDel(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
		"c": true,
	}
	s.client.HMSet(key, values)

	count, err := s.client.HDel(key, "a")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(1))
	l, err := s.client.HLen(key)
	c.Assert(err, IsNil)
	c.Assert(l, Equals, int64(2))

	count, err = s.client.HDel(key, "a", "b", "c")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(2))

	s.client.Set(key, "garbage")
	_, err = s.client.HDel(key, "x")
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHExists(c *C) {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
	}
	s.client.HMSet(key, values)

	isExist, err := s.client.HExists(key, "a")
	c.Assert(err, IsNil)
	c.Assert(isExist, Equals, true)

	isExist, err = s.client.HExists(key, "b")
	c.Assert(err, IsNil)
	c.Assert(isExist, Equals, false)

	s.client.Set(key, "garbage")
	_, err = s.client.HExists(key, "x")
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHIncrBy(c *C) {
	key := RandSeq(16)
	field := RandSeq(3)
	values := map[string]interface{}{
		field: 1,
	}
	s.client.HMSet(key, values)

	value, err := s.client.HIncrBy(key, field, 10)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, int64(11))

	s.client.HSet(key, field, "1.1")
	_, err = s.client.HIncrBy(key, field, 1)
	c.Assert(err, NotNil)

	s.client.Set(key, "garbage")
	_, err = s.client.HIncrBy(key, "x", 4)
	c.Assert(err, NotNil)
}

func (s *RedisTestSuite) TestHIncrByFloat(c *C) {
	key := RandSeq(16)
	field := RandSeq(3)
	values := map[string]interface{}{
		field: 1,
	}
	s.client.HMSet(key, values)

	value, err := s.client.HIncrByFloat(key, field, 10.5)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, float64(11.5))

	s.client.HSet(key, field, "1.1")
	value, err = s.client.HIncrByFloat(key, field, 10.5)
	c.Assert(err, IsNil)
	c.Assert(value, Equals, 11.6)

	s.client.HSet(key, field, "a")
	_, err = s.client.HIncrByFloat(key, field, 1.1)
	c.Assert(err, NotNil)

	s.client.Set(key, "garbage")
	_, err = s.client.HIncrByFloat(key, "x", 1.5)
	c.Assert(err, NotNil)
}
