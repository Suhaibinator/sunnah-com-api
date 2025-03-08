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

	reOpenQuran = regexp.MustCompile(`javascript:openquran\((.+?)\)`)
	reNewline   = regexp.MustCompile(`\n+`)
	reSpace     = regexp.MustCompile(` +`)
)

func negativeLookBackReplacement(re *regexp.Regexp, text string, replacement string) string {
	// Find all matches with their indices
	matches := re.FindAllStringIndex(text, -1)
	if matches == nil {
		return text
	}

	var result strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// Append text before the match
		result.WriteString(text[lastIndex:start])

		// Check if the next character is '('
		if end < len(text) && text[end] == '(' {
			// Do not replace; write the original match
			result.WriteString(text[start:end])
		} else {
			// Replace with the given replacement
			result.WriteString(replacement)
		}

		lastIndex = end
	}

	// Append the rest of the text after the last match
	result.WriteString(text[lastIndex:])
	return result.String()
}

// outerHTML returns the outer HTML of a goquery Selection
func outerHTML(s *goquery.Selection) (string, error) {
	var buf bytes.Buffer
	for _, node := range s.Nodes {
		if err := html.Render(&buf, node); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// fixHTML cleans up HTML text by removing unnecessary attributes, tags, and whitespace.
// It takes a string 'text' and a boolean 'removeWrapper'. If 'removeWrapper' is true,
// it removes wrapping <p> tags from the beginning and end of the text.
func fixHTML(text string, removeWrapper bool) string {
	// Remove leading and trailing whitespace
	text = strings.TrimSpace(text)

	// Remove carriage return characters
	text = strings.ReplaceAll(text, "\r", "")

	// Check if text contains any HTML tags
	if !strings.Contains(text, "<") || !strings.Contains(text, ">") {
		return text // Return as is if no HTML tags are found
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		// If parsing fails, return the original trimmed text
		return text
	}

	// If the document body is empty or contains only whitespace, return original text
	bodyText := strings.TrimSpace(doc.Find("body").Text())
	if bodyText == "" {
		return text
	}

	// Remove 'id' and 'name' attributes from all anchor tags
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("id")
		s.RemoveAttr("name")
	})

	// Collect the HTML of the direct children of the body tag
	var builder strings.Builder
	doc.Find("body").Children().Each(func(i int, s *goquery.Selection) {
		// Skip elements without text content
		if strings.TrimSpace(s.Text()) == "" {
			return
		}
		htmlStr, err := outerHTML(s)
		if err != nil {
			return
		}
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(htmlStr)
	})

	// Get the final text from the builder
	result := builder.String()

	// If the result is empty (e.g., all HTML elements were removed), return the original text
	if result == "" {
		return text
	}

	// Optionally remove wrapping <p> tags from the start and end
	if removeWrapper {
		result = reHtmlParagraphTag.ReplaceAllString(result, "")
	}

	// Remove tags like <c_q10> or </c_q10>
	result = reCqTag.ReplaceAllString(result, "")

	return result
}

// standardizeTerms standardizes specific terms in the text
func standardizeTerms(text string) string {
	terms := []struct {
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

	for _, term := range terms {
		text = strings.ReplaceAll(text, term.old, term.new)
	}

	text = negativeLookBackReplacement(reAllahsMessenger, text, "Allah's Messenger (\ufdfa) ")
	text = negativeLookBackReplacement(reMessengerOfAllah, text, "he Messenger of Allah (\ufdfa) ")
	text = negativeLookBackReplacement(reProphet, text, "he Prophet (\ufdfa) ")

	return text
}

func fixHyperlinks(text string) string {
	// Replace 'href="/' with 'href="https://sunnah.com/'
	text = strings.ReplaceAll(text, `href="/`, `href="https://sunnah.com/`)

	// Define the regex pattern to find 'javascript:openquran(...)'

	// Use FindAllStringSubmatchIndex to get the indices of all matches
	matches := reOpenQuran.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return text // No matches, return as is
	}

	// Build the new text
	var result strings.Builder
	lastIndex := 0
	for _, match := range matches {
		// match is [start,end, group1start, group1end]
		start := match[0]
		end := match[1]
		group1start := match[2]
		group1end := match[3]

		// Append text before the match
		result.WriteString(text[lastIndex:start])

		linkMatch := text[group1start:group1end]
		parts := strings.Split(linkMatch, ",")
		if len(parts) != 3 {
			log.Printf("Invalid link match: %s\n", linkMatch)
			continue
		}
		surahStr := strings.TrimSpace(parts[0])
		begin := strings.TrimSpace(parts[1])
		endVerse := strings.TrimSpace(parts[2])

		surahInt, err := strconv.Atoi(surahStr)
		if err != nil {
			log.Printf("Invalid surah number: %s\n", surahStr)
			continue
		}
		surahInt += 1

		newURL := fmt.Sprintf("https://quran.com/%d/%s-%s", surahInt, begin, endVerse)
		result.WriteString(newURL)

		lastIndex = end
	}
	// Append the rest of the text
	result.WriteString(text[lastIndex:])

	return result.String()
}

// CleanupText cleans up text by removing unnecessary whitespace and fixing HTML
func CleanupText(text string) string {
	if text == "" {
		return text
	}

	text = reNewline.ReplaceAllString(text, "\n")
	text = reSpace.ReplaceAllString(text, " ")
	text = fixHTML(text, false)
	text = fixHyperlinks(text)
	text = strings.TrimSpace(text)
	return text
}

// CleanupEnText cleans up English text and standardizes terms
func CleanupEnText(text string) string {
	if text == "" {
		return text
	}
	text = CleanupText(text)
	text = standardizeTerms(text)
	return text
}

// CleanupChapterTitle cleans up chapter titles
func CleanupChapterTitle(text string) string {
	if text == "" {
		return text
	}
	reNewline := regexp.MustCompile(`\n+`)
	reSpace := regexp.MustCompile(` +`)
	text = reNewline.ReplaceAllString(text, "\n")
	text = reSpace.ReplaceAllString(text, " ")
	text = fixHTML(text, true)
	text = fixHyperlinks(text)
	text = strings.TrimSpace(text)
	return text
}

// CleanupEnChapterTitle cleans up English chapter titles and standardizes terms
func CleanupEnChapterTitle(text string) string {
	if text == "" {
		return text
	}
	text = CleanupChapterTitle(text)
	text = standardizeTerms(text)
	return text
}
