package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const cacheKey = "cache"

var (
	radarrHost = flag.String("radarr-host", "radarr", "Radarr host")
	radarrPort = flag.Int("radarr-port", 7878, "Radarr port")
	radarrKey  = flag.String("radarr-key", "", "Radarr key")
	serveHost  = flag.String("serve-host", "0.0.0.0", "Serve host")
	servePort  = flag.Int("serve-port", 7879, "Serve port")

	templates = template.Must(template.ParseFiles("main.gohtml"))
	printer   = message.NewPrinter(language.English)
)

func main() {

	flag.Parse()

	if *radarrKey == "" {
		fmt.Println("radarr-key is required")
		return
	}

	cache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 10, MaxCost: 10, BufferItems: 64})
	if err != nil {
		panic(err)
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {

		var b []byte

		value, found := cache.Get(cacheKey)
		if !found {

			//goland:noinspection HttpUrlsUsage
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/v3/movie?apikey=%s", *radarrHost, *radarrPort, *radarrKey))
			if err != nil {
				fmt.Println(err)
				return
			}

			//goland:noinspection GoUnhandledErrorResult
			defer resp.Body.Close()

			b, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			cache.SetWithTTL(cacheKey, string(b), 1, time.Hour)

		} else if val, ok := value.([]byte); ok {
			b = val
		} else {
			fmt.Println("cache value is not []byte")
			return
		}

		var movies []Movie
		err = json.Unmarshal(b, &movies)
		if err != nil {
			fmt.Println(err)
			return
		}

		slices.SortFunc(movies, func(a, b Movie) int {
			return cmp.Or(
				-cmp.Compare(a.DigitalRelease.Unix(), b.DigitalRelease.Unix()),
				-cmp.Compare(a.PhysicalRelease.Unix(), b.PhysicalRelease.Unix()),
				cmpBool(a.HasFile, b.HasFile),
				cmp.Compare(a.SortTitle, b.SortTitle),
			)
		})

		data := Data{Movies: movies}
		data.Shows = strings.Replace(fmt.Sprintf("//%s%s", r.Host, r.URL.String()), "movies", "shows", 1)

		err = templates.ExecuteTemplate(w, "main.gohtml", data)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	addr := fmt.Sprintf("%s:%d", *serveHost, *servePort)
	fmt.Println("Listening on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func cmpBool(a, b bool) int {
	switch {
	case a && !b:
		return 1
	case !a && b:
		return -1
	default:
		return 0
	}
}

type Data struct {
	Movies []Movie
	Shows  string
}

type Movie struct {
	Title            string `json:"title"`
	OriginalTitle    string `json:"originalTitle"`
	OriginalLanguage struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"originalLanguage"`
	AlternateTitles []struct {
		SourceType      string `json:"sourceType"`
		MovieMetadataId int    `json:"movieMetadataId"`
		Title           string `json:"title"`
		ID              int    `json:"id"`
	} `json:"alternateTitles"`
	SecondaryYearSourceId int       `json:"secondaryYearSourceId"`
	SortTitle             string    `json:"sortTitle"`
	SizeOnDisk            int64     `json:"sizeOnDisk"`
	Status                string    `json:"status"`
	Overview              string    `json:"overview"`
	InCinemas             time.Time `json:"inCinemas,omitempty"`
	DigitalRelease        time.Time `json:"digitalRelease,omitempty"`
	Images                []struct {
		CoverType string `json:"coverType"`
		Url       string `json:"url"`
		RemoteUrl string `json:"remoteUrl"`
	} `json:"images"`
	Website             string        `json:"website"`
	Year                int           `json:"year"`
	YouTubeTrailerID    string        `json:"youTubeTrailerId"`
	Studio              string        `json:"studio"`
	Path                string        `json:"path"`
	QualityProfileID    int           `json:"qualityProfileId"`
	HasFile             bool          `json:"hasFile"`
	MovieFileID         int           `json:"movieFileId"`
	Monitored           bool          `json:"monitored"`
	MinimumAvailability string        `json:"minimumAvailability"`
	IsAvailable         bool          `json:"isAvailable"`
	FolderName          string        `json:"folderName"`
	Runtime             int           `json:"runtime"`
	CleanTitle          string        `json:"cleanTitle"`
	IMDBID              string        `json:"imdbId,omitempty"`
	TMDBID              int           `json:"tmdbId"`
	TitleSlug           string        `json:"titleSlug"`
	RootFolderPath      string        `json:"rootFolderPath"`
	Certification       string        `json:"certification,omitempty"`
	Genres              []string      `json:"genres"`
	Tags                []interface{} `json:"tags"`
	Added               time.Time     `json:"added"`
	Ratings             struct {
		IMDB struct {
			Votes int     `json:"votes"`
			Value float64 `json:"value"`
			Type  string  `json:"type"`
		} `json:"imdb,omitempty"`
		TMDB struct {
			Votes int     `json:"votes"`
			Value float64 `json:"value"`
			Type  string  `json:"type"`
		} `json:"tmdb"`
		Metacritic struct {
			Votes int    `json:"votes"`
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"metacritic,omitempty"`
		RottenTomatoes struct {
			Votes int    `json:"votes"`
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"rottenTomatoes,omitempty"`
	} `json:"ratings"`
	MovieFile struct {
		MovieID      int       `json:"movieId"`
		RelativePath string    `json:"relativePath"`
		Path         string    `json:"path"`
		Size         int64     `json:"size"`
		DateAdded    time.Time `json:"dateAdded"`
		ReleaseGroup string    `json:"releaseGroup,omitempty"`
		Edition      string    `json:"edition"`
		Languages    []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"languages"`
		Quality struct {
			Quality struct {
				ID         int    `json:"id"`
				Name       string `json:"name"`
				Source     string `json:"source"`
				Resolution int    `json:"resolution"`
				Modifier   string `json:"modifier"`
			} `json:"quality"`
			Revision struct {
				Version  int  `json:"version"`
				Real     int  `json:"real"`
				IsRepack bool `json:"isRepack"`
			} `json:"revision"`
		} `json:"quality"`
		CustomFormatScore int `json:"customFormatScore"`
		IndexerFlags      int `json:"indexerFlags"`
		MediaInfo         struct {
			AudioBitrate          int     `json:"audioBitrate"`
			AudioChannels         float64 `json:"audioChannels"`
			AudioCodec            string  `json:"audioCodec"`
			AudioLanguages        string  `json:"audioLanguages"`
			AudioStreamCount      int     `json:"audioStreamCount"`
			VideoBitDepth         int     `json:"videoBitDepth"`
			VideoBitrate          int     `json:"videoBitrate"`
			VideoCodec            string  `json:"videoCodec"`
			VideoFps              float64 `json:"videoFps"`
			VideoDynamicRange     string  `json:"videoDynamicRange"`
			VideoDynamicRangeType string  `json:"videoDynamicRangeType"`
			Resolution            string  `json:"resolution"`
			RunTime               string  `json:"runTime"`
			ScanType              string  `json:"scanType"`
			Subtitles             string  `json:"subtitles"`
		} `json:"mediaInfo"`
		QualityCutoffNotMet bool   `json:"qualityCutoffNotMet"`
		ID                  int    `json:"id"`
		SceneName           string `json:"sceneName,omitempty"`
		OriginalFilePath    string `json:"originalFilePath,omitempty"`
	} `json:"movieFile,omitempty"`
	Collection struct {
		Title  string `json:"title,omitempty"`
		TmdbID int    `json:"tmdbId"`
	} `json:"collection,omitempty"`
	Popularity float64 `json:"popularity"`
	Statistics struct {
		MovieFileCount int      `json:"movieFileCount"`
		SizeOnDisk     int64    `json:"sizeOnDisk"`
		ReleaseGroups  []string `json:"releaseGroups"`
	} `json:"statistics"`
	ID              int       `json:"id"`
	PhysicalRelease time.Time `json:"physicalRelease,omitempty"`
	SecondaryYear   int       `json:"secondaryYear,omitempty"`
}

func (m Movie) Poster() string {
	for _, image := range m.Images {
		if image.CoverType == "poster" {
			q := url.Values{}
			q.Add("url", image.RemoteUrl)
			q.Add("output", "webp")
			q.Add("h", "400")
			q.Add("q", "100")
			return "https://images.weserv.nl/?" + q.Encode()
		}
	}
	return "https://critics.io/img/movies/poster-placeholder.png"
}

func (m Movie) Date() template.HTML {

	if m.DigitalRelease.Unix() > 0 {
		return template.HTML(m.DigitalRelease.Format("_2 Jan 2006"))
	}

	if m.PhysicalRelease.Unix() > 0 {
		return template.HTML(m.PhysicalRelease.Format("_2 Jan 2006"))
	}

	return "Unknown"
}

func (m Movie) IMDB() string {
	return strconv.FormatFloat(float64(m.Ratings.IMDB.Value)*10, 'f', 0, 64)
}

func (m Movie) Trending() string {
	return printer.Sprintf("%d", int(m.Popularity))
}
