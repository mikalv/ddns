package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

type HostBackend interface {
	GetHost(string) (*Host, error)

	SetHost(*Host) error
}

type RedisBackend struct {
	expirationSeconds int
	pool              *redis.Pool
}

func NewRedisBackend(server string, expirationDays int) *RedisBackend {
	return &RedisBackend{
		expirationSeconds: expirationDays * 24 * 60 * 60,
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,

			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server)
				if err != nil {
					return nil, err
				}
				return c, err
			},

			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (r *RedisBackend) GetHost(name string) (*Host, error) {
	conn := r.pool.Get()
	defer conn.Close()

	host := Host{Hostname: name}

	var err error
	var data []interface{}

	if data, err = redis.Values(conn.Do("HGETALL", host.Hostname)); err != nil {
		return nil, err
	}

	if err = redis.ScanStruct(data, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

func (r *RedisBackend) SetHost(host *Host) error {
	conn := r.pool.Get()
	defer conn.Close()

	var err error

	if _, err = conn.Do("HMSET", redis.Args{}.Add(host.Hostname).AddFlat(host)...); err != nil {
		return err
	}

	if _, err = conn.Do("EXPIRE", host.Hostname, r.expirationSeconds); err != nil {
		return err
	}

	return nil
}

type Host struct {
	Hostname string `redis:"-"`
	Ip       string `redis:"ip"`
	Token    string `redis:"token"`
}

func (h *Host) GenerateAndSetToken() {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	hash.Write([]byte(h.Hostname))

	h.Token = fmt.Sprintf("%x", hash.Sum(nil))
}

// Returns true when this host has a IPv4 Address and false if IPv6
func (h *Host) IsIPv4() bool {
	if strings.Contains(h.Ip, ".") {
		return true
	}

	return false
}
