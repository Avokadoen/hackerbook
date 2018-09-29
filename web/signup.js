var PW = new String();

function handleSignup(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    if(validateInput(event.username, event.email, event.password, $('#confirmPasswordInput').val())){
        req.open("POST", "/signup", true);
        req.setRequestHeader('Content-Type', 'application/json');
        req.send(JSON.stringify({
            email:      event.email.value,
            username:   event.username.value,
            password:   event.password.value
        }));


        req.onload = function() {
            answer = this.responseText;
            document.getElementById("errorMessage").innerHTML = answer;
        }
    }
    else{
        // Something went wrong, deal with it?
    }
}



$("#usernameInput").on('input', function (entry){
    if(entry.target.value.match(/[^a-zA-Z0-9 ]/g)){
        document.getElementById("usernameMessage").innerHTML = "Username can only contain letters and numbers";
    }
    else if(entry.target.value.length > 15){
        document.getElementById("usernameMessage").innerHTML = "Username can only be 14 characters long";
    }
    else{
        document.getElementById("usernameMessage").innerHTML = "";
    }
});

$("#emailInput").on('input', function (entry){
    if(!entry.target.value.match(/^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/)){
        document.getElementById("emailMessage").innerHTML = "Email not valid";
    }
    else{
        document.getElementById("emailMessage").innerHTML = "";
    }
});

$("#passwordInput").on('input', function (entry){
    if(entry.target.value.match(/[^a-zA-Z0-9 ]/g)){
        document.getElementById("passwordMessage").innerHTML = "Password can only contain letters and numbers";
    }
    else{
        document.getElementById("passwordMessage").innerHTML = "";
    }
    PW = entry.target.value;
});

$("#confirmPasswordInput").on('input', function (entry){
    if(entry.target.value != PW){
        document.getElementById("confirmPasswordMessage").innerHTML = "Passwords doesn't match!";
    }
    else{
        document.getElementById("confirmPasswordMessage").innerHTML = "";
    }
});


function validateUsername(username){
    if(username.value.match(/[^a-zA-Z0-9 ]/g)){
        return false;
    }
    else if(username.value.length > 15){
        return false;
    }
    else{
        return true;
    }
}

function validateEmail(email){
    if(!email.value.match(/^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/)){
        return false;
    }
    else{
        return true;
    }
}

function validatePassword(password){
    if(password.value.match(/[^a-zA-Z0-9 ]/g)){
        return false;
    }
    else{
        return true;
    }
}

function validateConfirmPassword(confirmPassword, password){
    if(password.value != confirmPassword){
        return false;
    }
    else{
        return true;
    }
}

// TODO: Create more informative info back to user, where and what went wrong
function validateInput(username, email, password, confirmPassword){
    if(!validateUsername(username)){
        document.getElementById("errorMessage").innerHTML = "Username not valid";
        return false;
    }
    else if(!validateEmail(email)){
        document.getElementById("errorMessage").innerHTML = "Email not valid";
        return false;
    }
    else if(!validatePassword(password)){
        document.getElementById("errorMessage").innerHTML = "Password not valid";
        return false;
    }
    else if(!validateConfirmPassword(confirmPassword, password)){
        document.getElementById("errorMessage").innerHTML = "Passwords does not match";
        return false;
    }
    else{
        return true;
    }
}