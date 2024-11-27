package redis

import (
	"context"
	"errors"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	databases map[string]*redis.Client
	config    RedisConfig
}

func RedisDbId(name string) (int, bool) {
	switch name {
	case "APPL_DB":
		return 0, true
	case "COUNTERS_DB":
		return 2, true
	case "CONFIG_DB":
		return 4, true
	case "STATE_DB":
		return 6, true
	}

	return 0, false
}

type RedisConfig struct {
	Address  string `env:"REDIS_ADDRESS" env-default:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	Network  string `env:"REDIS_NETWORK" env-default:"tcp"`
}

func NewClient() (Client, error) {
	var cfg RedisConfig
	c := Client{}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return c, errors.New("failed to read redis config")
	}

	c.config = cfg
	c.databases = make(map[string]*redis.Client)

	return c, nil
}

func (c *Client) connect(dbName string) error {
	dbId, ok := RedisDbId(dbName)
	if ok {
		c.databases[dbName] = redis.NewClient(&redis.Options{
			Network:  c.config.Network,
			Addr:     c.config.Address,
			Password: c.config.Password,
			DB:       dbId,
		})
		return nil
	}

	return errors.New("database not defined")
}

func (c *Client) selectClient(dbName string) (*redis.Client, error) {
	var client *redis.Client

	_, ok := RedisDbId(dbName)

	if ok {
		client, ok = c.databases[dbName]

		if !ok {
			err := c.connect(dbName)
			if err != nil {
				return nil, err
			}

			client = c.databases[dbName]
		}

		return client, nil
	}

	return nil, errors.New("database not defined")
}

// Issue a HGETALL on key in a selected database
func (c *Client) HgetAllFromDb(ctx context.Context, dbName, key string) (map[string]string, error) {
	client, err := c.selectClient(dbName)
	if err != nil {
		return nil, err
	}

	data, err := client.HGetAll(ctx, key).Result()
	return data, err
}

func (c *Client) HsetToDb(ctx context.Context, dbName, key string, data map[string]string) error {
	client, err := c.selectClient(dbName)
	if err != nil {
		return err
	}

	client.HSet(ctx, key, data)

	return nil
}

func (c *Client) KeysFromDb(ctx context.Context, dbName, pattern string) ([]string, error) {
	client, err := c.selectClient(dbName)
	if err != nil {
		return nil, err
	}

	keys, err := client.Keys(ctx, pattern).Result()

	return keys, err
}

func (c *Client) Close() {
	for name, client := range c.databases {
		client.Close()
		delete(c.databases, name)
	}
}
