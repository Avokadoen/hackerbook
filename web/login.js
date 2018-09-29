function handleLogin(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    req.open("POST", "/postlogin", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        username: event.username,
        password: event.password
    }));


    req.onload = function() {
        answer = this.responseText;
        document.getElementById("loginmessage").innerHTML = answer + " " + event.username.value + " " + event.password.value;
    }
}

$("#usernameInput").on('input', function (entry){
    document.getElementById("testpara").innerHTML = entry.target.value;
});