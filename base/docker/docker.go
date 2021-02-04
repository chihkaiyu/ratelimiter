package docker

import (
	"fmt"
	"net"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
)

type repoInfo struct {
	repo      string
	tag       string
	port      int
	env       []string
	isReady   func(host, port string) error
	clearFunc func(port string) error
}

var (
	runningContainers = []*dockertest.Resource{}
	RepoTags          = map[string]*repoInfo{
		"redis": &repoInfo{
			repo:      "redis",
			tag:       "6.0.10-alpine",
			port:      6379,
			env:       []string{},
			isReady:   DialPort,
			clearFunc: ClearRedis,
		},
	}
)

func RunExternal(repos []string) ([]string, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.WithField("err", err).Error("Could not connect to docker")
		return []string{}, err
	}

	ports := []string{}
	for _, r := range repos {
		info, ok := RepoTags[r]
		if !ok {
			logrus.WithField("repo", r).Warn("not defined repo")
			return []string{}, fmt.Errorf("unknown repo")
		}

		resource, err := pool.Run(info.repo, info.tag, info.env)
		if err != nil {
			logrus.WithField("err", err).Error("pool.Run failed")
			return []string{}, fmt.Errorf("repo run failed: %s", info.repo)
		}
		if err := pool.Retry(func() error {
			p := resource.GetPort(fmt.Sprintf("%d/tcp", info.port))
			if err := info.isReady("localhost", p); err != nil {
				return err
			}
			ports = append(ports, p)
			runningContainers = append(runningContainers, resource)
			return nil
		}); err != nil {
			logrus.Fatalf("Could not connect to docker: %s", err)
			return []string{}, fmt.Errorf("repo initialize failed: %s", info.repo)
		}
	}

	return ports, nil
}

func RemoveExternal() error {
	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.WithField("err", err).Error("Could not connect to docker")
		return err
	}

	for _, r := range runningContainers {
		if err := pool.Purge(r); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":  err,
				"repo": r.Container.ID,
			}).Error("pool.Purge failed")
			continue
		}
	}

	return nil
}

func DialPort(host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.Close()
	return nil
}

func ClearRedis(port string) error {
	context := ctx.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:" + port,
	})
	defer client.Close()

	if _, err := client.Ping(context).Result(); err != nil {
		context.WithField("err", err).Error("client.Ping failed")
		return err
	}

	if _, err := client.FlushAll(context).Result(); err != nil {
		context.WithField("err", err).Error("client.FlushAll failed")
		return err
	}

	return nil
}
