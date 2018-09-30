var loginClone;
var loggedInUser = "";

function tryCookieLogin(){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    req.open("POST", window.location.origin + "/cookielogin", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send();

    req.onload = function() {
        answer = this.responseText;
        //document.getElementById("loginMessage").innerHTML = answer;
        if(answer !== "" && this.status === 200){
            loggedInUser = answer;
            loginClone = $("#login").clone();
            $('#login').html("You are logged in as " + loggedInUser  + "<input type=\"button\" onClick=\"handleSignout()\" value=\"Signout\">");
        }
    }
}

function handleLogin(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    req.open("POST", window.location.origin + "/postlogin", true); //Postlogin does currently not work from a topic
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        username: event.username.value,
        password: event.password.value
    }));


    req.onload = function() {
        answer = this.responseText;
        document.getElementById("loginMessage").innerHTML = answer;
        if(answer === "login successful"){
            loggedInUser = event.username.value;
            loginClone = $("#login").clone();
            $('#login').html("You are logged in as " + event.username.value  + "<input type=\"button\" onClick=\"handleSignout()\" value=\"Signout\">");
        }
    }
}

function handleSignout(){
    $('#login').html(loginClone);
    document.getElementById("loginMessage").innerHTML = "Signed out successfully";
}

function isLoggedIn(){
    if(loggedInUser !== ""){
        return true;
    }
    else{
        // Not logged in, give error message?
        return false;
    }
}

