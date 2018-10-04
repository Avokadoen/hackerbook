var adminAccess = false;

function isAdmin(){
    var req = new XMLHttpRequest();
    req.open("POST", window.location.origin + "/verifyadmin", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send();

    req.onload = function() {
        answer = this.responseText;
        if(answer === "Admin granted" && this.status === 200){
            adminAccess = true;
            $("div.admin").show();
           }
    }
}

function createNewCategory(event){
    var req = new XMLHttpRequest();
    req.open("POST", "/admincreatenewcategory", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        category:      event.category.value
    }));


    req.onload = function() {
        answer = this.responseText;
        if(answer === "Category inserted") {
            location.reload(true);
        }
        else {
            document.getElementById("AdminErrorMessage").innerHTML = "Was not able to create new category";
        }
    }
}