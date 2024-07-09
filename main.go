package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {

	fmt.Println("Hello World")

	var templates = template.Must(template.ParseFiles("main.gohtml"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Hello World")

		resp, err := http.Get("http://radarr:7878/api/v3/movie?apikey=" + os.Getenv("RADARR_API_KEY"))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var movies []Movie
		err = json.Unmarshal(b, &movies)
		if err != nil {
			fmt.Println(err)
			return
		}

		movies = filter(movies, func(movie Movie) bool {
			return movie.DigitalRelease.Year() > 2000
		})

		sort.Slice(movies, func(i, j int) bool {
			return movies[i].DigitalRelease.Unix() > movies[j].DigitalRelease.Unix()
		})

		err = templates.ExecuteTemplate(w, "main.gohtml", Data{Movies: movies})
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	err := http.ListenAndServe("0.0.0.0:7879", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

type Data struct {
	Movies []Movie
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
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
	return ""
}

func (m Movie) Date() string {
	return m.DigitalRelease.Format("_2 Jan 2006")
}

func (m Movie) IMDB() string {
	return strconv.FormatFloat(float64(m.Ratings.IMDB.Value)*10, 'f', 0, 64)
}
