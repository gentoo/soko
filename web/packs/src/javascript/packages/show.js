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


var devMode = window.localStorage.getItem('devMode');

if(devMode == null){
  devMode = 'false';
  window.localStorage.setItem('devMode', devMode);
}

// update current page
if(devMode == 'true'){
  var cols = document.getElementsByClassName('dev-mode');
  for(i = 0; i < cols.length; i++) {
    cols[i].classList.remove("d-none");
  }
  document.getElementById("activate-dev-mode").innerHTML = "Switch to Standard Mode";
}else{
  var cols = document.getElementsByClassName('dev-mode');
  for(i = 0; i < cols.length; i++) {
    cols[i].classList.add("d-none");
  }
  document.getElementById("activate-dev-mode").innerHTML = "Switch to Developer Mode";
}

// register listener for future changes
$( "#activate-dev-mode" ).click(function() {
  if(devMode == 'true'){
    var cols = document.getElementsByClassName('dev-mode');
    for(i = 0; i < cols.length; i++) {
      cols[i].classList.add("d-none");
    }
    devMode = 'false';
    window.localStorage.setItem('devMode', devMode);
    document.getElementById("activate-dev-mode").innerHTML = "Switch to Developer Mode";
  }else{
    var cols = document.getElementsByClassName('dev-mode');
    for(i = 0; i < cols.length; i++) {
      cols[i].classList.remove("d-none");
    }
    devMode = 'true';
    window.localStorage.setItem('devMode', devMode);
    document.getElementById("activate-dev-mode").innerHTML = "Switch to Standard Mode";
  }
});
