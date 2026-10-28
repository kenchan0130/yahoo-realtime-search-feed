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
| `/feed`   | return a RSS or JSON feed. |

### `/feed`

#### GET `/feed`

This endpoint supports the following query parameters.

| parameter | description                 |
|-----------|-----------------------------|
| `q`       | search query.               |
| `limit`   | number of items to display. |
| `nocache` | fetch the latest results without the cache when set to `1`, `true` or no value (e.g. `/feed?q=foo&nocache`). |
| `format`  | response format: `rss` (default) or `json`. `json` returns a [JSON Feed](https://www.jsonfeed.org/version/1.1/) as `application/feed+json` (e.g. `/feed?q=foo&format=json`). Other values return status `400`. |

#### Search query syntax

`q` is passed to Yahoo! Realtime Search as it is, so its search operators are available.
The value of `q` must be URL-encoded (UTF-8), including spaces, `+` and multibyte characters, and sent as `/feed?q=<URL-encoded value>` (e.g. `/feed?q=golang+-rust`).

| pattern      | meaning                                                         | example                   | URL-encoded value             |
|--------------|-----------------------------------------------------------------|---------------------------|-------------------------------|
| `A`          | entries containing `A`.                                         | `golang`                  | `golang`                      |
| `A B`        | AND: entries containing both `A` and `B`.                       | `golang rust`             | `golang+rust`                 |
| `A OR B`     | OR: entries containing `A` or `B`.                              | `golang OR rust`          | `golang+OR+rust`              |
| `A -B`       | NOT: entries containing `A`, excluding those containing `B`.    | `golang -rust`            | `golang+-rust`                |
| `A -B -C`    | excludes entries containing **either** `B` or `C`.              | `golang -rust -python`    | `golang+-rust+-python`        |
| `A (B OR C)` | grouping: `A` AND (`B` OR `C`).                                 | `golang (rust OR python)` | `golang+%28rust+OR+python%29` |
| `A B OR C`   | same as `A (B OR C)`, because OR binds tighter than AND.        | `golang rust OR python`   | `golang+rust+OR+python`       |
| `A OR B -B`  | exclusions apply to the whole query, so no entries contain `B`. | `golang OR rust -rust`    | `golang+OR+rust+-rust`        |
| multibyte    | the operators above also work with Japanese and other terms.    | `猫 -犬`                  | `%E7%8C%AB+-%E7%8A%AC`        |

Notes:

- Matching is case-insensitive (`-rust` also excludes `Rust`).
- Use the half-width `-` and `OR`. The full-width minus `－` is not an exclusion operator and is searched as a term.
- Exclusions are applied to the post body. A reply's leading `@username` may still contain an excluded term.
- These operators are provided by Yahoo! Realtime Search, and their behavior may change on Yahoo's side.

#### Query length limit

Yahoo! Realtime Search rejects a query that is too long, and then `/feed` returns status `400` with the reason as text.
The limit is not a simple number of characters, because Yahoo splits terms internally, and multibyte characters reach it much sooner.
As a rough guide (measured, not officially documented):

| characters | approximate limit                   |
|------------|-------------------------------------|
| ASCII      | about 3,500 characters (~450 terms) |
| multibyte  | about 600 characters (~90 terms)    |

Search results are cached per `q` for 5 minutes by default.
The duration can be changed with the `CACHE_TTL` environment variable (e.g. `CACHE_TTL=1m`), and `CACHE_TTL=0` disables the cache.
A request with `nocache` also updates the cache with the latest results.

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
