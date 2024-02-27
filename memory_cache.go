package services

import (
	"sync"
)

type MemoryCache struct {
	videoCache        map[string]*VideoResults
	channelCache      map[string]*ChannelInfo
	playlistCache     map[string]*VideoResults
	videoDetailsCache map[string]*VideoResults
	sync.Mutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		videoCache:        make(map[string]*VideoResults),
		channelCache:      make(map[string]*ChannelInfo),
		playlistCache:     make(map[string]*VideoResults),
		videoDetailsCache: make(map[string]*VideoResults),
	}
}

// GetVideo retrieves a video from Cache.
func (c *MemoryCache) GetVideo(key string) *VideoResults {
	c.Lock()
	defer c.Unlock()
	return c.videoCache[key]
}

// SetVideo stores a video to Cache.
func (c *MemoryCache) SetVideo(key string, video *VideoResults) {
	c.Lock()
	defer c.Unlock()
	c.videoCache[key] = video
}

// GetChannel retrieves a channel from Cache.
func (c *MemoryCache) GetChannel(key string) *ChannelInfo {
	c.Lock()
	defer c.Unlock()
	return c.channelCache[key]
}

// SetChannel stores a channel to Cache.
func (c *MemoryCache) SetChannel(key string, channel *ChannelInfo) {
	c.Lock()
	defer c.Unlock()
	c.channelCache[key] = channel
}

// GetPlaylist retrieves a playlist from Cache.
func (c *MemoryCache) GetPlaylist(key string) *VideoResults {
	c.Lock()
	defer c.Unlock()
	return c.playlistCache[key]
}

// SetPlaylist stores a playlist to Cache.
func (c *MemoryCache) SetPlaylist(key string, playlist *VideoResults) {
	c.Lock()
	defer c.Unlock()
	c.playlistCache[key] = playlist
}

// GetVideoDetail retrieves a VideoDetail from Cache.
func (c *MemoryCache) GetVideoDetail(key string) *VideoResults {
	c.Lock()
	defer c.Unlock()
	return c.videoDetailsCache[key]
}

// SetVideoDetail stores a VideoDetail to Cache.
func (c *MemoryCache) SetVideoDetail(key string, detail *VideoResults) {
	c.Lock()
	defer c.Unlock()
	c.videoDetailsCache[key] = detail
}

func (c *MemoryCache) GetServiceName() string {
	return "memory-cache"
}
