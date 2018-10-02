var adminAccess = false;

function isAdmin(){
    var req = new XMLHttpRequest();
    if(isLoggedIn()){
        req.open("POST", window.location.origin + "/validateadmin", true);
        req.setRequestHeader('Content-Type', 'application/json');
        req.send(JSON.stringify({
            username: loggedInUser
        }));

        req.onload = function() {
            answer = this.responseText;
            //document.getElementById("loginMessage").innerHTML = answer;
            if(answer === "Admin success" && this.status === 200){
                adminAccess = true;
                $("div.admin").show();
               }
        }
    }
}

function createNewCategory(event){
    var req = new XMLHttpRequest();
    req.open("POST", "/createNewCategory", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        category:      event.category.value
    }));


    req.onload = function() {
        answer = this.responseText;
        document.getElementById("errorMessage").innerHTML = answer;
        if(answer === "") {
            document.getElementById("signedUpMessage").innerHTML = "Okey! Cool! Soooo, click here to login now :)";
        }
    }
}