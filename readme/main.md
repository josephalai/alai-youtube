# YouTube API Wrapper

This Go package is a sophisticated interface for the YouTube API, designed to streamline the process of interacting with YouTube's data. It includes functionalities such as video ID searching, tag retrieval, channel video fetching, and playlist obtaining, all while emphasizing performance through a robust caching mechanism.

## Documentation Structure

- [Getting Started](./main.md) - Start here if you're new to the project.
- [Caching Implementation Guide](./cache.md) - Learn how to enhance performance with our caching interface.
- [Gin-Gonic Web Server Setup](./gin-gonic.md) - Instructions for setting up the YouTube API wrapper with Gin-Gonic.


## Features

- **Singleton Service Instance**: Guarantees a single instance of the YouTube API service, providing a unified access point and enhancing efficiency.
- **API Key Configuration**: Facilitates the dynamic setup of the YouTube API key, essential for authenticating requests and ensuring secure access.
- **Caching Mechanism**: Implements a caching strategy to temporarily store channel information, video details, and playlists. This approach significantly reduces API requests, lowers latency, and improves overall performance.
- **Error Handling**: Advanced error management techniques are employed to identify, log, and address issues encountered during API interactions.
- **Comprehensive Data Retrieval**: Offers the ability to obtain detailed statistics and information about videos and channels, such as view counts, subscriber numbers, and video metadata.

## Getting Started

Ensure you have Go installed on your machine. Incorporate this package into your project by importing it as follows:

```go
import "github.com/josephalai/alai-youtube"
```

### Initialization

Instantiate the `YoutubeApi` service with the `GetInstance` function. This step allows for the optional inclusion of an API key and a custom cache:

```go
apiInstance := services.GetInstance(map[string]interface{}{
    "apiKey": "YOUR_API_KEY",
    // Optionally, specify a custom cache
    "cache": YourCustomCacheInstance,
})
```

### Usage Examples

**Obtaining Channel Information:**

```go
channelInfo, err := apiInstance.GetChannelInfo("CHANNEL_ID")
if err != nil {
    log.Printf("Error fetching channel info: %v\n", err)
} else {
    fmt.Printf("Channel Info: %+v\n", channelInfo)
}
```

**Tag Searching and Retrieval:**

```go
tags, err := apiInstance.SearchAndRetrieveTags("SEARCH_QUERY")
if err != nil {
    log.Printf("Error retrieving tags: %v\n", err)
} else {
    fmt.Printf("Tags: %v\n", tags)
}
```

### Advanced Features and Integration

- **Caching Mechanism Integration**: For details on how to implement and integrate the caching mechanism to enhance performance, see [cache.md](readme/cache.md).
- **Gin-Gonic Server Implementation**: To learn how to use gin-gonic for setting up a server to utilize the YouTube wrapper, refer to [gin-gonic.md](readme/gin-gonic.md).

### Contributing

We encourage contributions to enhance this package. Adhere to standard Go project guidelines for pull requests.

### License

This project is under the MIT License. Refer to the LICENSE file for more details.

---
