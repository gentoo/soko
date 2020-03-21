package about

import (
	"net/http"
)

// Index shows the landing page of the about pages
func Index(w http.ResponseWriter, r *http.Request) {
	renderAboutTemplate(w, r, "index")
}
