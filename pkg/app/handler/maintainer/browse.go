package maintainer

import (
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"sort"
	"strings"
)

// Show renders a template to show a given maintainer page
func Browse(w http.ResponseWriter, r *http.Request) {

	tabName := "projects"
	if strings.HasSuffix(r.URL.Path, "/gentoo-projects") {
		tabName = "projects"
	} else if strings.HasSuffix(r.URL.Path, "/gentoo-developers") {
		tabName = "gentoo-developers"
	} else if strings.HasSuffix(r.URL.Path, "/proxied-maintainers") {
		tabName = "proxied-maintainers"
	}

	var maintainers []*models.Maintainer
	database.DBCon.Model(&maintainers).Select()

	sort.Slice(maintainers, func(i, j int) bool {
		return maintainers[i].Name < maintainers[j].Name
	})

	renderBrowseTemplate(createBrowseData(tabName, maintainers), w)

}
