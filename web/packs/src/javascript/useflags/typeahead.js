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
        template: '<span>{{name}}</span> <span class="kk-suggest-detail">{{description}}</span>'
      }
    },
    callback: {
      onClick: function(node, a, item, event) {
        window.location = item.href;
      }
    }
  });
});
