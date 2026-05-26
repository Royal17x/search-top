package api

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/Royal17x/search-top/internal/metrics"
)

const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>search-top</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        *{box-sizing:border-box;margin:0;padding:0}
        body{font-family:'SF Mono','Fira Code',monospace;background:#0f1117;color:#e2e8f0;min-height:100vh;padding:2rem}
        .header{margin-bottom:2rem}
        .header h1{font-size:1.4rem;color:#7c3aed;letter-spacing:-.5px}
        .header p{color:#4a5568;font-size:.75rem;margin-top:.3rem}
        .pulse{display:inline-block;width:6px;height:6px;background:#22c55e;border-radius:50%;margin-right:.4rem;animation:pulse 2s infinite}
        @keyframes pulse{0%,100%{opacity:1}50%{opacity:.2}}
        .card{background:#1e2230;border:1px solid #2d3748;border-radius:8px;padding:1.25rem;margin-bottom:1rem}
        .card-title{font-size:.65rem;text-transform:uppercase;letter-spacing:1px;color:#4a5568;margin-bottom:1rem}
        .query-row{display:flex;align-items:center;gap:.75rem;padding:.45rem 0;border-bottom:1px solid #1a1f2e}
        .query-row:last-child{border-bottom:none}
        .rank{color:#4a5568;font-size:.7rem;width:1.5rem;text-align:right;flex-shrink:0}
        .query-text{flex:1;font-size:.875rem}
        .bar-wrap{width:100px;height:3px;background:#1a1f2e;border-radius:2px;overflow:hidden}
        .bar{height:100%;background:linear-gradient(90deg,#7c3aed,#a78bfa);border-radius:2px;transition:width .4s ease}
        .count{color:#64748b;font-size:.75rem;width:2.5rem;text-align:right;flex-shrink:0}
        .empty{color:#4a5568;font-size:.8rem;text-align:center;padding:1.5rem 0}
        .tags{display:flex;flex-wrap:wrap;gap:.4rem;margin-bottom:.875rem;min-height:1.5rem}
        .tag{background:#2d1f4e;border:1px solid #7c3aed33;color:#a78bfa;padding:.15rem .5rem;border-radius:4px;font-size:.75rem;display:flex;align-items:center;gap:.3rem}
        .tag button{background:none;border:none;color:#7c3aed;cursor:pointer;font-size:.9rem;line-height:1;padding:0}
        .add-form{display:flex;gap:.5rem}
        .add-form input{flex:1;background:#0f1117;border:1px solid #2d3748;color:#e2e8f0;padding:.35rem .7rem;border-radius:4px;font-family:inherit;font-size:.8rem}
        .add-form input:focus{outline:none;border-color:#7c3aed}
        .add-form button{background:#7c3aed;border:none;color:#fff;padding:.35rem .9rem;border-radius:4px;cursor:pointer;font-family:inherit;font-size:.8rem}
        .add-form button:hover{background:#6d28d9}
        .htmx-indicator{opacity:0;transition:opacity .2s}
        .htmx-request .htmx-indicator{opacity:1}
    </style>
</head>
<body>
    <div class="header">
        <h1>search-top</h1>
        <p><span class="pulse"></span>trending queries &middot; last 5 min &middot; live</p>
    </div>

    <div class="card">
        <div class="card-title">top queries</div>
        <div hx-get="/partial/top"
             hx-trigger="load, every 2s"
             hx-swap="innerHTML">
        </div>
    </div>

    <div class="card">
        <div class="card-title">stop list</div>
        <div id="sl"
             hx-get="/partial/stoplist"
             hx-trigger="load"
             hx-swap="innerHTML">
        </div>
    </div>
</body>
</html>`

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func (s *Server) handlePartialTop(w http.ResponseWriter, _ *http.Request) {
	items := s.cache.Get()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if len(items) == 0 {
		fmt.Fprint(w, `<div class="empty">no data · run seed to populate</div>`)
		return
	}

	maxCount := items[0].Count
	var sb strings.Builder
	for i, item := range items {
		pct := 0
		if maxCount > 0 {
			pct = int(float64(item.Count) / float64(maxCount) * 100)
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="query-row"><span class="rank">#%d</span>`+
				`<span class="query-text">%s</span>`+
				`<div class="bar-wrap"><div class="bar" style="width:%d%%"></div></div>`+
				`<span class="count">%d</span></div>`,
			i+1, html.EscapeString(item.Query), pct, item.Count,
		))
	}
	fmt.Fprint(w, sb.String())
}

func (s *Server) handlePartialStoplist(w http.ResponseWriter, _ *http.Request) {
	renderStoplist(w, s)
}

func (s *Server) handlePartialStoplistAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		if word := strings.TrimSpace(r.FormValue("word")); word != "" {
			s.stoplist.Add(word)
			metrics.StoplistSize.Set(float64(len(s.stoplist.AllWords())))
		}
	}
	renderStoplist(w, s)
}

func (s *Server) handlePartialStoplistRemove(w http.ResponseWriter, r *http.Request) {
	s.stoplist.Remove(r.PathValue("word"))
	metrics.StoplistSize.Set(float64(len(s.stoplist.AllWords())))
	renderStoplist(w, s)
}

func renderStoplist(w http.ResponseWriter, s *Server) {
	words := s.stoplist.AllWords()
	sort.Strings(words)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var sb strings.Builder
	sb.WriteString(`<div class="tags">`)
	if len(words) == 0 {
		sb.WriteString(`<span style="color:#4a5568;font-size:.75rem">empty</span>`)
	}
	for _, word := range words {
		sb.WriteString(fmt.Sprintf(
			`<span class="tag">%s`+
				`<button hx-delete="/partial/stoplist/%s"`+
				` hx-target="#sl" hx-swap="innerHTML">×</button></span>`,
			html.EscapeString(word),
			url.PathEscape(word),
		))
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<form class="add-form"` +
		` hx-post="/partial/stoplist"` +
		` hx-target="#sl"` +
		` hx-swap="innerHTML"` +
		` hx-on::after-request="this.reset()">` +
		`<input type="text" name="word" placeholder="block a word..." autocomplete="off">` +
		`<button type="submit">block</button></form>`)

	fmt.Fprint(w, sb.String())
}
