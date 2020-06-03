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