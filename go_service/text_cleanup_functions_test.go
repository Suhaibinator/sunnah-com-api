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
			expected: `<p>This statement, may Allah have mercy on you, of accusation regarding the [<i>Mu’an’an</i>] chains is an invented one, produced without precedent, and there is no one who supports him from <i>Ahl ul-Ilm</i> in that. The widespread opinion, which is agreed upon between <i>Ahl ul-Ilm</i>, with reports and transmissions early and recent, is that each trustworthy narrator who transmits a Ḥadīth from his equal, with the feasible probability for [the transmitter] to meet [who he transmits from] and hear from him due to their being together in the same era, even if there never came a report that they met or spoke face to face, then the transmission is affirmed, and [using it as a] proof is appropriate, unless there is clear evidence that this transmitter did not meet who he transmits from or that he did not hear anything from him.\n\nThen as for when the matter is ambiguous regarding the possibility which we explained previously, then the transmission is always [accepted ]as coming by way of ‘hearing’ until there is evidence [otherwise] which we pointed out. Thus it is said to the inventor of this opinion whose speaker is as we have described, or to his defender- you have provided in the sum total of your statement that the report of the single trustworthy narrator on authority of the single trustworthy narrator is a proof which is required to act upon, then you introduced into it the condition afterwards, and you said ‘until we know that [the transmitter] had met once or more and heard something from [the one he transmits from]’. So have you found this condition which you stipulated from anyone [of <i>Ahl ul-Ilm</i>] who also required it? And if not then bring me evidence of what you allege. Thus if he claims there is a statement from one of the scholars of the Salaf for what he alleged in introducing the condition in affirming reports, [then] confirm it; [however] neither he, nor others, will ever find a way to produce it, even though he claims about what he alleges there is evidence to rely on. It is said ‘What is that evidence?’\n\nThus if he said: ‘I said it since I found transmitters of reports, early and recent, transmitting Ḥadīth from each other, and [the transmitter] did not ever see or hear anything from [from the one he transmits from]. Thus when I saw them permitting the transmission of Ḥadīth between them like this, <i>Irsāl</i>, without hearing [between transmitters], while the <i>Mursal</i> from the transmissions, in the foundation of our view and that of <i>Ahl ul-Ilm</i> in reports, is that it is not a proof; on account of what I described from the weakness, I rely on researching the hearing of the transmitter in each report on authority of [who he transmits from]. Thus when I unexpectedly come upon his hearing from [the one he transmits from] due to the low amount of a thing [i.e. transmissions on his authority], all of what he transmits on his authority becomes fixed to me thereafter. And if knowledge of [his actually hearing from whom he transmits from] is too distant from me, I withhold from the report and according to me it does not have a position of proof due to the possibility of <i>Irsāl</i> in it.’\n\n Thus it is said to him: Then if the reason for your weakening the [<i>Mu’an’an</i>] report and your abandoning relying on it is due to the possibility of <i>Irsāl in it</i>, it obligates you to not affirm a chain of <i>Mu’an’an</i> until you see it has hearing [<i>Simā’</i>] from its first [transmitter] to its last.\n\nAnd according to us it is possible that the Ḥadīth [you described] which has come to us on authority of Hishām bin Urwah, on authority of his father, on authority of Ā’ishah- we know with certainty that Hishām heard from his father and that his father heard from Ā’ishah, just as we know that Ā’ishah heard from the Prophet, peace and blessings upon him- it is possible that when Hishām does not say in the transmission that he transmits on authority of his father the words ‘I heard’ or ‘He informed me’, that there could be between him and his father another person who informed [Hishām] of it on authority of his father in this transmission, and he did not hear it from his father when he preferred transmitting it <i>Mursal</i>, and it is not attributed to who he really heard it from.\n\nJust as that is possible from Hishām, on authority of his father, then it is also possible for his father on authority of Ā’ishah, and like that all chains for Ḥadīth in which the ‘hearing’ [of each transmitter] from the other is not mentioned. And if it was known in some transmissions that every single one of them did hear from his companion a great deal, then it is still possible for each one of them to drop in some of the transmissions, such that he hears from someone else some of his Ḥadīth, then expedites on authority [of his most famous companion] occasionally, while not designating who he [actually] heard from. And at times he is afraid and designates who he [actually] related the Ḥadīth from and abandons <i>Irsāl</i>. What we mention from this is found in Ḥadīth, from the actions of trustworthy Muhaddithīn and A’immah of <i>Ahl ul-Ilm</i>; and we will mention several of their transmissions upon the pathway which we mentioned demonstrating through them the great amount of [the above], if Allah, exalted is He, wills. Thus from that [are the following]:\n\nThat Ayyūb as-Sakhtiyānī, Ibn al-Mubārak, Wakī’, Ibn Numayr, and a group of others transmitted on authority of Hishām bin Urwah, on authority of his father, on authority of Ā’ishah, may Allah be pleased with her, she said: ‘I applied scent to the Messenger of Allah, peace and blessings upon him, at the time of entering and leaving <i>Ihrām</i>, with the most pleasant [scent] I found’.\n\nThus Layth bin Sa’d, Dāwud al-Attār, Humayd bin al-Aswad, Wuhayb bin Khālid, and Abū Usāmah transmitted this transmission on authority of none other than Hishām, he said, Uthmān bin Urwah informed me, on authority of Urwah, on authority of Ā’ishah, on authority of the Prophet, peace and blessings upon him; and Hishām transmitted, on authority of his father, on authority of Ā’ishah, she said: ‘The Prophet, peace and blessings upon him, when he was in ‘Itikaf lowered his head towards me, then I combed [his hair] and I was menstruating’. Then Mālik bin Anas transmitted the exact narration, on authority of az-Zuhrī, on authority of Urwah, on authority of Amrah , on authority of Ā’ishah, on authority of the Prophet, peace and blessings upon him.\n\nAz-Zuhrī and Sālih bin Abī Hassān transmitted on authority of Abī Salamah, on authority of Ā’ishah: ‘The Prophet, peace and blessings upon him, would kiss while fasting’.\n\nThus Yahyā bin Abī Kathīr said about this report regarding ‘kissing’, Abū Salamah bin Abd ar-Rahman informed me that Umar bin Abd al-Azīz informed him that Urwah informed him that Ā’ishah informed him that: ‘The Prophet, peace and blessings upon him, would kiss her while fasting’.\n\nIbn Uyaynah and others transmitted on authority of Amr bin Dīnār, on authority of Jābir, he said ‘The Messenger of Allah, peace and blessings upon him, [allowed us] to eat horse meat and prohibited us from donkey meat’. And Hammād bin Zayd transmitted it, on authority of Amr, on authority of Muhammad bin Alī, on authority of Jābir, on authority of the Prophet, peace and blessings upon him. And this manner of transmitting narrations is abundant, its enumeration being much, and what we mentioned is sufficient for those who possess understanding. Thus when the reason [for weakening these types of transmissions]- according to the one whose opinion we described before in terms of the invalidation of Ḥadīth and weakening them when it is not known that the transmitter heard anything through the one he transmits from- is that <i>Irsāl</i> is possible in them, his opinion leads to his being obligated to abandon relying on transmissions of those who are known to have heard through who they transmit from unless there is mention of ‘hearing’ in the report itself, due to what we clarified before of the A’immah who related reports that at times they would expedite the Ḥadīth as <i>Irsāl</i>, and not mention who they heard it from, and at times they would be so inclined, so they would provide the chain for the report in the form that they heard it- they would report [a narration] through ‘descent’ [from a peer or someone below them in age or status] if it was descended and with ‘elevation’ [with less narrators between them and the Prophet, peace and blessings upon him] if it was elevated, just as we explained about them. We are not aware of anyone from the <i>A’immah</i> of the <i>Salaf</i> who when he sought to act upon reports and investigate the soundness or weakness of the chains of transmission like [those of] Ayyūb as-Sakhtiyānī, Ibn Awn, Mālik bin Anas, Shu’bah bin al-Hajjāj, Yahyā bin Sa’īd al-Qattān, Abd ar-Rahman bin Mahdī and those after them from the people of Ḥadīth, he examined the situation regarding [the manner of] ‘hearing’ in the chains, like what is claimed in the opinion of the one we described previously.\n\nThose who investigated among [the scholars of Ḥadīth] would only investigate the ‘hearing’ of the transmitters of Ḥadīth they transmitted from when the transmitter was among those who were known for <i>Tadlīs</i> in Ḥadīth and famous for it. Thus when they investigated [a transmitter’s manner of] ‘hearing’ in his transmissions and they would research that about him in order to distance themselves from the defect of <i>Tadlīs</i>. Thus to research that about the non-<i>Mudallis</i>, from the perspective of the one who alleged what he did in the opinion we related, then we have not heard of that from anyone we designated and do not designate from the <i>A’immah</i>.\n\nThus from that is Abd Allah bin Yazīd al-Ansārī , who saw the Prophet, peace and blessings upon him; he transmitted a Ḥadīth on authority of Hudhayfah and Abī Mas’ūd al-Ansārī attributing it to the Prophet, peace and blessings upon him, and there is no mention of ‘hearing’ in his transmission from either of them . Also, we have not preserved in any of the transmissions that Abd Allah bin Yazīd ever met Hudhayfah or Abū Mas’ūd face to face for Ḥadīth. We have not found mention in an actual transmission his seeing either of them and we have not heard from any of <i>Ahl ul-Ilm</i> who have passed or who we have met who charged with weakness these two reports who Abd Allah bin Yazīd transmitted on authority of Hudhayfah and Abū Mas’ūd. Rather according to those we met from <i>Ahl ul-Ilm</i> in Ḥadīth those two [reports] and whatever is similar to them are among the authentic and strong chains; they held the view of acting by what was related by them, and relied upon what came from the <i>Sunan</i> and <i>Āthār</i> [in that manner]. And it is weak and abandoned in the allegation of the one whose view we related before, until ‘hearing’ of the transmitter is obtained from whoever transmits [them]. And even if we took to enumerating the authentic reports according to <i>Ahl ul-Ilm</i> whereof they are weak in the allegation of this speaker and we counted them, truly we would not be able to fully examine its mention and enumerate all of them; rather we prefer to place several as a symbol for what we remain silent on.\n\nAbū Uthmān an-Nahdī and Abū Rāfi’ as-Sā’igh both were from among those who witnessed the age of Jahiliyyah [the time before Islam in the Arabian Peninsula] and were among the Companions of the Messenger of Allah, peace and blessings upon him, who witnessed the battle of Badr, and so on and so forth. They both related reports on authority of [the Companions] until they [related Ḥadīth from younger Companions] the likes of Abū Hurayrah and Ibn Umar. Each of these two transmitted a single Ḥadīth on authority of Ubayy bin K’ab, on authority of the Prophet, peace and blessings upon him, and we did not hear in an actual transmission that they had seen Ubayy with their own eyes, or heard anything from him.\n\nAbū Amr ash-Shaybānī witnessed <i>al-Jahiliyyah</i> and was an adult during the time of the Prophet, peace and blessings upon him, and Abū Ma’mar Abd Allah bin Sakhbarah each transmitted two reports on authority of Abū Mas’ūd al-Ansārī, on authority of the Prophet, peace and blessings upon him.\n\nUbayd bin Umayr transmitted a Ḥadīth on authority of Umm Salamah, wife of the Prophet, peace and blessings upon him, on authority of the Prophet, peace and blessings upon him, and Ubayd bin Umayr was born in the time of the Prophet, peace and blessings upon him.\n\nQays bin Abī Hāzim transmitted three reports on authority of Abū Mas’ūd al-Ansārī, on authority of the Prophet, peace and blessings upon him and he witnessed the time of the Prophet, peace and blessings upon him.\n\nAbd ar-Rahman bin Abī Laylā transmitted a Ḥadīth on authority of Anas bin Mālik, on authority of the Prophet, peace and blessings upon him, and he heard from Umar bin al-Khattāb and accompanied Alī.\n\nRib’ī bin Hirāsh transmitted two Ḥadīth on authority of Imrān bin Husayn, on authority of the Prophet, peace and blessings upon him; and a Ḥadīth on authority of Abū Bakrah, on authority of the Prophet, peace and blessings upon him. Rib’ī heard from Alī bin Abī Tālib and transmitted on his authority.\n\nNāfi’ bin Jubayr bin Mut’im transmitted a Ḥadīth on authority of Abī Shurayh al-Khuzā’ī, on authority of the Prophet, peace and blessings upon him.\n\nAn-Nu’mān bin Abī Ayyāsh transmitted three Ahādīth on authority of Abū Sa’īd al-Khudrī, on authority of the Prophet, peace and blessings upon him.\n\nAtā’ bin Yazīd al-Laythī transmitted a Ḥadīth on authority of Tamīm ad-Dārī, on authority of the Prophet, peace and blessings upon him.\n\nSulaymān bin Yasār transmitted a Ḥadīth on authority of Rāfi’ bin Khadīj, on authority of the Prophet, peace and blessings upon him.\n\nHumayd bin Abd ar-Rahman al-Himyarī transmitted narrations on authority of Abū Hurayrah, on authority of the Prophet, peace and blessings upon him. Thus all of these Tabi’īn we named, whose transmissions are on authority of Companions, are not recorded in separate transmissions to have heard directly from them, to our knowledge, and are not recorded to have met them in the course of the actual report. They are sound chains of transmission according to those who possess knowledge of reports and transmissions; we do not know of them ever weakening anything of them or asking about whether they heard from each other, since the ‘hearing’ of each one of them from his companion is possible, without anyone rejecting [that], due to them all being together in the same time period.\n\nThis opinion that the speaker invented, which we related, regarding weakening the Ḥadīth, for the reason which he described, is too inferior to be relied upon or [too inferior] for its mention to be stirred up since it was an invented opinion and a backward discussion which no one from <i>Ahl ul-Ilm</i> stated before and those who came after them denounced it. Thus there is no need to for us to refute it with more than what we have already explained since the standing of the speech and its speaker is that which we described, and Allah is the one with whom aid is sought in repelling what differs from the school of the scholars and in Him alone complete trust is placed.</p>`},
		{
			name:     "Clean HTML and standardize",
			text:     "<p>he Prophet said and Allah's Messenger mentioned</p>",
			expected: "<p>he Prophet (ﷺ) said and Allah&#39;s Messenger (ﷺ) mentioned</p>",
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
