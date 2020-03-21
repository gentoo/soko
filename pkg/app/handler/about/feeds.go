package about

import "net/http"

// Feeds shows the feeds about page
func Feeds(w http.ResponseWriter, r *http.Request) {
	renderAboutTemplate(w, r, "feeds")
}
