$(document).on('ready page:load kkuleomi:ajax', function(event) {
  $('[data-toggle="tooltip"]').tooltip();

  $('.kk-i18n-date').each(function(idx) {
    // TODO: Support different date formats
    var me = $(this);
    me.text(moment.unix(me.data('utcts')).local().format('ddd, D MMM YYYY HH:mm'));
  });
});
