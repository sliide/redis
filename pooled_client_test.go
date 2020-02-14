package redis

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

func TestPoolClientTestSuite(t *testing.T) {
	suite.Run(t, &PoolClientTestSuite{})
}

type PoolClientTestSuite struct {
	suite.Suite

	pool     *dockertest.Pool
	resource *dockertest.Resource
	client   Client
}

// TODO: Add more test

func (s *PoolClientTestSuite) SetupSuite() {
	t := s.T()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Failed to create docker client: %s", err)
	}

	resource, err := pool.Run("redis", "latest", []string{})
	if err != nil {
		t.Fatalf("Failed to run Redis docker container: %s", err)
	}

	s.pool = pool
	s.resource = resource

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

		client := NewClient(url)
		if err := client.Ping(); err != nil {
			return err
		}
		s.client = NewClient(url)
		return nil
	}); err != nil {
		pool.Purge(resource)
		t.Fatalf("Failed to connect to Redis docker container: %s", err)
	}
}

func (s *PoolClientTestSuite) TearDownSuite() {
	s.client.Close()
	s.pool.Purge(s.resource)
}

func (s PoolClientTestSuite) TestEval() {

	key := RandSeq(5)

	script := fmt.Sprintf("redis.call(\"SET\", \"%s\", 1); return 1;", key)

	val, err := s.client.Eval(script, 0)
	s.Require().NoError(err)

	one, ok := val.(int64)
	s.Assert().True(ok)
	s.Assert().Equal(int64(1), one)
}
