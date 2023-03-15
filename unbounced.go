package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BouncedEmail struct {
	MessageID string `json:"messageId"`
	Email     string `json:"email"`
}

type BlockedContact struct {
	Email     string `json:"email"`
	MessageId string `json:"messageId"`
}

type GetTransacBlockedContactsResponse struct {
	Contacts []BlockedContact `json:"contacts"`
}

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

		/*
			res, err := resendEmail(email.MessageID)
			if err != nil {
				return err
			}

			fmt.Println("Resend result:", res)
		*/
	}

	return len(bouncedEmails), nil
}

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

func resendEmail(messageID string) (string, error) {
	req, err := http.NewRequest("POST", sendinblueURL+"smtp/emailMessageId/"+messageID+"/send", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("api-key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Status: %v, Message: %v", resp.Status, result["message"]), nil
}

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
