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

func (l LinkShortener) Respond(target string, line string) {
	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, l.AlwaysShortenEmbeds, l.MaxUrlLength)
	if replaces > 0 {
		l.Client.Privmsg(target, newText)
	}
}

func (l LinkShortener) Hear(target string, line string) {
	shortenEmbeds := l.AlwaysShortenEmbeds || strings.HasSuffix(line, l.Client.Me.Nick+" link ")

	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, shortenEmbeds, l.MaxUrlLength)
	if replaces > 0 {
		l.Client.Privmsg(target, newText)
	}
}

// ShortenUrls shortens URLs in the given text. It only shortens URLs if they
// are longer than the specified maxLength and not embeddable (unless
// shortenEmbeds is true).
func ShortenUrls(text string, shortenEmbeds bool, maxLength int) (int, string) {
	var replacements []string

	matches := urlRegex.FindAllString(text, -1)
	for _, match := range matches {
		// Determine whether to shorten URL
		if len(match) > maxLength && (shortenEmbeds || !isEmbeddable(match)) {
			uri, err := goisgd.Shorten(match)
			if err != nil {
				log.Println("LinkShortener Error:", match, err)
				continue
			}

			replacements = append(replacements, match, uri)
		}
	}

	r := strings.NewReplacer(replacements...)

	return len(replacements) / 2, r.Replace(text)
}

func isEmbeddable(url string) bool {
	return strings.HasSuffix(url, "jpg") ||
		strings.HasSuffix(url, "jpeg") ||
		strings.HasSuffix(url, "png") ||
		strings.HasSuffix(url, "gif") ||
		strings.Contains(url, "youtube.com")
}
