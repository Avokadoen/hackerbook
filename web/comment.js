function postComment(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    req.open("POST", "/postcomment", true);
    req.setRequestHeader('Content-Type', 'application/json');
    req.send(JSON.stringify({
        username: event.username.value, //Fetch username somehow
        comment: event.commentInput.value

    }));


    req.onload = function() {
        answer = this.responseText;
        document.getElementById("commentMessage").innerHTML = answer;
        if(answer === "Comment posted"){ //If successful, display message

        }
    }
}