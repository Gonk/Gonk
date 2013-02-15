package modules

import (
	"log"
	"regexp"
	"strings"

	"github.com/NickPresta/GoURLShortener"
	irc "github.com/fluffle/goirc/client"
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
	shortenEmbeds := l.AlwaysShortenEmbeds || strings.HasPrefix(line, l.Client.Me.Nick+" link ")

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
	replaces, newText := ShortenUrls(line, l.AlwaysShortenEmbeds, true, l.MaxUrlLength)
	if replaces > 0 {
		responded = true

		l.Client.Privmsg(target, newText)
	}

	return
}

// ShortenUrls shortens URLs in the given text. By default, it only shortens
// URLs if they are longer than the specified maxLength and not embeddable.
// Supplying a true value to shortenEmbeds or shortenImages will change that
// behavior.
// Returns a count of shortened URLs and the (potentially) modified text.
func ShortenUrls(text string, shortenEmbeds bool, shortenImages bool, maxLength int) (int, string) {
	var replacements []string

	matches := urlRegex.FindAllString(text, -1)
	for _, match := range matches {
		// Determine whether to shorten URL
		if len(match) > maxLength && (shortenEmbeds || !IsEmbeddable(match) || (IsImage(match) && shortenImages)) {
			uri, err := goisgd.Shorten(match)
			if err != nil {
				log.Println("LinkShortener Error:", match, err)
				continue
			}

			// Enable shortened image URLs to be embedded by capable clients
			if IsImage(match) {
				uri += "#.png"
			}

			replacements = append(replacements, match, uri)
		}
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
