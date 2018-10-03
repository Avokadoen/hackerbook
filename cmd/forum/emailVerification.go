package main

import (
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
	"net/http"
)

func EmailVerification(w http.ResponseWriter, r *http.Request, user database.SignUpUser) {
	/*emailToken := app.CreateHash(user.Email)
	token := database.EmailToken{
		Username: user.Username,
		Token:    emailToken,
	}
	Server.Database.InsertToCollection(database.TableEmailToken, token)
	link := r.URL.Hostname() + "/email-verification@token=" + emailToken
	// TODO send email with link
	serverName := os.Getenv("SMTPSERVER")
	email := os.Getenv("EMAIL")
	port := os.Getenv("SMTPPORT")
	emailPw := os.Getenv("EMAILPW")
	auth := smtp.PlainAuth("", email, emailPw, serverName)
	message := "Click here to verify user : " + link
	err := smtp.SendMail(serverName+":"+port, auth, email, []string{user.Email}, []byte(message))
	if err != nil {
		log.Printf("Failed to send mail: %v", err)
	}*/
	//tlsconfig := &tls.Config{
	//	InsecureSkipVerify: true,
	//	ServerName: serverName,
	//}
	//c, err := smtp.Dial(tlsconfig.ServerName + ":" + port)
	//if err != nil {
	//	log.Printf("Dialup failed: %v" ,err)
	//}
	////c.StartTLS(tlsconfig)
	//if err := c.Auth(auth); err != nil {
	//	log.Printf("Connection authentication failed: %v",err)
	//}
	//
	//// Set the sender and recipient first
	//if err := c.Mail(email); err != nil {
	//	log.Printf("Set sender failed: %v",err)
	//}
	//if err := c.Rcpt(user.Email); err != nil {
	//	log.Printf("Set recipient failed: %v",err)
	//}
	//wc, err := c.Data()
	//if err != nil {
	//	log.Printf("Failed something: %v",err) //TODO beire melding
	//}
	//
	//message := "Click here to verify user : " + link
	//_, err = wc.Write([]byte(message))
	//if err != nil{
	//	log.Printf("Failed to write message: %v",err)
	//}
	//
	//err = wc.Close()
	//if err != nil{
	//	log.Printf("Failed to send email: %v",err)
	//}
	//c.Quit()
}
