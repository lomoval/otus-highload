package service

import (
	"app/models"
	"github.com/OrlovEvgeny/go-mcache"
	timesortedlist "github.com/go-zen-chu/time-sorted-list"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
	"time"
)

var cache *mcache.CacheDriver
var newsLock = sync.Mutex{}

type listLockWrapper struct {
	lock sync.Mutex
	list timesortedlist.ITimeSortedList
}

func init() {
	cache = mcache.New()
}

func IsActiveUser(userID int64) bool {
	_, ok := cache.Get(strconv.FormatInt(userID, 10))
	return ok
}

func SetActiveUser(userID int64, timeout time.Duration) {
	cache.Set(strconv.FormatInt(userID, 10), nil, timeout)
}

func SetCacheNews(userID int64, value interface{}) {
	cache.Set(userNewsKey(userID), value, time.Hour*10)
}

func userNewsKey(userID int64) string {
	return strconv.FormatInt(userID, 10) + "news"
}

func RemoveCacheNews(userID int64) {
	cache.Remove(userNewsKey(userID))
}

func CachedNews(userID int64) []models.News {
	c, ok := cache.Get(userNewsKey(userID))
	if !ok {
		return nil
	}
	lockedList := c.(*listLockWrapper)

	news := make([]models.News, 0, lockedList.list.Len())
	items := lockedList.list.GetItemsUntil(time.Now().Unix())

	for i := len(items) - 1; i >= 0; i-- {
		news = append(news, (items[i].Item).(models.News))
	}
	return news
}

func cachedNews(userID int64) (interface{}, bool) {
	return cache.Get(strconv.FormatInt(userID, 10) + "news")
}

func AddCacheNews(userID int64, news []models.News) {
	cache, ok := cachedNews(userID)
	var maxNewsId int64
	if !ok {
		newsLock.Lock()
		cache, ok = cachedNews(userID)
		if !ok {
			dbNews, err := GetFriendsNews(userID)
			if err != nil {
				log.Err(err).Msgf("failed to get user nes to init cache")
				newsLock.Unlock()
				return
			}

			cache = &listLockWrapper{
				list: timesortedlist.NewTimeSortedList(1000),
			}
			SetCacheNews(userID, cache)
			lockedList := cache.(*listLockWrapper)
			for _, n := range dbNews {
				lockedList.list.AddItem(n.Timestamp.Unix(), n)
			}
			maxNewsId = dbNews[0].ID
		}
		newsLock.Unlock()
		if maxNewsId > 0 {
			var filteredNews []models.News

			for _, n := range news {
				if n.ID > maxNewsId {
					filteredNews = append(filteredNews, n)
				}
			}
			news = filteredNews
		}
	}

	lockedList := cache.(*listLockWrapper)
	lockedList.lock.Lock()
	defer lockedList.lock.Unlock()
	for _, n := range news {
		lockedList.list.AddItem(n.Timestamp.Unix(), n)
	}
}
