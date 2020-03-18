$(function() {
  $('#q').typeahead({
    order: 'asc',
    dynamic: true,
    delay: 500,
    source: {
      packages: {
        display: 'name',
        href: function(item) { return '/packages/' + item.category + '/' + item.name; },
        url: [{
          type: 'GET',
          url: "/packages/suggest.json",
          data: {
            q: "{{query}}"
          }
        }, 'results'],
        template: '<span class="kk-suggest-cat">{{category}}</span>/<span class="kk-suggest-pkg">{{name}}</span> <span class="kk-suggest-detail">{{description}}</span>'
      }
    },
    callback: {
      onClick: function(node, a, item, event) {
        window.location = item.href;
      }
    }
  });
});
