package redis_test

import (
	"github.com/sliide/redis"

	. "gopkg.in/check.v1"
)

type PoolClientTestSuite struct {
	RedisTestSuite
}

var _ = Suite(
	&PoolClientTestSuite{},
)

func (s *PoolClientTestSuite) SetUpSuite(c *C) {
	s.client = redis.NewClient("localhost:6379")
}
