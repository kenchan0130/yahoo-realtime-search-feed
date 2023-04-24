# Yahoo! realtime search feed

[![Build status][github-actions-image]][github-actions-url]

[github-actions-image]: https://github.com/kenchan0130/yahoo-realtime-search-feed/workflows/CI/badge.svg
[github-actions-url]: https://github.com/kenchan0130/yahoo-realtime-search-feed/actions?query=workflow%3A%22CI%22

A web application that provides entry of Yahoo! Realtime Search as RSS feed.

Sample is [here](https://yahoo-realtime-search-feed.onrender.com/).

## Endpoints

| endpoint  | content                |
|-----------|------------------------|
| `/`       | redirect to `/health`. |
| `/health` | return "ok" as text.   |
| `/feed`   | return a RSS feed.     |

### `/feed`

#### GET `/feed`

This endpoint supports the following query parameters.

| parameter | description                 |
|-----------|-----------------------------|
| `q`       | search query.               |
| `limit`   | number of items to display. |

## Development

```sh
go get -u github.com/kenchan0130/yahoo-realtime-search-feed
```

You may also clone this project instead.
And, please run the program.

```sh
go run main.go
```

## Deploy

Any changes to the main branch are automatically deployed to [render](https://render.com/) by GitHub Action.
