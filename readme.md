# piko

Light and simple media downloader with support for:
- Youtube - single /watch?v= videos and playlists(only 100 first videos)
- Soundcloud - single songs
- Imgur - albums
- Facebook - single and multiple images/videos in one post
- Twitter - \*/status/\* links, single and multiple images/videos of single posts
- Instagram - single and multiple images of single posts
- 4chan - all images and videos of a thread and it's posts

TODO:
- Twitch - recording livestreams and watching with `-stdout | mpv -`
- Youtube - support more than 100 videos in playlists(might need API key which has quota limit)
- Soundcloud - support playlists
- Facebook - support downloading all images/videos posted by a page
- Twitter - support downloading all images/videos posted by an account
- Instagram - support downloading videos and downloading all images/videos of an account
- 4chan - support downloading all images of alive threads in a board

# Installation

Grab the executable for your OS from [the releases page of this repository](https://github.com/mlvzk/piko/releases) and put it in your $PATH (on Linux typically `/usr/bin/`)

Or from source:
```sh
go get -u github.com/mlvzk/piko/...
```

## Optional dependencies

- ffmpeg (youtube, for better video and audio quality)

# Usage

```sh
piko --help
piko [urls...]
```

# Examples

```sh
# downloads the video(with audio) to a file with default name format (see --discover example below)
# if ffmpeg is not installed, the quality might be bad
piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

```sh
# tell youtube service to choose best quality
# save to file with name format %[title].%[ext] (see --discover example below)
piko --option quality=best --format "%[title].%[ext]" 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

```sh
# you can discover options and meta information for formats with --discover flag
piko --discover 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'

# output:
Default Name: %[title].%[ext]
Meta:
        ext=mkv
        title=Rick Astley - Never Gonna Give You Up (Video)
        author=RickAstleyVEVO
Available Options:
        onlyAudio:
                - yes
                - no
        quality:
                - best
                - medium
                - worst
        useFfmpeg:
                - yes
                - no
Default Options:
        useFfmpeg=yes
        onlyAudio=no
        quality=medium
```

```sh
# output to stdout, pipe to mpv which reads from stdin
piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --stdout | mpv -
```

```sh
# download only audio, with best quality, output to stdout, pipe to mpv which reads from stdin
piko --option onlyAudio=yes --option quality=best 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --stdout | mpv -
```

# Contributors

- [mlvzk](https://github.com/mlvzk) - creator and maintainer
- [gwu](https://github.com/gwimm) - the idea came from gwu's [shovel](http://gitlab.com/gwu/shovel) project
