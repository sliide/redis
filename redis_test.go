package redis

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
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

func TestRedisTestSuite(t *testing.T) {

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Failed to create docker client: %s", err)
	}

	resource, err := pool.Run("redis", "latest", []string{})
	if err != nil {
		t.Fatalf("Failed to run Redis docker container: %s", err)
	}

	defer func() {
		pool.Purge(resource)
	}()

	var client Client
	if err := pool.Retry(func() error {
		host := resource.GetBoundIP("6379/tcp")
		if h := os.Getenv("DOCKER_HOST"); h != "" {
			u, err := url.Parse(h)
			if err == nil {
				host = u.Hostname()
			}
		}

		port := resource.GetPort("6379/tcp")
		url := fmt.Sprintf("%s:%s", host, port)

		client = NewClient(url)
		if err := client.Ping(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Failed to connect to Redis docker container: %s", err)
	}

	// Actual
	suite.Run(t, &RedisTestSuite{
		client: client,
	})

	// Memory
	suite.Run(t, &RedisTestSuite{
		client: NewMemoryClient(),
	})
}

type RedisTestSuite struct {
	suite.Suite
	client Client
}

func (s *RedisTestSuite) TestGetSetDelMGetConcurrentAccess() {

	loopSize := 100
	key := RandSeq(32)

	var wg sync.WaitGroup
	wg.Add(loopSize * 4) // four testings

	go func() {
		for i := 0; i < loopSize; i++ {
			go func() {
				defer wg.Done()
				s.client.Set(key, 32)
			}()
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go func() {
				defer wg.Done()
				s.client.Del(key)
			}()
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go func() {
				defer wg.Done()
				s.client.Get(key)
			}()
		}
	}()

	go func() {
		for i := 0; i < loopSize; i++ {
			go func() {
				defer wg.Done()
				s.client.MGet(key, key, key, key)
			}()
		}
	}()

	wg.Wait()
}

func (s *RedisTestSuite) TestIncrBy() {
	key := RandSeq(32)

	incrVal, err := s.client.IncrBy(key, 10)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(10), incrVal)

	s.Assert().NoError(s.client.Set(key, 1))

	newVal, err := s.client.IncrBy(key, 10)
	s.Assert().NoError(err)

	s.Assert().Equal(int64(11), newVal)

	val, err := s.client.Get(key)
	s.Assert().NoError(err)
	s.Assert().Equal(strconv.Itoa(11), val)
}

func (s *RedisTestSuite) TestIncrByFloat() {
	key := RandSeq(32)

	incrVal, err := s.client.IncrByFloat(key, 10.5)
	s.Assert().NoError(err)
	s.Assert().Equal(float64(10.5), incrVal)

	s.Assert().NoError(s.client.Set(key, 1.1))
	newVal, err := s.client.IncrByFloat(key, 10.5)

	s.Assert().NoError(err)

	s.Assert().Equal(float64(11.6), newVal)

	val, err := s.client.Get(key)
	s.Assert().NoError(err)

	n, ok := numberToFloat64(val)
	s.Assert().True(ok)
	s.Assert().Equal(11.6, n)
}

func (s *RedisTestSuite) TestExpire() {
	key := RandSeq(32)

	s.Assert().NoError(s.client.Set(key, "1"))

	val, err := s.client.Get(key)
	s.Assert().NoError(err)
	s.Assert().Equal("1", val)

	set, err := s.client.Expire(key, 1)
	s.Assert().NoError(err)
	s.Assert().True(set)

	time.Sleep(2 * time.Second)

	val, err = s.client.Get(key)
	s.Assert().Error(err)
	s.Assert().Empty(val)
}

func (s *RedisTestSuite) TestExpireNotExistingKey() {
	key := RandSeq(32)

	set, err := s.client.Expire(key, 1)
	s.Assert().False(set)
	s.Assert().NoError(err)
}

func (s *RedisTestSuite) TestSetEx() {
	key := RandSeq(32)

	s.Assert().NoError(s.client.SetEx(key, 2, "1"))

	val, err := s.client.Get(key)
	s.Assert().NoError(err)
	s.Assert().Equal("1", val)

	time.Sleep(3 * time.Second)

	val, err = s.client.Get(key)
	s.Assert().Error(err)
	s.Assert().Empty(val)
}

func (s *RedisTestSuite) TestRPush() {
	key := RandSeq(32)

	for i := 0; i < 2; i++ {
		count, err := s.client.RPush(key, strconv.Itoa(i))
		s.Assert().Equal(int64(i+1), count)
		s.Assert().NoError(err)
	}

	vals, err := s.client.LRange(key)
	s.Assert().NoError(err)
	s.Assert().Equal([]string{"0", "1"}, vals)
}

func (s *RedisTestSuite) TestRedis() {

	key := RandSeq(32)
	pop := RandSeq(32)
	val := RandSeq(32)
	val2 := RandSeq(32)
	val3 := RandSeq(32)

	key2 := RandSeq(32)

	v, err := s.client.Get(key)
	s.Assert().Error(err)

	err = s.client.Set(key, val)
	s.Assert().NoError(err)

	v, err = s.client.Get(key)
	s.Assert().NoError(err)
	s.Assert().Equal(val, v)

	v, err = s.client.LPop(pop)
	s.Assert().Error(err)

	count, err := s.client.LPush(pop, val)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.LPush(pop, val2)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)

	count, err = s.client.LPush(pop, val3)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(3), count)

	v, err = s.client.LPop(pop)
	s.Assert().NoError(err)
	s.Assert().Equal(val3, v)

	v, err = s.client.LPop(pop)
	s.Assert().NoError(err)
	s.Assert().Equal(val2, v)

	v, err = s.client.LPop(pop)
	s.Assert().NoError(err)
	s.Assert().Equal(val, v)

	err = s.client.Set(key2, "2")
	s.Assert().NoError(err)

	value, err := s.client.Incr(key2)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(3), value)

	value, err = s.client.Incr(key2)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(4), value)

	value, err = s.client.Incr(key2)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(5), value)

	v, err = s.client.Get(key2)
	s.Assert().NoError(err)
	s.Assert().Equal("5", v)
}

func (s *RedisTestSuite) TestGetNonExistentKey() {
	v, err := s.client.Get("NotExsting")
	s.Assert().Error(err)
	s.Assert().Empty(v)
}

func (s *RedisTestSuite) TestMGet() {

	keys := []string{}
	for i := 0; i < 5; i++ {
		key := RandSeq(10)
		s.client.Set(key, fmt.Sprintf("%d", i))
		keys = append(keys, key)
	}

	values, err := s.client.MGet(keys...)
	s.Assert().NoError(err)

	expectedValues := []string{}
	for _, key := range keys {
		val, err := s.client.Get(key)
		s.Assert().NoError(err)
		expectedValues = append(expectedValues, val)
	}

	s.Assert().Equal(5, len(values))
	for i := 0; i < 5; i++ {
		s.Assert().Equal(expectedValues[i], values[i])
	}
}

func (s *RedisTestSuite) TestMGetWithFailedKeys() {
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

	s.Assert().NoError(err)

	s.Assert().Equal(3, len(values))
	s.Assert().NotEmpty(values[0])
	s.Assert().NotEmpty(values[1])
	s.Assert().Empty(values[2])
}

func (s *RedisTestSuite) TestZAdd() {
	key := RandSeq(32)
	count, err := s.client.ZAdd(key, 0.0, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.ZCount(key, "-inf", "+inf")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.ZAdd(key, 0.0, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(0), count)
}

func (s *RedisTestSuite) TestZCount() {
	key := RandSeq(32)
	_, err := s.client.ZAdd(key, -1.0, "a")
	s.Assert().NoError(err)
	_, err = s.client.ZAdd(key, 1.0, "b")
	s.Assert().NoError(err)

	count, err := s.client.ZCount(key, 0, "+inf")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.ZCount(key, "-inf", 0)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.ZCount(key, -10, 10)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)
}

func (s *RedisTestSuite) TestZCountNotExistentKey() {
	count, err := s.client.ZCount("NotExisting", "-inf", "+inf")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(0), count)
}

func (s *RedisTestSuite) TestSAdd() {
	key := RandSeq(32)
	count, err := s.client.SAdd(key, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.SAdd(key, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(0), count)

	count, err = s.client.SAdd(key, "a", "b")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)

	count, err = s.client.SAdd(key, "c", "d")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)

	err = s.client.Set(key, "string")
	s.Assert().NoError(err)

	_, err = s.client.SAdd(key, "a")
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestSMembers() {
	key := RandSeq(32)

	count, err := s.client.SAdd(key, "a", "b", "c", "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(3), count)

	members, err := s.client.SMembers(key)
	s.Assert().NoError(err)
	s.Assert().Len(members, 3)
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
	s.Assert().Equal(true, hasA)
	s.Assert().Equal(true, hasB)
	s.Assert().Equal(true, hasC)

	err = s.client.Set(key, "string")
	s.Assert().NoError(err)

	_, err = s.client.SMembers(key)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestSetResetsExpire() {
	key := RandSeq(32)
	s.client.SetEx(key, 1, 1)
	s.client.Set(key, 1)

	time.Sleep(2 * time.Second)
	v, err := s.client.Get(key)

	s.Assert().NoError(err)
	s.Assert().Equal("1", v)
}

func (s *RedisTestSuite) TestSetNXEX() {
	existingKey := RandSeq(32)
	s.client.Set(existingKey, 1)

	val, err := s.client.SetNxEx(existingKey, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(0), val)

	nonExistingKey := RandSeq(32)

	val, err = s.client.SetNxEx(nonExistingKey, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), val)

	time.Sleep(2 * time.Second)

	// Should have expire at this point
	val, err = s.client.SetNxEx(nonExistingKey, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), val)
}

func (s *RedisTestSuite) TestHSet() {
	key := RandSeq(32)
	field := "test"
	value := "unit"

	// returns true when field or key does not exist
	doesNotExist, err := s.client.HSet(key, field, value)
	s.Assert().NoError(err)
	s.Assert().Equal(true, doesNotExist)

	// stores value
	storedValue, err := s.client.HGet(key, field)
	s.Assert().NoError(err)
	s.Assert().Equal(value, storedValue)

	// returns false when field does exist
	doesNotExist, err = s.client.HSet(key, field, value)
	s.Assert().NoError(err)
	s.Assert().Equal(false, doesNotExist)

	// returns error when it is not a hash
	s.client.Set(key, value)
	doesNotExist, err = s.client.HSet(key, field, value)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHGet() {
	key := RandSeq(16)
	field := RandSeq(5)
	value := RandSeq(3)

	// key does not exist
	_, err := s.client.HGet(key, field)
	s.Assert().Equal(redis.ErrNil, err)

	// field does not exist on key
	s.client.HSet(key, "exists", 0)
	_, err = s.client.HGet(key, field)
	s.Assert().Equal(redis.ErrNil, err)

	// field exists on key
	s.client.HSet(key, field, value)
	v, err := s.client.HGet(key, field)
	s.Assert().NoError(err)
	s.Assert().Equal(value, v)

	// key is not a hash
	s.client.Set(key, value)
	_, err = s.client.HGet(key, field)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHMSet() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HSet(key, "original", "there")

	err := s.client.HMSet(key, values)
	s.Assert().NoError(err)
	actualValues, err := s.client.HGetAll(key)

	s.Assert().Equal(map[string]string{
		"original": "there",
		"a":        "1",
		"b":        "x",
	}, actualValues)

	s.client.Set(key, "garbage")
	err = s.client.HMSet(key, values)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHMGet() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a":     1,
		"b":     "x",
		"extra": "true",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HMGet(key, "a", "b")
	s.Assert().NoError(err)
	s.Assert().Equal(map[string]string{
		"a": "1",
		"b": "x",
	}, queriedValues)

	s.client.Set(key, "garbage")
	_, err = s.client.HMGet(key, "a", "b")
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHGetAll() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HGetAll(key)
	s.Assert().NoError(err)
	s.Assert().Equal(map[string]string{
		"a": "1",
		"b": "x",
	}, queriedValues)

	s.client.Set(key, "garbage")
	_, err = s.client.HGetAll(key)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHLen() {
	key := RandSeq(16)
	l, err := s.client.HLen(key)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(0), l)

	s.client.HSet(key, "1", "length")
	l, err = s.client.HLen(key)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), l)

	s.client.HSet(key, "2", "length")
	l, err = s.client.HLen(key)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), l)

	s.client.Set(key, "garbage")
	_, err = s.client.HLen(key)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHKeys() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	keys, err := s.client.HKeys(key)
	sort.StringSlice(keys).Sort()
	s.Assert().NoError(err)
	s.Assert().Equal([]string{"a", "b"}, keys)

	s.client.Set(key, "garbage")
	_, err = s.client.HKeys(key)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHVals() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
	}
	s.client.HMSet(key, values)

	queriedValues, err := s.client.HVals(key)
	sort.StringSlice(queriedValues).Sort()
	s.Assert().NoError(err)
	s.Assert().Equal([]string{"1", "x"}, queriedValues)

	s.client.Set(key, "garbage")
	_, err = s.client.HVals(key)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHScan() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"aaa": 1,
		"aba": "x",
	}
	s.client.HMSet(key, values)

	hash, err := s.client.HScan(key, "aa*")
	s.Assert().NoError(err)
	s.Assert().Len(hash, 1)
	s.Assert().Equal("1", hash["aaa"])
}

func (s *RedisTestSuite) TestHDel() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
		"b": "x",
		"c": true,
	}
	s.client.HMSet(key, values)

	count, err := s.client.HDel(key, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(1), count)
	l, err := s.client.HLen(key)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), l)

	count, err = s.client.HDel(key, "a", "b", "c")
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)

	s.client.Set(key, "garbage")
	_, err = s.client.HDel(key, "x")
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHExists() {
	key := RandSeq(16)
	values := map[string]interface{}{
		"a": 1,
	}
	s.client.HMSet(key, values)

	isExist, err := s.client.HExists(key, "a")
	s.Assert().NoError(err)
	s.Assert().Equal(true, isExist)

	isExist, err = s.client.HExists(key, "b")
	s.Assert().NoError(err)
	s.Assert().Equal(false, isExist)

	s.client.Set(key, "garbage")
	_, err = s.client.HExists(key, "x")
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHIncrBy() {
	key := RandSeq(16)
	field := RandSeq(3)
	values := map[string]interface{}{
		field: 1,
	}
	s.client.HMSet(key, values)

	value, err := s.client.HIncrBy(key, field, 10)
	s.Assert().NoError(err)
	s.Assert().Equal(int64(11), value)

	s.client.HSet(key, field, "1.1")
	_, err = s.client.HIncrBy(key, field, 1)
	s.Assert().Error(err)

	s.client.Set(key, "garbage")
	_, err = s.client.HIncrBy(key, "x", 4)
	s.Assert().Error(err)
}

func (s *RedisTestSuite) TestHIncrByFloat() {
	key := RandSeq(16)
	field := RandSeq(3)
	values := map[string]interface{}{
		field: 1,
	}
	s.client.HMSet(key, values)

	value, err := s.client.HIncrByFloat(key, field, 10.5)
	s.Assert().NoError(err)
	s.Assert().Equal(float64(11.5), value)

	s.client.HSet(key, field, "1.1")
	value, err = s.client.HIncrByFloat(key, field, 10.5)
	s.Assert().NoError(err)
	s.Assert().Equal(11.6, value)

	s.client.HSet(key, field, "a")
	_, err = s.client.HIncrByFloat(key, field, 1.1)
	s.Assert().Error(err)

	s.client.Set(key, "garbage")
	_, err = s.client.HIncrByFloat(key, "x", 1.5)
	s.Assert().Error(err)
}
