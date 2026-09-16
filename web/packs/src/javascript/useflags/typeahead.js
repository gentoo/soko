import { escapeHtml, highlightHtml } from '../html';

$(function() {
  $('#q').typeahead({
    order: "asc",
    dynamic: true,
    source: {
      use: {
        display: 'name',
        href: function(item) { return '/useflags/' + item.name; },
        url: [{
          type: 'GET',
          url: "/useflags/suggest.json",
          data: {
            q: "{{query}}"
          }
        }, 'results'],
        template: function (query, item) {
          return '<span>' + highlightHtml(item.name, query) + '</span> ' +
            '<span class="kk-suggest-detail">' + escapeHtml(item.description) + '</span>';
        }
      }
    },
    callback: {
      onClick: function(node, a, item, event) {
        window.location = item.href;
      }
    }
  });
});
