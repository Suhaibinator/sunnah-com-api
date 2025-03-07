package go_api

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Suhaibinator/api/go_persistence"
	"github.com/Suhaibinator/api/go_service"
)

type CollectionMeta struct {
	Language   string `json:"lang"`
	Title      string `json:"title"`
	ShortIntro string `json:"shortIntro"`
}

type Collection struct {
	Name                 string           `json:"name"`
	HasBooks             bool             `json:"hasBooks"`
	HasChapters          bool             `json:"hasChapters"`
	CollectionMeta       []CollectionMeta `json:"collection"`
	TotalHadith          int              `json:"totalHadith"`
	TotalAvailableHadith int              `json:"totalAvailableHadith"`
}

func ConvertDbCollectionToApiCollection(dbCollection *go_persistence.HadithCollection) *Collection {
	// Note: Python version checks for shortIntroArabic field, but we don't have it in our struct
	// so we use the same shortIntro for both languages

	collection := Collection{
		Name:        dbCollection.Name,
		HasBooks:    dbCollection.HasBooks == "yes",
		HasChapters: dbCollection.HasChapters == "yes",
		CollectionMeta: []CollectionMeta{
			{Language: "en", Title: dbCollection.EnglishTitle, ShortIntro: dbCollection.ShortIntro},
			{Language: "ar", Title: dbCollection.ArabicTitle, ShortIntro: dbCollection.ShortIntro},
		},
		TotalHadith:          *dbCollection.TotalHadith,
		TotalAvailableHadith: dbCollection.NumHadith,
	}

	return &collection
}

type PaginatedCollections struct {
	Collections []Collection `json:"data"`
	Total       int          `json:"total"`
	Limit       int          `json:"limit"`
	PrevPage    int          `json:"previous"`
	NextPage    int          `json:"next"`
}

type BookMeta struct {
	Language string `json:"lang"`
	Name     string `json:"name"`
}

type Book struct {
	BookNumber        string     `json:"bookNumber"`
	BookMeta          []BookMeta `json:"book"`
	HadithStartNumber int        `json:"hadithStartNumber"`
	HadithEndNumber   int        `json:"hadithEndNumber"`
	NumberOfHadith    int        `json:"numberOfHadith"`
}

func ConvertDbBookToApiBook(dbBook *go_persistence.Book) *Book {
	if dbBook == nil {
		return nil
	}

	var englishName, arabicName string
	if dbBook.EnglishName != nil {
		englishName = *dbBook.EnglishName
	}
	if dbBook.ArabicName != nil {
		arabicName = *dbBook.ArabicName
	}

	book := Book{
		BookNumber: go_service.GetBookNumberFromBookId(dbBook.OurBookID),
		BookMeta: []BookMeta{
			{Language: "en", Name: englishName},
			{Language: "ar", Name: arabicName},
		},
		HadithStartNumber: dbBook.FirstNumber,
		HadithEndNumber:   dbBook.LastNumber,
		NumberOfHadith:    dbBook.TotalNumber,
	}
	return &book
}

type PaginatedBooks struct {
	Books []Book `json:"data"`
	Total int    `json:"total"`
	Limit int    `json:"limit"`
	Prev  *int   `json:"previous"`
	Next  *int   `json:"next"`
}

type ChapterMeta struct {
	Language      string  `json:"lang"`
	ChapterNumber string  `json:"chapterNumber"`
	ChapterTitle  string  `json:"chapterTitle"`
	Intro         *string `json:"intro"`
	Ending        *string `json:"ending"`
}

type Chapter struct {
	BookNumber  string        `json:"bookNumber"`
	ChapterId   string        `json:"chapterId"`
	ChapterMeta []ChapterMeta `json:"chapter"`
}

func ConvertDbChapterToApiChapter(dbChapter *go_persistence.Chapter) *Chapter {
	// Format the babID to maintain the decimal format (e.g., "61.00")
	chapterId := fmt.Sprintf("%.2f", dbChapter.BabID)

	// Use the book number from arabicBookID
	bookNumber := fmt.Sprintf("%.0f", dbChapter.ArabicBookID)

	// Create pointers to the strings
	englishIntro := dbChapter.EnglishIntro
	englishEnding := dbChapter.EnglishEnding
	arabicIntro := dbChapter.ArabicIntro
	arabicEnding := dbChapter.ArabicEnding

	chapter := Chapter{
		BookNumber: bookNumber,
		ChapterId:  chapterId,
		ChapterMeta: []ChapterMeta{
			{
				Language:      "en",
				ChapterNumber: dbChapter.EnglishBabNumber,
				ChapterTitle:  dbChapter.EnglishBabName,
				Intro:         &englishIntro,
				Ending:        &englishEnding,
			},
			{
				Language:      "ar",
				ChapterNumber: dbChapter.ArabicBabNumber,
				ChapterTitle:  dbChapter.ArabicBabName,
				Intro:         &arabicIntro,
				Ending:        &arabicEnding,
			},
		},
	}
	return &chapter
}

type PaginatedChapters struct {
	Chapters []Chapter `json:"data"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Prev     *int      `json:"previous"`
	Next     *int      `json:"next"`
}

type HadithGradedBy struct {
	Grader string `json:"graded_by"`
	Grade  string `json:"grade"`
}

type HadithMeta struct {
	Language      string           `json:"lang"`
	ChapterNumber string           `json:"chapterNumber"`
	ChapterTitle  string           `json:"chapterTitle"`
	Urn           int              `json:"urn"`
	Body          string           `json:"body"`
	Grades        []HadithGradedBy `json:"grades"`
}

type Hadith struct {
	Collection   string       `json:"collection"`
	BookNumber   string       `json:"bookNumber"`
	ChapterId    string       `json:"chapterId"`
	HadithNumber string       `json:"hadithNumber"`
	HadithMeta   []HadithMeta `json:"hadith"`
}

// parseGrades attempts to parse grade data as JSON, with fallback to individual fields
func parseGrades(gradeValue string, collectionGrader string) []HadithGradedBy {
	if gradeValue == "" {
		return []HadithGradedBy{}
	}

	// Try to parse as JSON
	var grades []map[string]interface{}
	err := json.Unmarshal([]byte(gradeValue), &grades)
	if err == nil {
		// Successfully parsed as JSON
		result := make([]HadithGradedBy, len(grades))
		for i, grade := range grades {
			grader, _ := grade["graded_by"].(string)
			gradeVal, _ := grade["grade"].(string)
			result[i] = HadithGradedBy{
				Grader: grader,
				Grade:  gradeVal,
			}
		}
		return result
	}

	// Fallback to individual field
	return []HadithGradedBy{
		{
			Grader: collectionGrader,
			Grade:  gradeValue,
		},
	}
}

func ConvertDbHadithToApiHadith(dbHadith *go_persistence.Hadith) *Hadith {
	// Format the babID to maintain the decimal format (e.g., "61.00")
	babIDFloat, _ := strconv.ParseFloat(dbHadith.BabID, 64)
	chapterId := fmt.Sprintf("%.2f", babIDFloat)

	// Parse grades for English and Arabic
	englishGrades := parseGrades(dbHadith.EnglishGrade1, "")
	arabicGrades := parseGrades(dbHadith.ArabicGrade1, "")

	hadith := Hadith{
		Collection:   dbHadith.Collection,
		BookNumber:   dbHadith.BookNumber,
		ChapterId:    chapterId,
		HadithNumber: dbHadith.HadithNumber,
		HadithMeta: []HadithMeta{
			{
				Language:      "en",
				ChapterNumber: dbHadith.EnglishBabNumber,
				ChapterTitle:  dbHadith.EnglishBabName,
				Urn:           dbHadith.EnglishURN,
				Body:          dbHadith.EnglishText,
				Grades:        englishGrades,
			},
			{
				Language:      "ar",
				ChapterNumber: dbHadith.ArabicBabNumber,
				ChapterTitle:  dbHadith.ArabicBabName,
				Urn:           dbHadith.ArabicURN,
				Body:          dbHadith.ArabicText,
				Grades:        arabicGrades,
			},
		},
	}
	return &hadith
}

type PaginatedHadiths struct {
	Hadiths []Hadith `json:"data"`
	Total   int      `json:"total"`
	Limit   int      `json:"limit"`
	Prev    *int     `json:"previous"`
	Next    *int     `json:"next"`
}
