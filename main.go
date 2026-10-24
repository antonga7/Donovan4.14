package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"learnDonovan/chapter4/ex4.14/github"
)

type IssueCache struct {
	mu    sync.RWMutex
	items []*github.Issue
}

var globalCache = &IssueCache{}

var issueList = template.Must(template.New("issueList").Parse(`
<h1>Found {{.TotalCount}} issues</h1>
<table border ="1" cellpadding="5">
	<tr>
		<th>Номер</th>
		<th>Автор</th>
		<th>Milestone</th>
		<th>Заголовок</th>
	</tr>
	{{range .Items}}
	<tr>
		<td><a href="/issue?number={{.Number}}">#{{.Number}}</a></td>
		<td><a href="/search?q={{.User.Login}}">{{.User.Login}}</a></td>
		<td>{{if .Milestone}}<a href="/search?q={{.Milestone.Title}}">{{.Milestone.Title}} (#{{.Milestone.Number}})</a>{{end}}</td>
		<td>{{.Title}}</td>
	</tr>
	{{end}}
</table>
`))

var issueView = template.Must(template.New("issueView").Parse(`
<p><a href="/search">← Назад к списку задач</a></p>
<h1>Задача #{{.Number}}: {{.Title}}</h1>
<p><b>Автор:</b> {{.User.Login}}</p>
{{if .Milestone}}
	<p><b>Веха (Milestone):</b> {{.Milestone.Title}} (#{{.Milestone.Number}})</p>
{{end}}
<hr>
<h3>Описание проблемы:</h3>
<pre style="white-space: pre-wrap;">{{.Body}}</pre>
`))

func loadCache() {
	fmt.Println("loading cache")
	var wg sync.WaitGroup

	terms := []string{"is:open"}

	for page := 1; page <= 5; page++ {
		wg.Add(1)

		go func(p int) {
			defer wg.Done()

			result, err := github.SearchIssues(terms, p)
			if err != nil {
				log.Printf("Ошибка загрузки страницы %d: %v\n", p, err)
				return
			}
			globalCache.mu.Lock()
			globalCache.items = append(globalCache.items, result.Items...)
			globalCache.mu.Unlock()

			fmt.Printf("Страница %d успешно загружена (задач: %d)\n", p, len(result.Items))
		}(page)
	}

	wg.Wait()
	fmt.Printf("Кэш готов! Всего загружено задач: %d\n", len(globalCache.items))
}

func main() {
	loadCache()

	http.HandleFunc("/search", handler)
	http.HandleFunc("/issue", issueHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func issueHandler(w http.ResponseWriter, r *http.Request) {
	numstr := r.URL.Query().Get("number")
	num, err := strconv.Atoi(numstr)
	if err != nil {
		http.Error(w, "Некорректный номер задачи", http.StatusBadRequest)
		return
	}

	var found *github.Issue

	globalCache.mu.RLock()
	for _, item := range globalCache.items {
		if item.Number == num {
			found = item
			break
		}
	}
	globalCache.mu.RUnlock()

	if found == nil {
		http.Error(w, "Задача не найдена в кэше", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := issueView.Execute(w, found); err != nil {
		log.Println("Ошибка рендеринга задачи:", err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	fmt.Println("Ищем в кэше", q)

	var matched []*github.Issue

	globalCache.mu.RLock()
	for _, item := range globalCache.items {
		if q == "" ||
			strings.Contains(strings.ToLower(item.Title), q) ||
			strings.Contains(strings.ToLower(item.User.Login), q) ||
			(item.Milestone != nil && strings.Contains(strings.ToLower(item.Milestone.Title), q)) {
			matched = append(matched, item)
		}
	}
	globalCache.mu.RUnlock()

	result := github.IssuesSearchResult{
		TotalCount: len(matched),
		Items:      matched,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := issueList.Execute(w, result); err != nil {
		log.Println("Ошибка рендеринга шаблона: ", err)
	}

}
