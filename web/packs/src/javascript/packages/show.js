$(function() {
  var atom = $('#package-title').data('atom');

  $.ajax({
    url: '/packages/' + atom + '/changelog.html'
  }).done(function(data) {
    $('#changelog-container').html(data);
    $(document).trigger('kkuleomi:ajax');
  }).fail(function() {
    $('#changelog-container > li').html('<span class="fa fa-fw fa-3x fa-ban text-danger"></span><br><br>Changelog currently not available. Please check back later.');
  });
});


window.devMode = false;

$( "#activate-dev-mode" ).removeClass("d-none");

$( "#activate-dev-mode" ).click(function() {
  if(window.devMode){
    var cols = document.getElementsByClassName('dev-mode');
    for(i = 0; i < cols.length; i++) {
      cols[i].classList.add("d-none");
    }
    window.devMode = false;
  }else{
    var cols = document.getElementsByClassName('dev-mode');
    for(i = 0; i < cols.length; i++) {
      cols[i].classList.remove("d-none");
    }
    window.devMode = true;
  }
});
