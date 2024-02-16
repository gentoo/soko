package categories

import (
	"net/http"
	"soko/pkg/app/layout"

	"github.com/a-h/templ"
)

var categoriesViewTabs = []layout.SubTab{
	{
		Name: "Categories",
		Link: "/categories",
		Icon: "fa fa-list-ul mr-1",
	},
	{
		Name: "Added",
		Link: "/packages/added",
		Icon: "fa fa-history mr-1",
	},
	{
		Name: "Updated",
		Link: "/packages/updated",
		Icon: "fa fa-asterisk mr-1",
	},
	{
		Name: "Newly Stable",
		Link: "/packages/stable",
		Icon: "fa fa-check-circle-o mr-1",
	},
	{
		Name: "Keyworded",
		Link: "/packages/keyworded",
		Icon: "fa fa-circle-o mr-1",
	},
}

func RenderPage(w http.ResponseWriter, r *http.Request, title string, currentTab string, content templ.Component) {
	layout.TabbedLayout(title, "packages", "Packages", "fa fa-fw fa-cubes", "", categoriesViewTabs,
		currentTab, content).Render(r.Context(), w)
}
