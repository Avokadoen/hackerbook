function postTopic(event){
    var answer = "Something went wrong";
    var req = new XMLHttpRequest();

    if(isLoggedIn()){
        var url = [location.protocol, '//', location.host, location.pathname+"/newtopic"].join('');
        req.open("POST", url, true);
        req.setRequestHeader('Content-Type', 'application/json');
        req.send(JSON.stringify({
            username: loggedInUser, //Fetch username somehow
            title: event.title.value,
            content: event.topicInput.value
        }));


        req.onload = function() {
            answer = this.responseText;
            document.getElementById("postedTopicMessage").innerHTML = answer;
            if(req.status == 201) { //if StatusCreated
                location.reload(true) //reload, force new GET request, i.e. don't use cache
            }
        }
    }
    else {
        document.getElementById("postedTopicMessage").innerHTML = "Not logged in, please log in to post topic.";
    }
}


function postTopicVisible(){
    $("div.postNewTopic").show();
}
