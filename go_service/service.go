package go_service

import (
	"fmt"

	"github.com/Suhaibinator/api/go_persistence"
)

type ApplicationService struct {
	applicationPersistence *go_persistence.ApplicationPersistence
}

func NewApplicationService(applicationPersistence *go_persistence.ApplicationPersistence) *ApplicationService {
	return &ApplicationService{applicationPersistence: applicationPersistence}
}

func (as *ApplicationService) GetPaginatedHadithCollections(page int, limit int) ([]*go_persistence.HadithCollection, int64, error) {
	db_result, total, err := as.applicationPersistence.GetPaginatedHadithCollections(page, limit)
	return db_result, total, err
}

func (as *ApplicationService) GetHadithCollectionByName(name string) (*go_persistence.HadithCollection, error) {
	db_result, err := as.applicationPersistence.GetHadithCollectionByName(name)
	return db_result, err
}

func (as *ApplicationService) GetPaginatedBooksByCollection(collection string, page int, limit int) ([]*go_persistence.Book, int64, error) {
	db_result, total, err := as.applicationPersistence.GetPaginatedBooksByCollection(collection, page, limit)
	return db_result, total, err
}

func (as *ApplicationService) GetPaginatedHadithsByCollectionAndBookNumber(collection string, bookNumber string, page int, limit int) ([]*go_persistence.Hadith, int64, error) {
	db_result, total, err := as.applicationPersistence.GetPaginatedHadithsByCollectionAndBookNumber(collection, bookNumber, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, hadith := range db_result {
		processHadith(hadith)
	}
	return db_result, total, nil
}

func (as *ApplicationService) GetBookByCollectionAndBookNumber(collection string, bookNumber string) (*go_persistence.Book, error) {
	if bookNumber == "introduction" {
		bookNumber = "-1"
	} else if bookNumber == "35b" {
		bookNumber = "-35"
	}
	db_result, err := as.applicationPersistence.GetBookByCollectionAndBookNumber(collection, bookNumber)
	return db_result, err
}

func (as *ApplicationService) GetPaginatedChaptersByCollectionAndBookNumber(collection string, bookNumber string, page int, limit int) ([]*go_persistence.Chapter, int64, error) {
	db_result, total, err := as.applicationPersistence.GetPaginatedChaptersByCollectionAndBookNumber(collection, bookNumber, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return db_result, total, nil
}

func (as *ApplicationService) GetChapterByCollectionAndBookNumberAndChapterNumber(collection string, bookNumber string, chapterNumber string) (*go_persistence.Chapter, error) {
	db_result, err := as.applicationPersistence.GetChapterByCollectionAndBookNumberAndChapterNumber(collection, bookNumber, chapterNumber)
	if err != nil {
		return nil, err
	}
	return db_result, nil
}

func (as *ApplicationService) GetHadithByCollectionAndHadithNumber(collection, hadithNumber string) (*go_persistence.Hadith, error) {
	db_result, err := as.applicationPersistence.GetHadithByCollectionAndHadithNumber(collection, hadithNumber)
	if err != nil {
		return nil, err
	}
	processHadith(db_result)
	return db_result, nil
}

func (as *ApplicationService) GetHadithByUrn(urn int) (*go_persistence.Hadith, error) {
	db_result, err := as.applicationPersistence.GetHadithByUrn(urn)
	if err != nil {
		return nil, err
	}
	processHadith(db_result)
	return db_result, nil
}

func (as *ApplicationService) GetRandomHadith() (*go_persistence.Hadith, error) {
	db_result, err := as.applicationPersistence.GetRandomHadithInCollection("riyadussalihin")
	if err != nil {
		return nil, err
	}
	processHadith(db_result)
	return db_result, nil
}

var bookIdToNumber = map[int]string{
	-1:  "introduction",
	-35: "35b",
}

func GetBookNumberFromBookId(bookId int) string {
	if bookNumber, ok := bookIdToNumber[bookId]; ok {
		return bookNumber
	}
	return fmt.Sprint(bookId)
}

// Note: We no longer need to process chapters here as we're doing it in the API layer

func processHadith(hadith *go_persistence.Hadith) {
	hadith.EnglishText = CleanupEnText(hadith.EnglishText)
	hadith.ArabicText = CleanupText(hadith.ArabicText)

	hadith.EnglishBabName = CleanupEnChapterTitle(hadith.EnglishBabName)
	hadith.ArabicBabName = CleanupChapterTitle(hadith.ArabicBabName)
}
