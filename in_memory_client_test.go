package redis_test

import (
	"github.com/sliide/redis"

	. "gopkg.in/check.v1"
)

type InMemoryRedisTestSuite struct {
	RedisTestSuite
}

var _ = Suite(
	&InMemoryRedisTestSuite{},
)

func (s *InMemoryRedisTestSuite) SetUpSuite(c *C) {
	s.client = redis.NewInMemoryClient()
}
