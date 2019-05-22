# piko

Light and simple media downloader with support for:
- Youtube
- Soundcloud
- Imgur
- Facebook
- Twitter
- Instagram
- 4chan

# Installation

Grab the executable for your OS from [the releases page of this repository](https://github.com/mlvzk/piko/releases) and put it in your $PATH (on Linux typically `/usr/bin/`)

Or from source:
```sh
go get -u github.com/mlvzk/piko/...
```

# Usage

```sh
piko --help
piko [urls...]
```

# Examples

```sh
piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

```sh
piko --option quality=best --format "%[title].%[ext]" 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

```sh
# you can discover options and meta information for formats with --discover flag
piko --discover 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

```sh
# output to stdout, open mpv to read from stdin
piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --stdout | mpv -
```