import { escapeHtml, highlightHtml } from '../html';

$(function () {
  $('#q').typeahead({
    order: 'asc',
    dynamic: true,
    delay: 500,
    source: {
      packages: {
        display: 'name',
        href: function (item) { return '/packages/' + item.category + '/' + item.name; },
        url: [{
          type: 'GET',
          url: "/packages/suggest.json",
          data: {
            q: "{{query}}"
          }
        }, 'results'],
        template: function (query, item) {
          return '<span class="kk-suggest-cat">' + escapeHtml(item.category) + '</span>/' +
            '<span class="kk-suggest-pkg">' + highlightHtml(item.name, query) + '</span> ' +
            '<span class="kk-suggest-detail">' + escapeHtml(item.description) + '</span>';
        }
      }
    },
    callback: {
      onClick: function (node, a, item, event) {
        window.location = item.href;
      }
    }
  });
});
