package redis

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type PoolClientTestSuite struct {
	RedisTestSuite
}

var _ = Suite(
	&PoolClientTestSuite{},
)

func (s *PoolClientTestSuite) SetUpSuite(c *C) {
	s.client = NewClient("localhost:6379")
}

func (s *PoolClientTestSuite) TestEval(c *C) {

	key := RandSeq(5)

	script := fmt.Sprintf("redis.call(\"SET\", \"%s\", 1); return 1;", key)

	val, err := s.client.Eval(script, 0)
	c.Assert(err, IsNil)
	one, ok := val.(int64)
	c.Assert(ok, Equals, true)
	c.Assert(one, Equals, int64(1))
}
