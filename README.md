# SoftSecOblig2


## Initial setup
- Make sure you have a working go environment of version 1.11 or higher:
  - To do this you can follow this [guide](https://golang.org/doc/install)
- [Download go dep](https://github.com/golang/dep)
- Do "go get -u gitlab.com/avokadoen/softsecoblig2/..."
- Run dep ensure in root of repo directory
- Fill the values of the .envtmpl to your db information. we used mlab, but you could host a local mongodb for local hosting
- For Captcha to work on registration page, you have to register on [Google's reCaptcha admin page](https://www.google.com/recaptcha/admin), and update to your public data-sitekey in the .env file

## How to host locally:
- After doing the initial setup do "go run ./cmd/forum/" in root folder

## How to host on heroku:

- Make sure you have the heroku CLI set up by following this [guide](https://devcenter.heroku.com/articles/getting-started-with-go#set-up)
- With a working CLI you can run <b>"heroku create"</b>
- For each variable in your .env file you have to do <b>"heroku config:set *variablename*=*value*"</b>. For more information follow this [guide](https://devcenter.heroku.com/articles/config-vars)
- Then do <b>"git push heroku master"</b>
- If all your variable are defined correctly according to your mongodb then you can run <b>"heroku open"</b>
    
