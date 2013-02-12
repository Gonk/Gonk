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
	AlwaysShortenImages bool
	MaxUrlLength        int
}

func (l LinkShortener) Respond(target string, line string) {
	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, l.AlwaysShortenImages, l.MaxUrlLength)
	if replaces > 0 {
		l.Client.Privmsg(target, newText)
	}
}

func (l LinkShortener) Hear(target string, line string) {
	shortenImages := l.AlwaysShortenImages || strings.HasSuffix(line, l.Client.Me.Nick+" link ")

	// Replace URLs and send result
	replaces, newText := ShortenUrls(line, shortenImages, l.MaxUrlLength)
	if replaces > 0 {
		l.Client.Privmsg(target, newText)
	}
}

// ShortenUrls shortens URLs in the given text. It only shortens URLs if they
// are longer than the specified maxLength and not images (unless shortenImages
// is true).
func ShortenUrls(text string, shortenImages bool, maxLength int) (int, string) {
	var replacements []string

	matches := urlRegex.FindAllString(text, -1)
	for _, match := range matches {
		// Determine whether to shorten URL
		if len(match) > maxLength && (shortenImages || !isImage(match)) {
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

func isImage(url string) bool {
	return strings.HasSuffix(url, "jpg") ||
		strings.HasSuffix(url, "jpeg") ||
		strings.HasSuffix(url, "png") ||
		strings.HasSuffix(url, "gif")
}
