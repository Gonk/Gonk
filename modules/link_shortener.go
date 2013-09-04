package modules

import (
	log "github.com/fluffle/golog/logging"
	"regexp"
	"strings"

	irc "github.com/Gonk/goirc/client"
	"github.com/NickPresta/GoURLShortener"
)

var urlRegex = regexp.MustCompile(`(http|https|ftp|ftps)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,4}(/\S*)?`)

// LinkShortener shortens any URLs in messages it hears and echos them back to
// the source.
type LinkShortener struct {
	Client              *irc.Conn
	AlwaysShortenEmbeds bool
	MaxUrlLength        int
}

func (l LinkShortener) Respond(target string, line string, from string) (responded bool) {
	shortenEmbeds := l.AlwaysShortenEmbeds || strings.HasPrefix(line, l.Client.Me().Nick+" link ")

	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, shortenEmbeds, true, l.MaxUrlLength)
	if replaces > 0 {
		responded = true

		l.Client.Privmsg(target, newText)
	}

	return
}

func (l LinkShortener) Hear(target string, line string, from string) (responded bool) {
	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, l.AlwaysShortenEmbeds, false, l.MaxUrlLength)
	if replaces > 0 {
		responded = true

		l.Client.Privmsg(target, newText)
	}

	return
}

// ShortenUrls shortens URLs in the given text. By default, it only shortens
// URLs if they are longer than the specified maxLength and not embeddable.
// Supplying a true value to shortenEmbeds or shortenImages will change that
// behavior. In addition, if shortenImages is true, the shortened URL will
// have '#.png' appended, in order to enable clients to embed the image.
// Returns a count of shortened URLs and the (potentially) modified text.
func ShortenUrls(text string, shortenEmbeds bool, shortenImages bool, maxLength int) (int, string) {
	var replacements []string

	work := 0
	workDone := make(chan bool)

	matches := urlRegex.FindAllString(text, -1)
	for _, match := range matches {
		work++

		go func(match string) {
			// Determine whether to shorten URL
			if len(match) > maxLength && (shortenEmbeds || !IsEmbeddable(match) || (IsImage(match) && shortenImages)) {
				uri, err := goisgd.Shorten(match)
				if err != nil {
					log.Error("LinkShortener Error:", match, err)

					workDone <- true
					return
				}

				// Enable shortened image URLs to be embedded by capable clients
				if shortenImages && IsImage(match) {
					uri += "#.png"
				}

				replacements = append(replacements, match, uri)
			}

			workDone <- true
		}(match)
	}

	for work > 0 {
		work--
		<-workDone
	}

	r := strings.NewReplacer(replacements...)

	return len(replacements) / 2, r.Replace(text)
}

// Returns true if the URL points to an image or other embeddable object
func IsEmbeddable(url string) bool {
	return IsImage(url) ||
		strings.Contains(url, "youtube.com")
}

// Returns true if the URL points to an image
func IsImage(url string) bool {
	return strings.HasSuffix(url, "jpg") ||
		strings.HasSuffix(url, "jpeg") ||
		strings.HasSuffix(url, "png") ||
		strings.HasSuffix(url, "gif")
}
