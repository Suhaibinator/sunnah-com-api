package go_persistence

import "gorm.io/gorm"

type ApplicationPersistence struct {
	dB *gorm.DB
}

func NewApplicationPersistence(dB *gorm.DB) *ApplicationPersistence {
	return &ApplicationPersistence{dB: dB}
}

// getRandomFunction returns the appropriate random function based on the database dialect
func getRandomFunction(db *gorm.DB) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "RAND()"
	case "postgres":
		return "RANDOM()"
	case "sqlite", "sqlite3":
		return "RANDOM()"
	default:
		// Default to RANDOM(), which is common in many SQL dialects
		return "RANDOM()"
	}
}

func (ap *ApplicationPersistence) GetPaginatedHadithCollections(page int, limit int) ([]*HadithCollection, int64, error) {
	var collections []*HadithCollection
	var total int64

	// Get total count
	err := ap.dB.Model(&HadithCollection{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = ap.dB.Order("collectionID").Offset((page - 1) * limit).Limit(limit).Find(&collections).Error
	if err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

func (ap *ApplicationPersistence) GetHadithCollectionByName(name string) (*HadithCollection, error) {
	var collection HadithCollection
	err := ap.dB.Where("name = ?", name).First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (ap *ApplicationPersistence) GetPaginatedBooksByCollection(collection string, page int, limit int) ([]*Book, int64, error) {
	var books []*Book
	var total int64

	// Get total count
	err := ap.dB.Model(&Book{}).Where("collection = ? AND status = ?", collection, 4).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = ap.dB.Where("collection = ? AND status = ?", collection, 4).Order("ABS(ourBookID)").Offset((page - 1) * limit).Limit(limit).Find(&books).Error
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (ap *ApplicationPersistence) GetBookByCollectionAndBookNumber(collection string, bookNumber string) (*Book, error) {
	var book Book
	err := ap.dB.Where("collection = ? AND ourBookID = ? AND status = ?", collection, bookNumber, 4).First(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (ap *ApplicationPersistence) GetPaginatedChaptersByCollectionAndBookNumber(collection string, bookNumber string, page int, limit int) ([]*Chapter, int64, error) {
	var chapters []*Chapter
	var total int64

	// Get total count using ChapterData
	err := ap.dB.Model(&Chapter{}).
		Where("collection = ? AND arabicBookID = ?", collection, bookNumber).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results using ChapterData
	err = ap.dB.Where("collection = ? AND arabicBookID = ?", collection, bookNumber).
		Order("babID").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&chapters).Error
	if err != nil {
		return nil, 0, err
	}

	// Set the BookNumber field for each chapter
	for _, chapter := range chapters {
		chapter.BookNumber = bookNumber
	}

	return chapters, total, nil
}

func (ap *ApplicationPersistence) GetChapterByCollectionAndBookNumberAndChapterNumber(collection string, bookNumber string, chapterNumber string) (*Chapter, error) {
	var chapter Chapter
	err := ap.dB.Where("collection = ? AND arabicBookID = ? AND babID = ?", collection, bookNumber, chapterNumber).
		First(&chapter).Error
	if err != nil {
		return nil, err
	}

	// Set the BookNumber field
	chapter.BookNumber = bookNumber

	return &chapter, nil
}

func (ap *ApplicationPersistence) GetPaginatedHadithsByCollectionAndBookNumber(collection string, bookNumber string, page int, limit int) ([]*Hadith, int64, error) {
	var hadiths []*Hadith
	var total int64

	// Get total count
	err := ap.dB.Model(&Hadith{}).Where("collection = ? AND bookNumber = ?", collection, bookNumber).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = ap.dB.Where("collection = ? AND bookNumber = ?", collection, bookNumber).Order("englishURN").Offset((page - 1) * limit).Limit(limit).Find(&hadiths).Error
	if err != nil {
		return nil, 0, err
	}

	return hadiths, total, nil
}

func (ap *ApplicationPersistence) GetHadithByCollectionAndBookNumberAndChapterNumberAndHadithNumber(collection string, bookNumber string, chapterNumber string, hadithNumber string) (*Hadith, error) {
	var hadith *Hadith
	err := ap.dB.Where("collection = ? AND bookNumber = ? AND babID = ? AND hadithNumber = ?", collection, bookNumber, chapterNumber, hadithNumber).First(&hadith).Error
	if err != nil {
		return nil, err
	}
	return hadith, nil
}

func (ap *ApplicationPersistence) GetHadithByCollectionAndHadithNumber(collection, hadithNumber string) (*Hadith, error) {
	var hadith *Hadith
	err := ap.dB.Where("collection = ? AND hadithNumber = ?", collection, hadithNumber).First(&hadith).Error
	if err != nil {
		return nil, err
	}
	return hadith, nil
}

func (ap *ApplicationPersistence) GetHadithByUrn(urn int) (*Hadith, error) {
	var hadith *Hadith
	err := ap.dB.Where("englishURN = ?", urn).First(&hadith).Error
	if err != nil {
		return nil, err
	}
	return hadith, nil
}

func (ap *ApplicationPersistence) GetRandomHadithInCollection(collection string) (*Hadith, error) {
	var hadith *Hadith
	err := ap.dB.Where("collection = ?", collection).Order(getRandomFunction(ap.dB)).First(&hadith).Error
	if err != nil {
		return nil, err
	}
	return hadith, nil
}
