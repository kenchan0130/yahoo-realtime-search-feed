package models

import (
	"fmt"
	"strconv"
	"time"
)

type unixTimestamp struct {
	time.Time
}

func (t *unixTimestamp) UnmarshalJSON(data []byte) error {
	v, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("strconv.Atoi(): %v", v)
	}
	t.Time = time.Unix(int64(v), 0)

	return nil
}

type PageData struct {
	Timeline *Timeline `json:"timeline,omitempty"`
}

type PageProps struct {
	PageData *PageData `json:"pageData,omitempty"`
}

type Props struct {
	PageProps *PageProps `json:"pageProps,omitempty"`
}

type Timeline struct {
	Entry *[]TimelineEntry `json:"entry,omitempty"`
}

type TimelineEntry struct {
	ID                   *string                 `json:"id,omitempty"`
	EncryptedId          *string                 `json:"encryptedId,omitempty"`
	URL                  *string                 `json:"url,omitempty"`
	Verified             *bool                   `json:"verified,omitempty"`
	DisplayText          *string                 `json:"displayText,omitempty"`
	DisplayTextBody      *string                 `json:"displayTextBody,omitempty"`
	DisplayTextFragments *string                 `json:"displayTextFragments,omitempty"`
	DisplayTextEntities  *string                 `json:"displayTextEntities,omitempty"`
	URLs                 *[]TimelineEntryURL     `json:"urls,omitempty"`
	Hashtags             *[]TimelineEntryHashtag `json:"hashtags,omitempty"`
	Mentions             *[]TimelineEntryMention `json:"mentions,omitempty"`
	ReplyMentions        *[]string               `json:"replyMentions,omitempty"`
	CreatedAt            *unixTimestamp          `json:"createdAt,omitempty"`
	ReplyCount           *int                    `json:"replyCount,omitempty"`
	ReplyURL             *string                 `json:"replyUrl,omitempty"`
	RTCount              *int                    `json:"rtCount,omitempty"`
	RTURL                *string                 `json:"rturl,omitempty"`
	LikesCount           *int                    `json:"likesCount,omitempty"`
	LikesURL             *string                 `json:"likesUrl,omitempty"`
	QuoteCount           *int                    `json:"quoteCount,omitempty"`
	UserID               *string                 `json:"userId,omitempty"`
	UserURL              *string                 `json:"userUrl,omitempty"`
	Name                 *string                 `json:"name,omitempty"`
	ScreenName           *string                 `json:"screenName,omitempty"`
	ProfileImage         *string                 `json:"profileImage,omitempty"`
	ProfileImageBig      *string                 `json:"profileImageBig,omitempty"`
	MediaType            *[]string               `json:"mediaType,omitempty"`
	Images               *[]TimelineEntryImage   `json:"images,omitempty"`
	TweetThemeNormal     *[]string               `json:"tweetThemeNormal,omitempty"`
	UserThemeNormal      *[]string               `json:"userThemeNormal,omitempty"`
	PossiblySensitive    *bool                   `json:"possiblySensitive,omitempty"`
	Relevance            *int                    `json:"relevance,omitempty"`
	Media                *[]TimelineEntryMedia   `json:"media,omitempty"`
}

type TimelineEntryHashtag struct {
	Text    *string `json:"text,omitempty"`
	Indices *[]int  `json:"indices,omitempty"`
}

type TimelineEntryImage struct {
	RawURL     *string                 `json:"rawUrl,omitempty"`
	URL        *string                 `json:"url,omitempty"`
	ResizedURL *string                 `json:"resizedUrl,omitempty"`
	DisplayURL *string                 `json:"displayUrl,omitempty"`
	Size       *TimelineEntryImageSize `json:"size,omitempty"`
}

type TimelineEntryImageSize struct {
	Small  *TimelineEntryImageSizeObject `json:"small,omitempty"`
	Medium *TimelineEntryImageSizeObject `json:"medium,omitempty"`
	Large  *TimelineEntryImageSizeObject `json:"large,omitempty"`
	Thumb  *TimelineEntryImageSizeObject `json:"thumb,omitempty"`
}

type TimelineEntryImageSizeObject struct {
	Resize *string `json:"resize,omitempty"`
	W      *int    `json:"w,omitempty"`
	H      *int    `json:"h,omitempty"`
}

type TimelineEntryMedia struct {
	Image        *TimelineEntryMediaObject `json:"image,omitempty"`
	Video        *TimelineEntryMediaObject `json:"video,omitempty"`
	AnimatedGif  *TimelineEntryMediaObject `json:"animatedGif,omitempty"`
	YouTube      *TimelineEntryMediaObject `json:"youTube,omitempty"`
	MetaImageURL *string                   `json:"metaImageUrl,omitempty"`
}

type TimelineEntryMediaObject struct {
	URL               *string                       `json:"url,omitempty"`
	DisplayURL        *string                       `json:"displayUrl,omitempty"`
	MediaURL          *string                       `json:"mediaUrl,omitempty"`
	Sizes             *TimelineEntryMediaImageSizes `json:"sizes,omitempty"`
	ThumbnailImageURL *string                       `json:"thumbnailImageUrl,omitempty"`
}

type TimelineEntryMediaImageSizes struct {
	Viewer *TimelineEntryMediaImageSizesViewer `json:"viewer,omitempty"`
}

type TimelineEntryMediaImageSizesViewer struct {
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
}

type TimelineEntryMention struct {
	ID         *string `json:"id,omitempty"`
	Name       *string `json:"name,omitempty"`
	ScreenName *string `json:"screenName,omitempty"`
	Indices    *[]int  `json:"indices,omitempty"`
}

type TimelineEntryURL struct {
	DisplayURL  *string `json:"displayUrl,omitempty"`
	ExpandedURL *string `json:"expandedUrl,omitempty"`
	URL         *string `json:"url,omitempty"`
	Indices     *[]int  `json:"indices,omitempty"`
}
