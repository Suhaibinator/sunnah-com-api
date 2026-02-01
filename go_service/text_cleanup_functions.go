package go_service

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	reHtmlParagraphTag = regexp.MustCompile(`^<p>|</p>$`)
	reCqTag            = regexp.MustCompile(`</?c_q\d+>`)

	reAllahsMessenger  = regexp.MustCompile(`Allah's Messenger `)
	reMessengerOfAllah = regexp.MustCompile(`he Messenger of Allah `)
	reProphet          = regexp.MustCompile(`he Prophet `)

	reOpenQuran = regexp.MustCompile(`javascript:openquran$begin:math:text$(.+?)$end:math:text$`)
	reNewline   = regexp.MustCompile(`\n+`)
	reSpace     = regexp.MustCompile(` +`)
)

func negativeLookBackReplacement(re *regexp.Regexp, text string, replacement string) string {
	matches := re.FindAllStringIndex(text, -1)
	if matches == nil {
		return text
	}

	var result strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]
		// Write everything before this match
		result.WriteString(text[lastIndex:start])

		// If the next char is '(' => do NOT replace
		if end < len(text) && text[end] == '(' {
			// Keep original
			result.WriteString(text[start:end])
		} else {
			// Replace
			result.WriteString(replacement)
		}
		lastIndex = end
	}

	// Append the rest
	result.WriteString(text[lastIndex:])
	return result.String()
}

// outerHTML renders a goquery Selection’s exact HTML (including the element itself).
func outerHTML(s *goquery.Selection) (string, error) {
	var buf bytes.Buffer
	for _, node := range s.Nodes {
		if err := html.Render(&buf, node); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// fixHTML is revised to collect the entire <body> HTML, rather than skipping
// or splitting by child elements. This avoids losing big chunks of text
// that got split into separate text nodes or partial elements.
func fixHTML(text string, removeWrapper bool) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r", "")

	// Always wrap in a top-level <html><body> so net/html parser doesn’t discard anything
	fullHTML := "<html><body>" + text + "</body></html>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fullHTML))
	if err != nil {
		return text // fallback
	}

	// Remove 'id' and 'name' attrs from all <a>
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("id")
		s.RemoveAttr("name")
	})

	// Get the entire body’s HTML content
	bodyHTML, err := doc.Find("body").Html()
	if err != nil {
		return text
	}
	result := bodyHTML

	// If asked, remove leading/trailing <p>...</p>
	if removeWrapper {
		result = reHtmlParagraphTag.ReplaceAllString(result, "")
	}

	// Remove <c_q...> tags
	result = reCqTag.ReplaceAllString(result, "")

	return result
}

// standardizeTerms matches what Python does with “PBUH”, “(saw)”, etc.
func standardizeTerms(text string) string {
	pairs := []struct {
		old string
		new string
	}{
		{"PBUH", "\ufdfa"},
		{"P.B.U.H.", "\ufdfa"},
		{"peace_be_upon_him", "\ufdfa"},
		{"(may peace be upon him)", "(\ufdfa)"},
		{"(saws)", "(\ufdfa)"},
		{"(SAW)", "(\ufdfa)"},
		{"(saw)", "(\ufdfa)"},
		{"he Apostle of Allah", "he Messenger of Allah"},
		{"he Apostle of Allaah", "he Messenger of Allah"},
		{"Allah's Apostle", "Allah's Messenger"},
		{"he Holy Prophet ", "he Prophet "},
	}
	for _, p := range pairs {
		text = strings.ReplaceAll(text, p.old, p.new)
	}

	// Then negativeLookBack for “Allah's Messenger ” => “Allah's Messenger (\ufdfa) ”
	text = negativeLookBackReplacement(reAllahsMessenger, text, "Allah's Messenger (\ufdfa) ")
	text = negativeLookBackReplacement(reMessengerOfAllah, text, "he Messenger of Allah (\ufdfa) ")
	text = negativeLookBackReplacement(reProphet, text, "he Prophet (\ufdfa) ")

	return text
}

// fixHyperlinks is unchanged, except that net/html parse might reorder newlines
func fixHyperlinks(text string) string {
	// e.g. href="/..." => href="https://sunnah.com/..."
	text = strings.ReplaceAll(text, `href="/`, `href="https://sunnah.com/`)

	matches := reOpenQuran.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return text
	}

	var result strings.Builder
	lastIndex := 0
	for _, match := range matches {
		start := match[0]
		end := match[1]
		group1start := match[2]
		group1end := match[3]

		// Everything before
		result.WriteString(text[lastIndex:start])

		linkMatch := text[group1start:group1end]
		parts := strings.Split(linkMatch, ",")
		if len(parts) != 3 {
			log.Printf("Invalid link match: %s\n", linkMatch)
			// keep as-is
			result.WriteString(text[start:end])
			lastIndex = end
			continue
		}
		surahStr := strings.TrimSpace(parts[0])
		begin := strings.TrimSpace(parts[1])
		endVerse := strings.TrimSpace(parts[2])

		surahInt, err := strconv.Atoi(surahStr)
		if err != nil {
			log.Printf("Invalid surah number: %s\n", surahStr)
			result.WriteString(text[start:end])
			lastIndex = end
			continue
		}
		surahInt++
		newURL := fmt.Sprintf("https://quran.com/%d/%s-%s", surahInt, begin, endVerse)
		result.WriteString(newURL)
		lastIndex = end
	}
	result.WriteString(text[lastIndex:])
	return result.String()
}

// ----------------------------------------------------------------------------
// Public cleanup functions
// ----------------------------------------------------------------------------

func CleanupText(text string) string {
	if text == "" {
		return text
	}
	// compress newlines/spaces
	text = reNewline.ReplaceAllString(text, "\n")
	text = reSpace.ReplaceAllString(text, " ")

	text = fixHTML(text, false)
	text = fixHyperlinks(text)
	text = strings.TrimSpace(text)
	return text
}

func CleanupEnText(text string) string {
	if text == "" {
		return text
	}

	// 1) Remove \r
	text = strings.ReplaceAll(text, "\r", "")

	// 2) Preserve double-newlines
	text = strings.ReplaceAll(text, "\n\n", "DOUBLE_NEWLINE_PLACEHOLDER")
	// 3) Collapse multiple newlines into one
	text = reNewline.ReplaceAllString(text, "\n")
	// 4) Restore double newlines
	text = strings.ReplaceAll(text, "DOUBLE_NEWLINE_PLACEHOLDER", "\n\n")

	// 5) Collapse spaces
	text = reSpace.ReplaceAllString(text, " ")

	// 6) Fix hyperlinks
	text = fixHyperlinks(text)

	// 7) Standardize terms
	text = standardizeTerms(text)

	// 8) HTML-encode apostrophes
	text = strings.ReplaceAll(text, "'", "&#39;")

	// 9) Trim
	text = strings.TrimSpace(text)

	// 10) Wrap in <p> if not already
	if !(strings.HasPrefix(text, "<p>") && strings.HasSuffix(text, "</p>")) {
		text = "<p>" + text + "</p>"
	}

	// That’s it — no fixHTML call here!
	// This should produce text that's extremely close to your Python code.

	return text
}

func CleanupChapterTitle(text string) string {
	if text == "" {
		return text
	}
	text = reNewline.ReplaceAllString(text, "\n")
	text = reSpace.ReplaceAllString(text, " ")
	text = fixHTML(text, true)
	text = fixHyperlinks(text)
	text = strings.TrimSpace(text)
	return text
}

func CleanupEnChapterTitle(text string) string {
	if text == "" {
		return text
	}
	text = CleanupChapterTitle(text)
	text = standardizeTerms(text)
	return text
}
