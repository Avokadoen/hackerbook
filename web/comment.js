function postComment(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    if(isLoggedIn()){
        req.open("POST", window.location.origin + "/postcomment", true);
        req.setRequestHeader('Content-Type', 'application/json');
        req.send(JSON.stringify({
            username: loggedInUser, //Fetch username somehow
            comment: event.commentInput.value

        }));


        req.onload = function() {
            answer = this.responseText;
            document.getElementById("commentMessage").innerHTML = answer;
            if(answer === "Comment posted"){ //If successful, display message

            }
        }
    }
    else {
        document.getElementById("commentMessage").innerHTML = "Not logged in, please log in to post comment.";
    }
}