var loginClone;

function handleLogin(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    req.open("POST", "/postlogin", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        username: event.username.value,
        password: event.password.value
    }));


    req.onload = function() {
        answer = this.responseText;
        document.getElementById("loginMessage").innerHTML = answer;
        if(answer === "login successful"){
            loginClone = $("#login").clone()
            $('#login').html("You are logged in as " + event.username.value  + "<input type=\"button\" onClick=\"handleSignout()\" value=\"Signout\">");
        }
    }
}

function handleSignout(){
    $('#login').html(loginClone);
    document.getElementById("loginMessage").innerHTML = "Signed out successfully";
}