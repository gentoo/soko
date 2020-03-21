package about

import "net/http"

// Feedback shows the feedback about page
func Feedback(w http.ResponseWriter, r *http.Request) {
	renderAboutTemplate(w, r, "feedback")
}
