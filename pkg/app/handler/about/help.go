package about

import "net/http"

// Help shows the help about page
func Help(w http.ResponseWriter, r *http.Request) {
	renderAboutTemplate(w, r, "help")
}
