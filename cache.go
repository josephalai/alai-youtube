package services

import (
	"github.com/go-redis/redis"
	"time"
)

type Cache interface {
	// Get, Set for videoCache
	GetVideo(key string) *VideoResults
	SetVideo(key string, video *VideoResults)
	// Get, Set for channelCache
	GetChannel(key string) *ChannelInfo
	SetChannel(key string, channel *ChannelInfo)
	// Get, Set for playlistCache
	GetPlaylist(key string) *VideoResults
	SetPlaylist(key string, playlist *VideoResults)
	// Get, Set for videoDetailsCache
	GetVideoDetail(key string) *VideoResults
	SetVideoDetail(key string, detail *VideoResults)
	GetServiceName() string
}

type Redis interface {
	Ping() *redis.StatusCmd
	Get(string) *redis.StringCmd
	Set(string, interface{}, time.Duration) *redis.StatusCmd
}
