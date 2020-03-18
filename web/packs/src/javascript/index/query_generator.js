function buildAdvancedQuery(){
    var query = ""
    document.querySelectorAll('#search-container > .row').forEach(function(element) {
        var term = element.querySelector('.form-control').value;

        if(!term.replace(/\s/g, '').length){
            return;
        }else{
            term = parseSearchTerm(term);
        }

        var operator = parseOperator(element.querySelector(".pgo-query-operator").value);
        var field = element.querySelector('.pgo-query-field').value;

        query += operator + field + ":" + term + " ";
    });
    document.getElementById('q').value = query;
}

function parseOperator(operator){
    switch(operator) {
        case "should match":
            return "";
        case "must match":
            return "+";
        case "must not match":
            return "-";
        default:
            return "";
    }
}

function parseSearchTerm(term){
    if (/\s/.test(term) && !/^\".*\"$/.test(term)) {
        return "\"" + term + "\""
    }else{
        return term
    }
}

function addInput(self){
    var new_input = document.querySelector('#search-container > .row').cloneNode(true);
    setEventListener(new_input);
    resetInput(new_input);
    document.querySelector('#search-container').append(new_input);
    checkDeleteButtons();
    checkAddButtons();
}

function resetInput(input) {
    input.querySelector('.form-control').value = '';
    input.querySelector('.pgo-query-operator').value = 'should match';
    input.querySelector('.pgo-query-field').value = 'name';
}

function deleteInput(self){
    getThirdParent(self).removeChild(getSecondParent(self));
    checkDeleteButtons();
    checkAddButtons();
}

function checkDeleteButtons(){
    if(document.querySelectorAll('#search-container > .row').length == 1){
        document.querySelectorAll('.pgo-query-delete-btn').forEach(function(element) {
            element.style.display = 'none';
        });
    }else{
        document.querySelectorAll('.pgo-query-delete-btn').forEach(function(element) {
            element.style.display = 'block';
        });
    }
}

function checkAddButtons(){
    document.querySelectorAll('.pgo-query-add-btn').forEach(function(element) {
        element.style.display = 'none';
    });

    document.querySelectorAll('.pgo-query-add-btn')[document.querySelectorAll('.pgo-query-add-btn').length - 1].style.display = 'block';
}

function getThirdParent(self) {
    return self.parentElement.parentElement.parentElement;
}

function getSecondParent(self) {
    return self.parentElement.parentElement;
}

function setEventListener(element){
    element.querySelector(".pgo-query-add-btn").addEventListener("click", addInput);
    element.querySelector(".pgo-query-delete-btn").addEventListener("click", function(){ deleteInput(this); });
}

checkDeleteButtons();

setEventListener(document);

document.getElementById("buildAdvancedQuery").addEventListener("click", function(){ buildAdvancedQuery(); });