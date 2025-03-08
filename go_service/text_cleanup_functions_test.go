package go_service

import (
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestNegativeLookBackReplacement(t *testing.T) {
	tests := []struct {
		name        string
		re          *regexp.Regexp
		text        string
		replacement string
		expected    string
	}{
		{
			name:        "Replace when not followed by '('",
			re:          regexp.MustCompile(`Allah's Messenger `),
			text:        "Allah's Messenger said something. Allah's Messenger (PBUH) said something else.",
			replacement: "Allah's Messenger (\ufdfa) ",
			expected:    "Allah's Messenger (\ufdfa) said something. Allah's Messenger (PBUH) said something else.",
		},
		{
			name:        "No matches",
			re:          regexp.MustCompile(`NonExistentPattern`),
			text:        "This text doesn't contain the pattern.",
			replacement: "Replacement",
			expected:    "This text doesn't contain the pattern.",
		},
		{
			name:        "Multiple replacements",
			re:          regexp.MustCompile(`he Prophet `),
			text:        "he Prophet said this. he Prophet did that. he Prophet (PBUH) mentioned this.",
			replacement: "he Prophet (\ufdfa) ",
			expected:    "he Prophet (\ufdfa) said this. he Prophet (\ufdfa) did that. he Prophet (PBUH) mentioned this.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := negativeLookBackReplacement(tt.re, tt.text, tt.replacement)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOuterHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		expected string
		hasError bool
	}{
		{
			name:     "Simple paragraph",
			html:     "<div><p>Test paragraph</p></div>",
			selector: "p",
			expected: "<p>Test paragraph</p>",
			hasError: false,
		},
		{
			name:     "Element with attributes",
			html:     "<div><a href='https://example.com' id='link'>Link</a></div>",
			selector: "a",
			expected: "<a href=\"https://example.com\" id=\"link\">Link</a>",
			hasError: false,
		},
		{
			name:     "Nested elements",
			html:     "<div><span>Outer <em>inner</em> text</span></div>",
			selector: "span",
			expected: "<span>Outer <em>inner</em> text</span>",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err)

			selection := doc.Find(tt.selector)
			result, err := outerHTML(selection)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFixHTML(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		removeWrapper bool
		expected      string
	}{
		{
			name:          "No HTML tags",
			text:          "Plain text without HTML",
			removeWrapper: false,
			expected:      "Plain text without HTML",
		},
		{
			name:          "Remove wrapper true",
			text:          "<p>Text inside paragraph</p>",
			removeWrapper: true,
			expected:      "Text inside paragraph",
		},
		{
			name:          "Remove wrapper false",
			text:          "<p>Text inside paragraph</p>",
			removeWrapper: false,
			expected:      "<p>Text inside paragraph</p>",
		},
		{
			name:          "Remove c_q tags",
			text:          "<p>Text with <c_q10>special</c_q10> tags</p>",
			removeWrapper: true,
			expected:      "Text with special tags",
		},
		{
			name:          "Clean anchor attributes",
			text:          "<p>Text with <a href='link' id='remove' name='remove'>link</a></p>",
			removeWrapper: true,
			expected:      "Text with <a href=\"link\">link</a>",
		},
		{
			name:          "Whitespace and carriage returns",
			text:          "  Text with \r\ncarriage returns  ",
			removeWrapper: false,
			expected:      "Text with \ncarriage returns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixHTML(tt.text, tt.removeWrapper)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStandardizeTerms(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Replace PBUH",
			text:     "The Prophet PBUH said",
			expected: "The Prophet (\ufdfa) \ufdfa said",
		},
		{
			name:     "Replace P.B.U.H.",
			text:     "The Prophet P.B.U.H. said",
			expected: "The Prophet (\ufdfa) \ufdfa said",
		},
		{
			name:     "Replace peace_be_upon_him",
			text:     "The Prophet peace_be_upon_him said",
			expected: "The Prophet (\ufdfa) \ufdfa said",
		},
		{
			name:     "Replace (may peace be upon him)",
			text:     "The Prophet (may peace be upon him) said",
			expected: "The Prophet (\ufdfa) said",
		},
		{
			name:     "Replace Allah's Apostle",
			text:     "Allah's Apostle said",
			expected: "Allah's Messenger (\ufdfa) said",
		},
		{
			name:     "Replace he Apostle of Allah",
			text:     "he Apostle of Allah said",
			expected: "he Messenger of Allah (\ufdfa) said",
		},
		{
			name:     "Replace he Holy Prophet",
			text:     "he Holy Prophet said",
			expected: "he Prophet (\ufdfa) said",
		},
		{
			name:     "Add PBUH to Allah's Messenger",
			text:     "Allah's Messenger said something",
			expected: "Allah's Messenger (\ufdfa) said something",
		},
		{
			name:     "Don't add PBUH when followed by (",
			text:     "Allah's Messenger (PBUH) said something",
			expected: "Allah's Messenger (\ufdfa) said something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := standardizeTerms(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFixHyperlinks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Replace relative links",
			text:     "<a href=\"/path/to/page\">Link</a>",
			expected: "<a href=\"https://sunnah.com/path/to/page\">Link</a>",
		},
		{
			name:     "Replace javascript:openquran",
			text:     "javascript:openquran(2,255,255)",
			expected: "https://quran.com/3/255-255",
		},
		{
			name:     "Multiple openquran links",
			text:     "First: javascript:openquran(1,1,7) Second: javascript:openquran(2,1,5)",
			expected: "First: https://quran.com/2/1-7 Second: https://quran.com/3/1-5",
		},
		{
			name:     "Invalid openquran format",
			text:     "Invalid: javascript:openquran(invalid)",
			expected: "Invalid: Invalid: javascript:openquran(invalid)",
		},
		{
			name:     "No links to replace",
			text:     "Text without any links",
			expected: "Text without any links",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixHyperlinks(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "Multiple newlines and spaces",
			text:     "Text with  multiple \n\n spaces   and newlines",
			expected: "Text with multiple \n spaces and newlines",
		},
		{
			name:     "HTML cleanup",
			text:     "<p id=\"remove\">Text with <a id=\"remove\" name=\"remove\" href=\"/link\">link</a></p>",
			expected: "<p id=\"remove\">Text with <a href=\"https://sunnah.com/link\">link</a></p>",
		},
		{
			name:     "Trim whitespace",
			text:     "  Text with whitespace  ",
			expected: "Text with whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanupText(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupEnText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: "",
		},
		{
			name: "Muslim Intro",
			text: `Know - may Allah, exalted is He, grant you success – that what is obligatory upon everyone who is aware of the distinction between the Sahīh transmissions and their weak, the trustworthy narrators from those who stand accused, is to not transmit from them except what is known for the soundness of its emergence and the protection of its narrators; and that they fear what may be from those accused (of deficiency in narrating) and the stubborn people of innovation.

The proof that what we have said is required above what opposes it is in the verse: <b><i>{Oh you who believe! If a sinful person comes to you with news, then verify it lest you afflict people through ignorance then you become sorry about what you did}</b>[al-Hujurāt: 6]</i>; and the verse: <b><i>{… from whom you are pleased with from the witnesses}</b>[al-Baqarah: 282]</i> and the verse: <b><i>{And let two who possess integrity among you bare witness}</b>[at-Talāq: 2]</i>. Thus it demonstrates what we mentioned from these two verses that the report of the sinful is dropped and not accepted, and that the testimony [Shahādah] of one who does not possess integrity is rejected, and the report [Khabar] as well- even though its significance is separated from the meaning of testimony in some respects, they are in agreement regarding the overall conditions they share since the report of the sinful is not acceptable according to Ahl ul-Ilm just as his testimony is rejected according to all of them . The Sunnah demonstrates the prohibition of transmitting abominable transmissions just as in the example from the Qur’ān regarding the prohibition of the report of the sinful.

There is a famous narration on authority of the Messenger of Allah, peace and blessings of Allah upon him, that: ‘Whoever relates on my authority a narration while aware that it is a lie, then he is one of the liars’ . Abū Bakr ibn Abī Shaybah narrated it to us that Wakī narrated to us, on authority of Shu’bah, on authority of al-Hakam, on authority of Abd ir-Rahman ibn Abī Laylā, on authority of Samurah bin Jundab. And also Abū Bakr ibn Abī Shaybah narrated to us, that Wakī narrated to us, on authority of Shu’bah and Sufyān, on authority of Habīb, on authority of Maymūn ibn Abī Shabīb, on authority of al-Mughīrat ibn Shu’bah, they both said that the Messenger of Allah, peace and blessings of Allah upon him, said the same thing.`,
			expected: `<p>Know - may Allah, exalted is He, grant you success – that what is obligatory upon everyone who is aware of the distinction between the Sahīh transmissions and their weak, the trustworthy narrators from those who stand accused, is to not transmit from them except what is known for the soundness of its emergence and the protection of its narrators; and that they fear what may be from those accused (of deficiency in narrating) and the stubborn people of innovation.\n\nThe proof that what we have said is required above what opposes it is in the verse: <b><i>{Oh you who believe! If a sinful person comes to you with news, then verify it lest you afflict people through ignorance then you become sorry about what you did}</i></b>[al-Hujurāt: 6]; and the verse: <b><i>{… from whom you are pleased with from the witnesses}</i></b>[al-Baqarah: 282] and the verse: <b><i>{And let two who possess integrity among you bare witness}</i></b>[at-Talāq: 2]. Thus it demonstrates what we mentioned from these two verses that the report of the sinful is dropped and not accepted, and that the testimony [Shahādah] of one who does not possess integrity is rejected, and the report [Khabar] as well- even though its significance is separated from the meaning of testimony in some respects, they are in agreement regarding the overall conditions they share since the report of the sinful is not acceptable according to Ahl ul-Ilm just as his testimony is rejected according to all of them . The Sunnah demonstrates the prohibition of transmitting abominable transmissions just as in the example from the Qur’ān regarding the prohibition of the report of the sinful.\n\nThere is a famous narration on authority of the Messenger of Allah, peace and blessings of Allah upon him, that: ‘Whoever relates on my authority a narration while aware that it is a lie, then he is one of the liars’ . Abū Bakr ibn Abī Shaybah narrated it to us that Wakī narrated to us, on authority of Shu’bah, on authority of al-Hakam, on authority of Abd ir-Rahman ibn Abī Laylā, on authority of Samurah bin Jundab. And also Abū Bakr ibn Abī Shaybah narrated to us, that Wakī narrated to us, on authority of Shu’bah and Sufyān, on authority of Habīb, on authority of Maymūn ibn Abī Shabīb, on authority of al-Mughīrat ibn Shu’bah, they both said that the Messenger of Allah, peace and blessings of Allah upon him, said the same thing.</p>`},
		{
			name:     "Clean HTML and standardize",
			text:     "<p>he Prophet said and Allah's Messenger mentioned</p>",
			expected: "<p>he Prophet (\ufdfa) said and Allah&#39;s Messenger mentioned</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanupEnText(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupChapterTitle(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "Multiple newlines and spaces",
			text:     "Chapter  title with \n\n multiple   spaces",
			expected: "Chapter title with \n multiple spaces",
		},
		{
			name:     "Remove paragraph wrapper",
			text:     "<p>Chapter title</p>",
			expected: "Chapter title",
		},
		{
			name:     "Fix hyperlinks",
			text:     "<p>Chapter with <a href=\"javascript:openquran(1,1,7)\">link</a></p>",
			expected: "Chapter with <a href=\"https://quran.com/2/1-7\">link</a>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanupChapterTitle(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupEnChapterTitle(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "Standardize terms in chapter title",
			text:     "<p>Chapter about Allah's Apostle PBUH</p>",
			expected: "Chapter about Allah&#39;s Apostle \ufdfa",
		},
		{
			name:     "Clean HTML, fix links and standardize",
			text:     "<p>Chapter about <a href=\"/link\">he Prophet</a> and javascript:openquran(1,1,7)</p>",
			expected: "Chapter about <a href=\"https://sunnah.com/link\">he Prophet</a> and https://quran.com/2/1-7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanupEnChapterTitle(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}
