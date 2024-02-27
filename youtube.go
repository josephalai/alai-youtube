package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/josephalai/alailog"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const SearchVideoIds = "https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=100&q=%s&type=video&order=date&relevanceLanguage=en&key=%s%v"
const GetTags = "https://www.googleapis.com/youtube/v3/videos?key=%s&fields=items(snippet(title,publishedAt,description,tags),id,statistics)&part=snippet,statistics&id=%v&order=date%v"
const GetChannelVideos = "https://www.googleapis.com/youtube/v3/channels/?part=snippet,contentDetails,statistics&id=%v&maxResults=50&key=%v"
const GetChannelPlaylist = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&maxResults=50&playlistId=%s&key=%s%s"

// YoutubeApi represents a service for interacting with the YouTube API.
type YoutubeApi struct {
	apiKey string
	Cache
}

type YoutubeService struct {
	Instance *YoutubeApi
	sync.Once
}

var youTubeServiceInstance = &YoutubeService{}

func GetInstance(optionalParams ...map[string]interface{}) *YoutubeApi {
	var opt map[string]interface{}
	var apiKey string
	var cache Cache = NewMemoryCache()
	if len(optionalParams) > 0 {
		opt = optionalParams[0]
		apiKey = opt["apiKey"].(string)
		if tCache, ok := opt["cache"].(Cache); ok {
			cache = tCache
		}
		log.Printf("api key set %s", apiKey)
	}
	youTubeServiceInstance.Do(func() {
		youTubeServiceInstance.Instance = NewYoutubeApi(apiKey, cache)
	})
	tags, err := youTubeServiceInstance.Instance.SearchAndRetrieveTags("alai")
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	log.Printf("tags: %v\n", tags)

	return youTubeServiceInstance.Instance
}

// NewYoutubeApi is now modified initialize the videoCache map
func NewYoutubeApi(apiKey string, cache Cache) *YoutubeApi {
	alailog.Printf("cache type: %s\n", cache.GetServiceName())
	return &YoutubeApi{
		apiKey: apiKey,
		Cache:  cache,
	}
}

func (yt *YoutubeApi) ApiKey() string {
	return yt.apiKey
}

// getChannelInfo queries the YouTube API for channel information using the given channel ID.
// It returns the channel information if found, otherwise returns an error.
// If the channel info is nil or has no items available, it returns an error.
func (yt *YoutubeApi) GetChannelInfo(channelId string) (*ChannelInfo, error) {
	if v := yt.Cache.GetChannel(channelId); v != nil {
		return v, nil
	}

	cInfo, err := getChannelInfo(channelId)
	if err != nil {
		return nil, errors.New("channel info not found")
	}
	if cInfo == nil || len(cInfo.Items) == 0 {
		return nil, errors.New("no item available in cInfo")
	}

	yt.Cache.SetChannel(channelId, cInfo)

	return cInfo, nil
}

// GetVideoCount converts the video count from string to integer and returns the result
// Parameters:
// - item: the item containing the video count value
// Returns:
// - int: the converted video count
// - error: an error message if there was an error converting the video count string to integer
func (yt *YoutubeApi) GetVideoCount(item *Item) (int, error) {
	vidCount, err := strconv.Atoi(item.Statistics.VideoCount)
	if err != nil {
		return 0, errors.New("internal server error")
	}

	return vidCount, nil
}

// getChannelPlaylist is a method of the YoutubeApi type that retrieves the playlist of videos for a given channel item.
// The method accepts an item pointer and a vidCount integer as parameters.
// If the item has non-nil ContentDetails and RelatedPlaylists, it calls the getChannelPlaylist function recursively with the uploads playlist ID and the vidCount value.
// If the getChannelPlaylist function returns an error, it returns an error with the message "internal server error".
// If the getChannelPlaylist function returns nil, it returns an error with the message "no results found".
// If the item's ContentDetails or RelatedPlaylists are nil, it returns an error with the message "contentDetails or RelatedPlaylists are nil".
func (yt *YoutubeApi) GetChannelPlaylist(item *Item, vidCount int) (*VideoResults, error) {
	cacheKey := item.Id + "-" + strconv.Itoa(vidCount)
	if v := yt.Cache.GetPlaylist(cacheKey); v != nil {
		return v, nil
	}

	if item.ContentDetails != nil && item.ContentDetails.RelatedPlaylists != nil {
		results, err := yt.getChannelPlaylist(item.ContentDetails.RelatedPlaylists.Uploads, vidCount)
		if err != nil {
			return nil, errors.New("internal server error")
		}
		if results == nil {
			return nil, errors.New("no results found")
		}

		// If no error and results obtained, add to cache
		yt.Cache.SetPlaylist(cacheKey, results)

		return results, nil
	} else {
		// If no error and results obtained, add to cache
		yt.Cache.SetPlaylist(cacheKey, nil)

		return nil, errors.New("contentDetails or RelatedPlaylists are nil")
	}
}

type TagSearchResults struct {
	Items []struct {
		Id *struct {
			VideoId string `bson:"videoId,omitempty" json:"videoId,omitempty"`
		} `bson:"id,omitempty" json:"id,omitempty"`
		Snippet *struct {
			PublishedAt  string     `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
			Title        string     `bson:"title,omitempty" json:"title,omitempty"`
			Description  string     `bson:"description,omitempty" json:"description,omitempty"`
			ChannelTitle string     `bson:"channelTitle,omitempty" json:"channelTitle,omitempty"`
			ChannelId    string     `bson:"channelId,omitempty" json:"channelId,omitempty"`
			Thumbnails   Thumbnails `bson:"thumbnails,omitempty" json:"thumbnails,omitempty"`
		} `bson:"snippet,omitempty" json:"snippet,omitempty"`
	} `bson:"items,omitempty" json:"items,omitempty"`
	NextPageToken string `bson:"nextPageToken,omitempty" json:"nextPageToken,omitempty"`
}

// Thumbnails represents different sizes of image URLs for a video
// The default thumbnail size
type Thumbnails struct {
	Default *struct {
		Url    string `bson:"url,omitempty" json:"url,omitempty"`
		Width  int    `bson:"width,omitempty" json:"width,omitempty"`
		Height int    `bson:"height,omitempty" json:"height,omitempty"`
	} `bson:"default,omitempty" json:"default,omitempty"`
	Medium *struct {
		Url    string `bson:"url,omitempty" json:"url,omitempty"`
		Width  int    `bson:"width,omitempty" json:"width,omitempty"`
		Height int    `bson:"height,omitempty" json:"height,omitempty"`
	} `bson:"medium,omitempty" json:"medium,omitempty"`
	High *struct {
		Url    string `bson:"url,omitempty" json:"url,omitempty"`
		Width  int    `bson:"width,omitempty" json:"width,omitempty"`
		Height int    `bson:"height,omitempty" json:"height,omitempty"`
	} `bson:"high,omitempty" json:"high,omitempty"`
}

// ChannelPlaylistVideoResults represents the results of a channel playlist video search.
// It contains information about the videos in the playlist, such as their ID, snippet, content details, and page information.
type ChannelPlaylistVideoResults struct {
	Items []struct {
		Id      string `bson:"id,omitempty" json:"id,omitempty"`
		Snippet *struct {
			PublishedAt  string     `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
			Title        string     `bson:"title,omitempty" json:"title,omitempty"`
			Description  string     `bson:"description,omitempty" json:"description,omitempty"`
			Thumbnails   Thumbnails `bson:"thumbnails,omitempty" json:"thumbnails,omitempty"`
			ChannelTitle string     `bson:"channelTitle,omitempty" json:"channelTitle,omitempty"`
		} `bson:"snippet,omitempty" json:"snippet,omitempty"`
		ContentDetails *struct {
			VideoId          string `bson:"videoId,omitempty" json:"videoId,omitempty"`
			VideoPublishedAt string `bson:"videoPublishedAt,omitempty" json:"videoPublishedAt,omitempty"`
		} `bson:"contentDetails,omitempty" json:"contentDetails,omitempty"`
	} `bson:"items,omitempty" json:"items,omitempty"`
	PageInfo *struct {
		TotalResults int `bson:"totalResults,omitempty" json:"totalResults,omitempty"`
	} `bson:"pageInfo,omitempty" json:"pageInfo,omitempty"`
	NextPageToken string `bson:"nextPageToken,omitempty" json:"nextPageToken,omitempty"`
}

// Item represents an item in a search result or playlist
// It contains various fields for the item's details such as ID, snippet, content details, and statistics.
// The ID field is a string that uniquely identifies the item.
// The Snippet field contains additional details about the item such as its published date, title, description,
// custom URL, channel title, thumbnails, localized title and description, and country.
// The thumbnails field contains different sizes of thumbnails for the item, including default, medium, and high.
// The Localized field contains localized title and description for the item.
// The Country field specifies the country of the item.
// The ContentDetails field contains additional details about the item's content,
// such as related playlists for likes and uploads.
// The Statistics field contains statistical information about the item, including view count,
// subscriber count, hidden subscriber count status, and video count.
type Item struct {
	Id      string `bson:"id,omitempty" json:"id,omitempty"`
	Snippet *struct {
		PublishedAt  string `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
		Title        string `bson:"title,omitempty" json:"title,omitempty"`
		Description  string `bson:"description,omitempty" json:"description,omitempty"`
		CustomUrl    string `bson:"customUrl,omitempty" json:"customUrl,omitempty"`
		ChannelTitle string `bson:"channelTitle,omitempty" json:"channelTitle,omitempty"`
		Thumbnails   struct {
			Default *struct {
				Url    string `bson:"url,omitempty" json:"url,omitempty"`
				Width  int    `bson:"width,omitempty" json:"width,omitempty"`
				Height int    `bson:"height,omitempty" json:"height,omitempty"`
			} `bson:"default,omitempty" json:"default,omitempty"`
			Medium *struct {
				Url    string `bson:"url,omitempty" json:"url,omitempty"`
				Width  int    `bson:"width,omitempty" json:"width,omitempty"`
				Height int    `bson:"height,omitempty" json:"height,omitempty"`
			} `bson:"medium,omitempty" json:"medium,omitempty"`
			High *struct {
				Url    string `bson:"url,omitempty" json:"url,omitempty"`
				Width  int    `bson:"width,omitempty" json:"width,omitempty"`
				Height int    `bson:"height,omitempty" json:"height,omitempty"`
			} `bson:"high,omitempty" json:"high,omitempty"`
		} `bson:"thumbnails,omitempty" json:"thumbnails,omitempty"`
		Localized *struct {
			Title       string `bson:"title,omitempty" json:"title,omitempty"`
			Description string `bson:"description,omitempty" json:"description,omitempty"`
		}
		Country string `bson:"country,omitempty" json:"country,omitempty"`
	} `bson:"snippet,omitempty" json:"snippet,omitempty"`
	ContentDetails *struct {
		RelatedPlaylists *struct {
			Likes   string `bson:"likes,omitempty" json:"likes,omitempty"`
			Uploads string `bson:"uploads,omitempty" json:"uploads,omitempty"`
		} `bson:"relatedPlaylists,omitempty" json:"relatedPlaylists,omitempty"`
	} `bson:"contentDetails,omitempty" json:"contentDetails,omitempty"`
	Statistics *struct {
		ViewCount             string `bson:"viewCount,omitempty" json:"viewCount,omitempty"`
		SubscriberCount       string `bson:"subscriberCount,omitempty" json:"subscriberCount,omitempty"`
		HiddenSubscriberCount bool   `bson:"hiddenSubscriberCount,omitempty" json:"hidden_subscriber_count,omitempty"`
		VideoCount            string `bson:"videoCount,omitempty" json:"videoCount,omitempty"`
	} `bson:"statistics,omitempty" json:"statistics,omitempty"`
}

// ChannelInfo contains information about a YouTube channel and its videos.
// It includes a list of Item objects and the next page token.
type ChannelInfo struct {
	Items         []*Item `bson:"items,omitempty" json:"items,omitempty"`
	NextPageToken string  `bson:"nextPageToken,omitempty" json:"nextPageToken,omitempty"`
}

// VideoResults contains the list of videos retrieved
type VideoResults struct {
	Items         []*Video `bson:"items,omitempty" json:"items,omitempty"`
	NextPageToken string   `bson:"nextPageToken,omitempty" json:"nextPageToken,omitempty"`
}

// Video represents a YouTube video.
type Video struct {
	Id string `bson:"id,omitempty" json:"id,omitempty"`

	Snippet *struct {
		ChannelId     string     `bson:"channelId,omitempty" json:"channelId,omitempty"`
		ChannelTitle  string     `bson:"channelTitle,omitempty" json:"channelTitle,omitempty"`
		PublishedAt   string     `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
		Title         string     `bson:"title,omitempty" json:"title,omitempty"`
		Description   string     `bson:"description,omitempty" json:"description,omitempty"`
		Thumbnails    Thumbnails `bson:"thumbnails,omitempty" json:"thumbnails,omitempty"`
		Tags          []string   `bson:"tags,omitempty" json:"tags,omitempty"`
		FormattedTags string     `bson:"formatted_tags,omitempty" json:"formatted_tags,omitempty"`
	} `bson:"snippet,omitempty" json:"snippet,omitempty"`

	Statistics *struct {
		ViewCount     string `bson:"viewCount,omitempty" json:"viewCount,omitempty"`
		LikeCount     string `bson:"likeCount,omitempty" json:"likeCount,omitempty"`
		DislikeCount  string `bson:"dislikeCount,omitempty" json:"dislikeCount,omitempty"`
		FavoriteCount string `bson:"favoriteCount,omitempty" json:"favoriteCount,omitempty"`
		CommentCount  string `bson:"commentCount,omitempty" json:"commentCount,omitempty"`
	} `bson:"statistics,omitempty" json:"statistics,omitempty"`
}

// MinViews is the minimum number of views required for a video to be included in the results of the `FindTags` function.
// Videos with view counts below the `MinViews` value will be filtered out.
// It is used to filter the `vidResults` by checking the `Statistics.ViewCount` field of each video and only including those with view counts greater than `MinViews`.
// Example usage:
// ```
// filteredItems := []*Video{}
//
//	for _, item := range vidResults.Items {
//	    if item.Statistics.ViewCount != "" {
//	        views, err := strconv.Atoi(item.Statistics.ViewCount)
//	        if err != nil {
//	            log.Printf("Failed to convert view count to integer, error: %v\n", err)
//	            return nil, err
//	        }
//	        if views > MinViews {
//	            // Append `item` to `filteredItems` if view count is greater than `MinViews`
//	            filteredItems = append(filteredItems, (*Video)(item))
//	        }
const MinViews int = 1000

// FindTags searches for videos on YouTube based on the input string and returns the videos along with their information.
// It takes the input string and the number of pages to search through as parameters.
// The function also accepts optional parameters as a map[string]interface{}.
//
// The videos are searched by replacing spaces in the input string with proper URL formatting.
// The nextPage variable is used to keep track of the next page of search results.
// The pageVar contains the formatting for the pageToken parameter in the API URL.
//
// The function defines a struct, VidSnippetInfo, to store information for aggregation with the search results.
// The information includes the channel title, channel ID, and thumbnails of the video.
// The vidIds map is used to store the video IDs as keys and the corresponding VidSnippetInfo structs as values.
//
// The function performs the search in a loop for the specified number of pages.
// It constructs the URL for the API request using the fSearch input, the API key, and the nextPageStr (if applicable).
// The response from the HTTP request
func (yt *YoutubeApi) FindTags(input string, numPages int, optionalParams ...map[string]interface{}) (*VideoResults, error) {
	// check if input already in videoCache and if so, return cached result
	if v := yt.Cache.GetVideo(input); v != nil {
		return v, nil
	}

	var videos = make([]string, 0)
	fSearch := strings.Replace(input, " ", "%20%", -1)
	nextPage := ""
	pageVar := "&pageToken=%v"

	type VidSnippetInfo struct {
		ChannelTitle string
		ChannelId    string
		Thumbnails   Thumbnails
	}
	vidIds := make(map[string]VidSnippetInfo)
	for i := 0; i < numPages; i++ {
		nextPageStr := ""
		if i > 0 {
			nextPageStr = fmt.Sprintf(pageVar, nextPage)
		}
		pageUrl := fmt.Sprintf(SearchVideoIds, fSearch, yt.ApiKey(), nextPageStr)

		resp, err := http.Get(pageUrl)
		if err != nil {
			log.Printf("Failed HTTP request, error: %v\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		log.Printf("GET %s status: %s\n", pageUrl, resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed reading body, error: %v\n", err)
			return nil, err
		}

		// log.Printf("Response body: %s\n", string(body))

		res := TagSearchResults{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			log.Printf("Error unmarshaling response to struct, error: %v\n", err)
			return nil, err
		}

		for _, vid := range res.Items {
			videos = append(videos, vid.Id.VideoId)
			vidIds[vid.Id.VideoId] = VidSnippetInfo{ChannelTitle: vid.Snippet.ChannelTitle, ChannelId: vid.Snippet.ChannelId, Thumbnails: vid.Snippet.Thumbnails}
		}
		nextPage = res.NextPageToken
		if nextPage == "" {
			break
		}
	}
	vidResults, err := yt.GetVideos(videos)
	if err != nil {
		log.Printf("Failed to get videos, error: %v\n", err)
		return nil, err
	}
	var filteredItems []*Video
	for _, item := range vidResults.Items {
		if item.Statistics.ViewCount != "" {
			views, err := strconv.Atoi(item.Statistics.ViewCount)
			if err != nil {
				log.Printf("Failed to convert view count to integer, error: %v\n", err)
				return nil, err
			}
			if views > MinViews {
				if snippetInfo, ok := vidIds[item.Id]; ok {
					item.Snippet.ChannelId = snippetInfo.ChannelId
					item.Snippet.ChannelTitle = snippetInfo.ChannelTitle
					item.Snippet.Thumbnails = snippetInfo.Thumbnails
				}
				filteredItems = append(filteredItems, (*Video)(item))
			}
		}
	}
	vidResults.Items = filteredItems

	// update videoCache with new results
	yt.Cache.SetVideo(input, vidResults)

	return vidResults, nil
}

// getChannelInfo hits the channel endpoint and returns the channel information
func getChannelInfo(channelId string) (*ChannelInfo, error) {
	pageUrl := fmt.Sprintf(GetChannelVideos, channelId, GetInstance().apiKey)

	resp, err := http.Get(pageUrl)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := ChannelInfo{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// getChannelPlaylist hits the playlist endpoint, returning playlist information
func (yt *YoutubeApi) getChannelPlaylist(playlistId string, numItems int) (*VideoResults, error) {
	numPages := calculateNumPages(numItems)

	videos, thumbnails, err := fetchPlaylistVideos(playlistId, numPages)
	if err != nil {
		return nil, err
	}

	getVideos, err := yt.GetVideos(videos)
	if err != nil {
		return nil, err
	}

	return processVideoItems(getVideos, thumbnails), nil
}

func calculateNumPages(numItems int) int {
	numPages := numItems / 50
	if numItems%50 > 0 {
		numPages += 1
	}
	return numPages
}

func fetchPlaylistVideos(playlistId string, numPages int) ([]string, map[string]Thumbnails, error) {
	var videos []string
	nextPage := ""
	thumbnails := make(map[string]Thumbnails)

	for i := 0; i < numPages; i++ {
		pageUrl := generatePageUrl(playlistId, nextPage, i)
		res, err := fetchVideoResultsFromAPI(pageUrl)
		if err != nil {
			return nil, nil, err
		}

		for _, vid := range res.Items {
			videos = append(videos, vid.ContentDetails.VideoId)
			thumbnails[vid.ContentDetails.VideoId] = vid.Snippet.Thumbnails
		}
		nextPage = res.NextPageToken
		if nextPage == "" {
			break
		}
	}
	return videos, thumbnails, nil
}

func generatePageUrl(playlistId, nextPage string, pageNum int) string {
	nextPageStr := ""
	if pageNum > 0 {
		nextPageStr = fmt.Sprintf("&pageToken=%v", nextPage)
	}
	return fmt.Sprintf(GetChannelPlaylist, playlistId, GetInstance().apiKey, nextPageStr)
}

func fetchVideoResultsFromAPI(url string) (*ChannelPlaylistVideoResults, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &ChannelPlaylistVideoResults{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func processVideoItems(videos *VideoResults, thumbnails map[string]Thumbnails) *VideoResults {
	for _, item := range videos.Items {
		if thumbs, ok := thumbnails[item.Id]; ok {
			item.Snippet.Thumbnails = thumbs
		}
	}
	return videos
}

// GetVideos hits the YouTube API to retrieve video information for the given input video IDs. It paginates through the input videos and aggregates the results into a single Video
func batchIteration(input []string) []string {
	var results []string
	for i := 0; i < len(input); i += 50 {
		end := i + 50
		if end > len(input) {
			end = len(input)
		}
		results = append(results, strings.Join(input[i:end], ","))
	}
	return results
}

func httpGetRequest(apiUrl string) ([]byte, error) {
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed HTTP request, error: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			if err != nil {
				log.Printf("error: %v\n", err)
			}
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading body, error: %w", err)
	}
	return body, nil
}

func unmarshalResponse(body []byte) (*VideoResults, error) {
	res := &VideoResults{}
	err := json.Unmarshal(body, res)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response body: %w", err)
	}
	return res, nil
}

func (yt *YoutubeApi) GetVideos(videoIds []string) (*VideoResults, error) {
	// Convert slice of videoIds to string to use as cache key
	videoIdsKey := strings.Join(videoIds, ",")

	if v := yt.Cache.GetVideoDetail(videoIdsKey); v != nil {
		return v, nil
	}

	input := batchIteration(videoIds)
	finalProduct := VideoResults{}
	pageVar := "&pageToken=%v"

	for _, fSearch := range input {
		nextPage := ""
		for i := 0; i < int(math.Ceil(float64(len(input))/float64(10))); i++ {
			nextPageStr := ""
			if i > 0 {
				nextPageStr = fmt.Sprintf(pageVar, nextPage)
			}
			apiUrl := fmt.Sprintf(GetTags, GetInstance().apiKey, fSearch, nextPageStr)
			body, err := httpGetRequest(apiUrl)
			if err != nil {
				return &finalProduct, err
			}

			res, err := unmarshalResponse(body)
			if err != nil {
				return &finalProduct, err
			}

			nextPage = res.NextPageToken
			if nextPage == "" {
				break
			}

			finalProduct.Items = append(finalProduct.Items, res.Items...)
		}
	}

	yt.Cache.SetVideoDetail(videoIdsKey, &finalProduct)

	return &finalProduct, nil
}

func (yt *YoutubeApi) SearchAndRetrieveTags(search string, pages ...int) (*VideoResults, error) {
	numPages := 1
	if pages != nil {
		if pages[0] > numPages {
			if pages[0] >= 5 {
				numPages = 5
			} else {
				numPages = pages[0]
			}
		}
	}
	return yt.FindTags(search, numPages)
}
