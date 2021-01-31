package core

import (
	"sync"
	"time"
)

type Cache struct {
	r *St

	maxCount int
	ttl      time.Duration

	m   map[string]*cacheVSt
	mMu sync.RWMutex
}

type cacheVSt struct {
	st time.Time

	name string
	mt   time.Time
	data []byte
}

func NewCache(r *St, maxCount int, ttl time.Duration) *Cache {
	c := &Cache{
		r:        r,
		maxCount: maxCount,
		ttl:      ttl,
		m:        map[string]*cacheVSt{},
	}

	if c.maxCount > 0 {
		go func() {
			for {
				time.Sleep(time.Minute)

				c.removeExpired()
			}
		}()
	}

	return c
}

func (c *Cache) GetAndRefresh(key string) (string, time.Time, []byte) {
	c.mMu.Lock()
	defer c.mMu.Unlock()

	now := time.Now()

	if cv, found := c.m[key]; found {
		cv.st = now
		return cv.name, cv.mt, cv.data
	}

	return "", now, nil
}

func (c *Cache) Set(key string, name string, mt time.Time, data []byte) {
	if c.maxCount == 0 {
		return
	}

	c.mMu.Lock()
	defer c.mMu.Unlock()

	now := time.Now()

	if cv, found := c.m[key]; found {
		cv.st = now
		cv.name = name
		cv.mt = mt
		cv.data = data
		return
	}

	c.m[key] = &cacheVSt{
		st:   now,
		name: name,
		mt:   mt,
		data: data,
	}

	if len(c.m) > c.maxCount {
		c.removeOldestOne()
	}
}

func (c *Cache) removeOldestOne() {
	var xK string
	var xV *cacheVSt

	for k, v := range c.m {
		if xV == nil || v.st.Before(xV.st) {
			xK = k
			xV = v
		}
	}

	if xK != "" {
		delete(c.m, xK)
	}
}

func (c *Cache) removeExpired() {
	c.mMu.Lock()
	defer c.mMu.Unlock()

	var expiredKeys []string

	now := time.Now()

	for k, v := range c.m {
		if v.st.Add(c.ttl).Before(now) {
			expiredKeys = append(expiredKeys, k)
		}
	}

	for _, k := range expiredKeys {
		delete(c.m, k)
	}
}
