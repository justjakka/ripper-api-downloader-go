package downloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/schollz/progressbar/v3"
)

func GetToken() (string, error) {
	req, err := http.NewRequest("GET", "https://beta.music.apple.com", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error while closing: %v\n", err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`/assets/index-legacy-[^/]+\.js`)
	indexJsUri := regex.FindString(string(body))

	req, err = http.NewRequest("GET", "https://beta.music.apple.com"+indexJsUri, nil)
	if err != nil {
		return "", err
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error while closing: %v\n", err.Error())
		}
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	regex = regexp.MustCompile(`eyJh([^"]*)`)
	token := regex.FindString(string(body))

	return token, nil
}

func GetMeta(albumId, token, storefront string) (*AutoGenerated, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://amp-api.music.apple.com/v1/catalog/%s/albums/%s", storefront, albumId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Origin", "https://music.apple.com")
	query := url.Values{}
	query.Set("omit[resource]", "autos")
	query.Set("include", "tracks,artists,record-labels")
	query.Set("include[songs]", "artists")
	query.Set("fields[artists]", "name")
	query.Set("fields[albums:albums]", "artistName,artwork,name,releaseDate,url")
	query.Set("fields[record-labels]", "name")
	// query.Set("l", "en-gb")
	req.URL.RawQuery = query.Encode()
	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error while closing: %v\n", err.Error())
		}
	}(do.Body)
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	obj := new(AutoGenerated)
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func CheckUrl(url string) (string, string) {
	pat := regexp.MustCompile(`^https://(?:beta\.music|music)\.apple\.com/(\w{2})(?:/album|/album/.+)/(?:id)?(\d+)(?:$|\?)`)
	matches := pat.FindAllStringSubmatch(url, -1)
	if matches == nil {
		return "", ""
	} else {
		return matches[0][1], matches[0][2]
	}
}

func GetReleaseInfo(url string) (string, error) {
	token, err := GetToken()
	if err != nil {
		return "", err
	}

	storefront, albumId := CheckUrl(url)

	meta, err := GetMeta(albumId, token, storefront)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s - %s", meta.Data[0].Attributes.ArtistName, meta.Data[0].Attributes.Name), nil
}

func SubmitAlbum(config *Config, url string, client *http.Client) (*JobQuery, error) {
	reqBody, err := json.Marshal(SubmittedUrl{Url: url})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", config.Url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("api-key", config.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error while closing: %v\n", err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 202 {
		var msg Message
		err = json.Unmarshal(body, &msg)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(msg.Msg)
	}

	var Jobinfo JobQuery
	err = json.Unmarshal(body, &Jobinfo)
	if err != nil {
		return nil, err
	}

	return &Jobinfo, nil
}

func CheckResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case 200, 201, 204:
		return nil
	}

	var msg Message
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		return err
	}
	return fmt.Errorf(msg.Msg)
}

func QueryJob(config *Config, job *JobQuery, client *http.Client, releaseinfo string) (*http.Response, error) {
	reqBody, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%sjob/", config.Url), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("api-key", config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	time.Sleep(2 * time.Second)

	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetDescription("[white]Waiting in queue[reset]"),
	)

	bar.Reset()

	resp, err := client.Do(req)
	if err != nil {
		err := bar.Finish()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	for resp.StatusCode == 201 {
		time.Sleep(3 * time.Second)

		resp, err = client.Do(req)
		if err != nil {
			err := bar.Finish()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
		err = CheckResponse(resp)
		if err != nil {
			return nil, err
		}
	}
	err = bar.Finish()
	if err != nil {
		return nil, err
	}

	bar = progressbar.NewOptions(
		-1,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetDescription(fmt.Sprintf("[yellow]Processing %s[reset]", releaseinfo)),
	)
	bar.Reset()

	for resp.StatusCode == 204 {
		time.Sleep(3 * time.Second)

		resp, err = client.Do(req)
		if err != nil {
			err := bar.Finish()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
		err = CheckResponse(resp)
		if err != nil {
			return nil, err
		}
	}

	err = bar.Finish()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		return resp, nil
	} else {
		var msg Message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &msg)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("invalid response %s", msg.Msg)
	}
}

func Download(config *Config, url string, client *http.Client, counter, maxDownloads int) (string, error) {
	releasetitle, err := GetReleaseInfo(url)
	if err != nil {
		return "", err
	}

	jobinfo, err := SubmitAlbum(config, url, client)
	if err != nil {
		return "", err
	}

	resp, err := QueryJob(config, jobinfo, client, releasetitle)

	if err != nil {
		return "", err
	}

	sanAlbumFolder := filepath.Join(config.Path, ForbiddenNames.ReplaceAllString(releasetitle, "_"))
	sanZipName := fmt.Sprintf("%s.zip", sanAlbumFolder)

	f, err := os.OpenFile(sanZipName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error while closing: %v\n", err.Error())
		}
	}(resp.Body)
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(fmt.Sprintf("[green]%d/%d[reset] %s", counter, maxDownloads, releasetitle)),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return "", err
	}
	err = bar.Finish()
	if err != nil {
		return "", err
	}
	fmt.Println()
	return sanZipName, nil
}
