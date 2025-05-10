package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var DEFAULT_PORT = "7878"
var PER_PAGE = 50
var MAX_USERS = 10

func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}
	return port
}

func startServer(mux *http.ServeMux) {
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	addr := "0.0.0.0:" + getServerPort()
	log.Printf("listening on %s...\n", addr)

	if certFile != "" && keyFile != "" {
		http.ListenAndServeTLS(addr, certFile, keyFile, mux)
	} else {
		http.ListenAndServe(addr, mux)
	}
}

type Queries struct {
	UserSources        *sql.Stmt
	AllArticles        *sql.Stmt
	SearchUserArticles *sql.Stmt
	SourceArticles     *sql.Stmt
}

func readStatement(db *sql.DB, path string) *sql.Stmt {
	qs, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare(string(qs))
	if err != nil {
		log.Fatal(err)
	}
	return stmt

}

func prepareQueries(db *sql.DB) Queries {
	queries := Queries{}
	queries.AllArticles = readStatement(db, "./queries/all_articles_paginate.sql")
	queries.SourceArticles = readStatement(db, "./queries/source_articles_paginate.sql")
	queries.SearchUserArticles = readStatement(db, "./queries/search_user_articles_paginate.sql")
	queries.UserSources = readStatement(db, "./queries/user_sources.sql")
	return queries
}

func main() {

	db, err := sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	qs := prepareQueries(db)

	mux := http.NewServeMux()

	searchUserArticlesHandler := WrappedHandler(db, &qs, SearchUserArticlesResponseHandler)
	sourceArticlesHandler := WrappedHandler(db, &qs, SourceArticlesResponseHandler)
	sourcesHandler := WrappedHandler(db, &qs, SourcesResponseHandler)

	mux.Handle("/user-sources/{user}/", sourcesHandler)
	mux.Handle("/user-sources/{user}", sourcesHandler)

	mux.Handle("/source-articles-paged/{source}/{page}/", sourceArticlesHandler)
	mux.Handle("/source-articles-paged/{source}/{page}", sourceArticlesHandler)

	mux.Handle("/search-user-articles-paged/{user}/{search}/{page}/", searchUserArticlesHandler)
	mux.Handle("/search-user-articles-paged/{user}/{search}/{page}", searchUserArticlesHandler)

	startServer(mux)
}

type Article struct {
	URI     string `json:"uri"`
	Title   string `json:"title"`
	PubDate int64  `json:"pubdate"`
}

type Source struct {
	URI   string `json:"uri"`
	Title string `json:"title"`
	ID    int64  `json:"id"`
}

type ArticlesResponse struct {
	Results []Article `json:"results"`
}

type SourcesResponse struct {
	Results []Source `json:"results"`
}

func parseReqInt(value string, fallback_value string, fallback_result int, min_result int, max_result int) int {
	if fallback_value == "" {
		fallback_value = "0"
	}

	r := fallback_result

	if value == "" {
		value = fallback_value
	}

	r, err := strconv.Atoi(value)
	if err != nil {
		r = fallback_result
		log.Printf("Number conversion failed for %v. Fallback: %v", value, r)
	}

	if r < min_result {
		r = min_result
	}

	if max_result > 0 {
		if r > max_result {
			r = max_result
		}
	}
	return r
}

func parseReqPage(value string) int {
	page := parseReqInt(value, "1", 1, 0, 10000)
	if page > 0 {
		page = page - 1
	}
	return page
}

type DBResponseHandler func(*sql.DB, *Queries, http.ResponseWriter, *http.Request)

func WrappedHandler(db *sql.DB, qs *Queries, f DBResponseHandler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		f(db, qs, w, r)
	}
	return http.HandlerFunc(fn)
}

func SourcesResponseHandler(db *sql.DB, qs *Queries, w http.ResponseWriter, r *http.Request) {
	userId := parseReqInt(r.PathValue("user"), "1", 1, 1, MAX_USERS)
	rows, err := qs.UserSources.Query(userId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var sources []Source
	for rows.Next() {
		var source Source

		err = rows.Scan(&source.URI, &source.Title, &source.ID)
		if err != nil {
			log.Fatal(err)
		}
		sources = append(sources, source)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-type", "application/json")

	response := SourcesResponse{Results: sources}

	enc := json.NewEncoder(w)
	err = enc.Encode(&response)
	if err != nil {
		log.Fatal(err)
	}
}

func readArticles(rows *sql.Rows) []Article {
	articles := make([]Article, 0)
	for rows.Next() {
		var article Article

		err := rows.Scan(&article.URI, &article.Title, &article.PubDate)
		if err != nil {
			log.Fatal(err)
		}
		articles = append(articles, article)
	}
	err := rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return articles
}

func SearchUserArticlesResponseHandler(db *sql.DB, qs *Queries, w http.ResponseWriter, r *http.Request) {
	user := parseReqInt(r.PathValue("user"), "0", 0, 0, 1000)
	search := r.PathValue("search")
	page := parseReqPage(r.PathValue("page"))

	rows, err := qs.SearchUserArticles.Query(search, user, page*PER_PAGE, PER_PAGE)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	articles := readArticles(rows)

	w.Header().Set("Content-type", "application/json")

	response := ArticlesResponse{Results: articles}

	enc := json.NewEncoder(w)
	err = enc.Encode(&response)
	if err != nil {
		log.Fatal(err)
	}
}

func SourceArticlesResponseHandler(db *sql.DB, qs *Queries, w http.ResponseWriter, r *http.Request) {
	source := parseReqInt(r.PathValue("source"), "0", 0, 0, 1000)
	page := parseReqPage(r.PathValue("page"))

	rows, err := qs.SourceArticles.Query(source, page*PER_PAGE, PER_PAGE)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	articles := readArticles(rows)

	w.Header().Set("Content-type", "application/json")

	response := ArticlesResponse{Results: articles}

	enc := json.NewEncoder(w)
	err = enc.Encode(&response)
	if err != nil {
		log.Fatal(err)
	}
}
