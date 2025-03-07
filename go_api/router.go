package go_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Suhaibinator/api/go_service"
	"github.com/gorilla/mux"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error struct {
		Details string `json:"details"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// writeJSONError writes a structured error response
func writeJSONError(w http.ResponseWriter, code int, details string) {
	var errorResponse ErrorResponse
	errorResponse.Error.Code = code
	errorResponse.Error.Details = details

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse)
}

type ApplicationRouter struct {
	router             *mux.Router
	applicationService *go_service.ApplicationService
}

func NewApplicationRouter(applicationService *go_service.ApplicationService) *ApplicationRouter {
	router := mux.NewRouter()
	return &ApplicationRouter{router: router, applicationService: applicationService}
}

func (ar *ApplicationRouter) RegisterRoutes() {
	router := ar.router

	// Add middleware for authentication
	//router.Use(ar.authMiddleware)

	// Home route
	router.HandleFunc("/", ar.homeHandler).Methods("GET")

	// Register v1 routes
	ar.registerV1Routes()
}

// authMiddleware checks for the x-aws-secret header
func (ar *ApplicationRouter) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication in debug mode or for the home route
		if r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the AWS secret from the environment
		awsSecret := os.Getenv("AWS_SECRET")
		if awsSecret == "" {
			// If AWS_SECRET is not set, assume we're in debug mode and skip authentication
			next.ServeHTTP(w, r)
			return
		}

		// Check the x-aws-secret header
		if r.Header.Get("x-aws-secret") != awsSecret {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid or missing API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ar *ApplicationRouter) registerV1Routes() {
	// Create a subrouter for v1 routes
	v1Router := ar.router.PathPrefix("/v1").Subrouter()

	// Create a subrouter for collections routes
	collectionsRouter := v1Router.PathPrefix("/collections").Subrouter()

	// Collections routes
	collectionsRouter.HandleFunc("", ar.apiGetAllCollections).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}", ar.apiCollectionHandler).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/books", ar.apiGetBooksInCollectionHandler).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/books/{bookNumber}", ar.apGetBookHandler).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/books/{bookNumber}/chapters", ar.apiGetChaptersInBookInCollection).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/books/{bookNumber}/chapters/{chapterId}", ar.apiGetChapterInBookInCollection).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/books/{bookNumber}/hadiths", ar.apiGetHadithsInBook).Methods("GET")
	collectionsRouter.HandleFunc("/{collectionName}/hadiths/{hadithNumber}", ar.apiGetHadithInCollectionByHadithNumber).Methods("GET")

	// Other v1 routes
	v1Router.HandleFunc("/hadiths/{urn:[0-9]+}", ar.apiGetHadithByUrn).Methods("GET")
	// v1Router.HandleFunc("/hadiths", ar.apiGetHadithsByCollectionAndBookAndChapter).Methods("GET")
	v1Router.HandleFunc("/hadiths/random", ar.apiHadithsRandomHandler).Methods("GET")
}

func (ar *ApplicationRouter) Run(port int) {
	log.Printf("Starting server on port %v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), ar.router)
}

func (ar *ApplicationRouter) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Welcome to sunnah.com API.</h1>"))
}

// TODO Paginated responses

func (ar *ApplicationRouter) apiGetAllCollections(w http.ResponseWriter, r *http.Request) {
	page, limit, paginationParameterRetrievalErr := getPaginationParameters(r)
	if paginationParameterRetrievalErr != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid pagination parameters: "+paginationParameterRetrievalErr.Error())
		return
	}

	// Get the collections from the service
	collections, total, err := ar.applicationService.GetPaginatedHadithCollections(page, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve collections: "+err.Error())
		return
	}

	// Convert the collections to API collections
	apiCollections := make([]*Collection, len(collections))
	for i, collection := range collections {
		apiCollections[i] = ConvertDbCollectionToApiCollection(collection)
	}

	// Convert to interface slice for PaginatedResponse
	data := make([]interface{}, len(apiCollections))
	for i, collection := range apiCollections {
		data[i] = collection
	}

	// Create paginated response
	response := NewPaginatedResponse(data, int(total), limit, page)

	result, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiCollectionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collectionName"]

	// Get the collection from the service
	collection, err := ar.applicationService.GetHadithCollectionByName(collectionName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve collection: "+err.Error())
		return
	}

	// Convert the collection to an API collection
	apiCollection := ConvertDbCollectionToApiCollection(collection)
	result, jsonErr := json.Marshal(apiCollection)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetBooksInCollectionHandler(w http.ResponseWriter, r *http.Request) {
	page, limit, paginationParameterRetrievalErr := getPaginationParameters(r)
	if paginationParameterRetrievalErr != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid pagination parameters: "+paginationParameterRetrievalErr.Error())
		return
	}

	collectionName := mux.Vars(r)["collectionName"]

	// Get the books from the service
	books, total, err := ar.applicationService.GetPaginatedBooksByCollection(collectionName, page, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve books: "+err.Error())
		return
	}

	// Convert the books to API books
	apiBooks := make([]*Book, len(books))
	for i, book := range books {
		apiBooks[i] = ConvertDbBookToApiBook(book)
	}

	// Convert to interface slice for PaginatedResponse
	data := make([]interface{}, len(apiBooks))
	for i, book := range apiBooks {
		data[i] = book
	}

	// Create paginated response
	response := NewPaginatedResponse(data, int(total), limit, page)

	result, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apGetBookHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := mux.Vars(r)["collectionName"]
	bookNumber := mux.Vars(r)["bookNumber"]

	// Get the book from the service
	book, err := ar.applicationService.GetBookByCollectionAndBookNumber(collectionName, bookNumber)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve book: "+err.Error())
		return
	}

	// Convert the book to an API book
	apiBook := ConvertDbBookToApiBook(book)
	result, jsonErr := json.Marshal(apiBook)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetChaptersInBookInCollection(w http.ResponseWriter, r *http.Request) {
	page, limit, paginationParameterRetrievalErr := getPaginationParameters(r)
	if paginationParameterRetrievalErr != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid pagination parameters: "+paginationParameterRetrievalErr.Error())
		return
	}

	vars := mux.Vars(r)
	collectionName := vars["collectionName"]
	bookNumber := vars["bookNumber"]

	// Get the chapters from the service
	chapters, total, err := ar.applicationService.GetPaginatedChaptersByCollectionAndBookNumber(collectionName, bookNumber, page, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve chapters: "+err.Error())
		return
	}

	// Convert the chapters to API chapters
	apiChapters := make([]*Chapter, len(chapters))
	for i, chapter := range chapters {
		apiChapters[i] = ConvertDbChapterToApiChapter(chapter)
	}

	// Convert to interface slice for PaginatedResponse
	data := make([]interface{}, len(apiChapters))
	for i, chapter := range apiChapters {
		data[i] = chapter
	}

	// Create paginated response
	response := NewPaginatedResponse(data, int(total), limit, page)

	result, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetChapterInBookInCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := mux.Vars(r)["collectionName"]
	bookNumber := mux.Vars(r)["bookNumber"]
	chapterId := mux.Vars(r)["chapterId"]

	// Get the chapter from the service
	chapter, err := ar.applicationService.GetChapterByCollectionAndBookNumberAndChapterNumber(collectionName, bookNumber, chapterId)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve chapter: "+err.Error())
		return
	}

	// Convert the chapter to an API chapter
	apiChapter := ConvertDbChapterToApiChapter(chapter)
	result, jsonErr := json.Marshal(apiChapter)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetHadithsInBook(w http.ResponseWriter, r *http.Request) {
	page, limit, paginationParameterRetrievalErr := getPaginationParameters(r)
	if paginationParameterRetrievalErr != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid pagination parameters: "+paginationParameterRetrievalErr.Error())
		return
	}

	vars := mux.Vars(r)
	collectionName := vars["collectionName"]
	bookNumber := vars["bookNumber"]

	// Get the hadiths from the service
	hadiths, total, err := ar.applicationService.GetPaginatedHadithsByCollectionAndBookNumber(collectionName, bookNumber, page, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve hadiths: "+err.Error())
		return
	}

	// Convert the hadiths to API hadiths
	apiHadiths := make([]*Hadith, len(hadiths))
	for i, hadith := range hadiths {
		apiHadiths[i] = ConvertDbHadithToApiHadith(hadith)
	}

	// Convert to interface slice for PaginatedResponse
	data := make([]interface{}, len(apiHadiths))
	for i, hadith := range apiHadiths {
		data[i] = hadith
	}

	// Create paginated response
	response := NewPaginatedResponse(data, int(total), limit, page)

	result, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetHadithInCollectionByHadithNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collectionName"]
	hadithNumber := vars["hadithNumber"]

	// Get the hadith from the service
	hadith, err := ar.applicationService.GetHadithByCollectionAndHadithNumber(collectionName, hadithNumber)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve hadith: "+err.Error())
		return
	}

	// Convert the hadith to an API hadith
	apiHadith := ConvertDbHadithToApiHadith(hadith)
	result, jsonErr := json.Marshal(apiHadith)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiGetHadithByUrn(w http.ResponseWriter, r *http.Request) {
	urn, err := strconv.Atoi(mux.Vars(r)["urn"])
	if err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid URN parameter: "+err.Error())
		return
	}

	// Get the hadith from the service
	hadith, err := ar.applicationService.GetHadithByUrn(urn)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve hadith: "+err.Error())
		return
	}

	// Convert the hadith to an API hadith
	apiHadith := ConvertDbHadithToApiHadith(hadith)
	result, jsonErr := json.Marshal(apiHadith)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (ar *ApplicationRouter) apiHadithsRandomHandler(w http.ResponseWriter, r *http.Request) {
	// Get the random hadith from the service
	hadith, err := ar.applicationService.GetRandomHadith()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve random hadith: "+err.Error())
		return
	}

	// Convert the hadith to an API hadith
	apiHadith := ConvertDbHadithToApiHadith(hadith)
	result, jsonErr := json.Marshal(apiHadith)
	if jsonErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to marshal JSON response: "+jsonErr.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
