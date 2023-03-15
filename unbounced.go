package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// `BouncedEmail` is a struct with two fields, `MessageID` and `Email`.
//
// The `json` tags are used to tell the `json` package how to encode and decode the struct.
//
// The `MessageID` field is a string, and the `json` package will encode it as a string.
//
// The `Email` field is also a string, and the `json` package will encode it as a string.
// @property {string} MessageID - The ID of the message that bounced.
// @property {string} Email - The email address that bounced.
type BouncedEmail struct {
	MessageID string `json:"messageId"`
	Email     string `json:"email"`
}

// BlockedContact is a struct with two fields, Email and MessageId, both of which are strings.
// @property {string} Email - The email address of the blocked contact.
// @property {string} MessageId - The ID of the message that was blocked.
type BlockedContact struct {
	Email     string `json:"email"`
	MessageId string `json:"messageId"`
}

// `GetTransacBlockedContactsResponse` is a struct with a field called `Contacts` of type
// `[]BlockedContact`.
// @property {[]BlockedContact} Contacts - The list of blocked contacts.
type GetTransacBlockedContactsResponse struct {
	Contacts []BlockedContact `json:"contacts"`
}

// It gets all the bounced emails from the database, unblocks them, and returns the number of emails
// that were unblocked
func handleBouncedEmails() (int, error) {
	bouncedEmails, err := getBouncedEmails()
	if err != nil {
		return 0, err
	}

	for _, email := range bouncedEmails {
		err = unblockEmail(email.Email)
		if err != nil {
			return 0, err
		}
	}

	return len(bouncedEmails), nil
}

// It gets the list of bounced emails from Sendinblue's API and returns it as a slice of `BouncedEmail`
// structs
func getBouncedEmails() ([]BouncedEmail, error) {
	response, err := getTransacBlockedContacts()
	if err != nil {
		return nil, err
	}

	bouncedEmails := make([]BouncedEmail, len(response.Contacts))
	for i, contact := range response.Contacts {
		bouncedEmails[i] = BouncedEmail{
			MessageID: contact.MessageId,
			Email:     contact.Email,
		}
	}

	return bouncedEmails, nil
}

// It makes a GET request to the Sendinblue API, and returns the response as a
// `GetTransacBlockedContactsResponse` struct
func getTransacBlockedContacts() (*GetTransacBlockedContactsResponse, error) {
	req, err := http.NewRequest("GET", sendinblueURL+"smtp/blockedContacts?sort=desc", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("api-key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result GetTransacBlockedContactsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// It makes a DELETE request to the Sendinblue API to unblock the email address
func unblockEmail(email string) error {
	fmt.Println(email)
	req, err := http.NewRequest("DELETE", sendinblueURL+"smtp/blockedContacts/"+email, nil)
	if err != nil {
		return err
	}

	req.Header.Set("api-key", apiKey)
	client := &http.Client{}
	_, err = client.Do(req)
	return err
}

// It calls the handleBouncedEmails() function, which returns the number of bounced emails processed
// and an error. If there's an error, it prints it. If there's no error, it prints a success message
// and then calls itself again if there were any bounced emails processed
func Unbounced() {
	num, err := handleBouncedEmails()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Bounced emails processed successfully.")
		if num > 0 {
			Unbounced()
		}
	}
}
