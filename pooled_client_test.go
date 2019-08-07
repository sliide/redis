package redis

import (
	"fmt"
	"log"

	"github.com/ory/dockertest"
	. "gopkg.in/check.v1"
)

type PoolClientTestSuite struct {
	RedisTestSuite

	pool     *dockertest.Pool
	resource *dockertest.Resource
}

var _ = Suite(
	&PoolClientTestSuite{},
)

func (s *PoolClientTestSuite) SetUpSuite(c *C) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Failed to create docker client: %s", err)
	}

	resource, err := pool.Run("redis", "latest", []string{})
	if err != nil {
		c.Fatalf("Failed to run Redis docker container: %s", err)
	}

	s.pool = pool
	s.resource = resource

	if err := pool.Retry(func() error {
		host := resource.GetBoundIP("6379/tcp")
		port := resource.GetPort("6379/tcp")
		url := fmt.Sprintf("%s:%s", host, port)

		s.client = NewClient(url)
		return nil
	}); err != nil {
		c.Fatalf("Failed to connect to Redis docker container: %s", err)
	}
}

func (s *PoolClientTestSuite) TearDownSuite(c *C) {
	s.client.Close()
	s.pool.Purge(s.resource)
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
