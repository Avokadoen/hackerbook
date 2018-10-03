function postComment(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    if(isLoggedIn()){
        var url = [location.protocol, '//', location.host, location.pathname+"/comment"].join('');
        req.open("POST", url, true);
        req.setRequestHeader('Content-Type', 'application/json');
        req.send(JSON.stringify({
            username: loggedInUser, //Fetch username somehow
            text: event.commentInput.value
        }));


        req.onload = function() {
            answer = this.responseText;
            document.getElementById("commentMessage").innerHTML = answer;
            if(req.status == 201) { //if StatusCreated
                location.reload(true) //reload, force new GET request, i.e. don't use cache
            }
        }
    }
    else {
        document.getElementById("commentMessage").innerHTML = "Not logged in, please log in to post comment.";
    }
}
