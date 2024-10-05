# ripper-api-downloader-go
Downloader for my ripper api written in go

## Dependencies:
* flac
* ffmpeg

## Usage:
`downloader link1 link2 ...`
or just `downloader`

## Config example (TOML format):
```toml
path = "/home/user/Downloads/"       # folder files will be downloaded to
url = "https://server.url/"          # api url
apikey = "test"                      # api key
convert = true                       # set to true if you want to convert from alac to flac (needs unarchive)
unarchive = true                     # set to true if you want to unarchive downloaded zip
```
