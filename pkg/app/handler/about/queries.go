package about

import "net/http"

// Queries shows the advanced search queries aboout page
func Queries(w http.ResponseWriter, r *http.Request) {
	renderAboutTemplate(w, r, "queries")
}
