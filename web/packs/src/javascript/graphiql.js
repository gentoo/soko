function graphQLFetcher(graphQLParams) {
    return fetch(
        window.graphqlEndpoint,
        {
            method: 'post',
            headers: {
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(graphQLParams),
            credentials: 'omit',
        },
    ).then(function (response) {
        return response.json().catch(function () {
            return response.text();
        });
    });
}


ReactDOM.render(
    React.createElement(GraphiQL, {
        fetcher: graphQLFetcher,
        defaultVariableEditorOpen: false,
        defaultQuery: '#\n' +
            '# Welcome to the packages.gentoo.org GraphQL API Explorer\n' +
            '#\n' +
            '# Powered by GraphiQL, an in-browser tool for exploring GraphQL APIs, as well as \n' +
            '# writing, validating, and testing GraphQL queries.\n' +
            '#\n' +
            '#\n' +
            '# Please click on "Examples & History" above to view some exemplary queries,\n' +
            '# or run the below query to get started.\n' +
            '#\n' +
            '{\n' +
            '  packages(Name: "gentoo-sources"){\n' +
            '    Atom,\n' +
            '    Maintainers {\n' +
            '      Name\n' +
            '    }\n' +
            '  }\n' +
            '}',
    }),
    document.getElementById('graphiql'),
);

/*
 * Add gentoo Logo
 */
var gentooLogo= "<img src=\"https://www.gentoo.org/assets/img/logo/gentoo-signet.svg\" style=\"height: 30px;vertical-align:middle;\">";
document.getElementsByClassName("title")[0].innerHTML = gentooLogo;

/*
 * Add examples
 */

document.getElementsByClassName("history-title")[0].innerHTML  = "Examples & History";
document.getElementsByClassName("toolbar-button")[3].innerHTML = "Examples & History";

var favorites = window.localStorage.getItem("graphiql:favorites");

if(favorites == null){
    window.localStorage.setItem("graphiql:favorites", '{"favorites":[{"query":"\\n{\\n  packages(Name: \\"gentoo-sources\\"){\\n    Atom,\\n    Maintainers {\\n      Name\\n    }\\n  }\\n}","variables":null,"label":"Search Packages By Name","favorite":true}]}');
    window.localStorage.setItem("graphiql:queries", '{"queries":[]}');
    window.localStorage.setItem("graphiql:query", '## Welcome to the packages.gentoo.org GraphQL API Explorer## Powered by GraphiQL, an in-browser tool for exploring GraphQL APIs, as well as # writing, validating, and testing GraphQL queries.### Please click on "Examples & History" above to view some exemplary queries,# or run the below query to get started.#{  packages(Name: "gentoo-sources"){    Atom,    Maintainers {      Name    }  }}');
    window.localStorage.setItem("graphiql:editorFlex", '1');
    window.localStorage.setItem("graphiql:docExplorerWidth", '350');
    window.localStorage.setItem("graphiql:variableEditorHeight", '200');
    window.localStorage.setItem("graphiql:historyPaneOpen", 'true');
    location.reload();
}
