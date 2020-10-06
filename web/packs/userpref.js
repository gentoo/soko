window.onload = function() {
    document.body.onscroll = function(){
        updateNav();
    };
};

document.addEventListener("turbolinks:load", function() {

    updateNav();
})

function updateNav() {
    var elements = [];

    if(window.location.href.includes("/user/preferences/packages")){
        elements = ['overview', 'dependencies', 'qa-report', 'pull-requests', 'bugs', 'security', 'changelog', 'tabs'];
    }else if(window.location.href.includes("/user/preferences/arches")){
        elements = ['keywords', 'defaults'];
    }

    for(var i = 0; i < elements.length; i++){
        if (document.getElementById(elements[i]).getBoundingClientRect().y <= window.innerHeight) {
            document.getElementById(elements[i]+"-tab").classList.add("active");
        } else {
            document.getElementById(elements[i]+"-tab").classList.remove("active");
        }
    }
}

if(document.getElementById("myModal") != null) {
    var modal = document.getElementById("myModal");

    var img1 = document.getElementById("img1");
    var img2 = document.getElementById("img2");
    var modalImg = document.getElementById("img01");
    var captionText = document.getElementById("caption");
    img1.onclick = function () {
        modal.style.display = "block";
        modalImg.src = this.src;
        captionText.innerHTML = this.alt;
    }
    img2.onclick = function () {
        modal.style.display = "block";
        modalImg.src = this.src;
        captionText.innerHTML = this.alt;
    }

    var span = document.getElementsByClassName("close")[0];

    span.onclick = function () {
        modal.style.display = "none";
    }
}