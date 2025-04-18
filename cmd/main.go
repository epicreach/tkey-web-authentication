// Package main starts the client application and provides the user with a choice between a command-line client and a web client.
// The command-line client allows the user to register and log in to the application.
// The web client allows the user to register and log in to the application through a web interface.
package main

import (
	"chalmers/tkey-group22/client/internal/auth"
	. "chalmers/tkey-group22/client/internal/structs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("Starting web client...")
	startWebClient()
}

// Starts http listeners for the web client to use
func startWebClient() {
	http.Handle("/api/register", enableCors(http.HandlerFunc(registerHandler)))
	http.Handle("/api/login", enableCors(http.HandlerFunc(loginHandler)))
	http.Handle("/api/add-public-key", enableCors(http.HandlerFunc(addPublicKeyHandler)))
	http.Handle("/api/remove-public-key", enableCors(http.HandlerFunc(removePublicKeyHandler)))
	fmt.Println("Client running on http://localhost:6060")
	http.ListenAndServe(":6060", nil)
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Gets a username to attempt to sign in on. Will return a signed challenge. It expects a POST
// request with a JSON body containing a username.

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Get origin from request header and replace port with 8080
	// We use this order to be able to know what to send to auth.Login
	origin := r.Header.Get("Origin")

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	username := requestBody["username"]
	user, signedChallenge, errMsg, err := auth.GetAndSign(origin, username)
	if err != nil {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	response := GetAndSignResponse{
		User:            user,
		SignedChallenge: signedChallenge,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Handles register requests from the web client
// It expects a POST request with a JSON body containing the username and a pubkey label
// The handler retrieves the new public key from the TKey and sends a request to the backend to add the new public key
//
// Possible responses:
// - 400 Bad Request: if the request body is invalid or cannot be parsed
// - 500 Internal Server Error: if there is an error adding the public key
// - 200 OK: if the public key is added successfully
//
//
//	Error messages:
//
//	If an error occurs the function will return an http Error containing both the error code but also an error message retrieved from the applications response
//	to the request. This response is later retrieved by the frontend and displayed to the user.

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Get origin from request header and replace port with 8080
	origin := r.Header.Get("Origin")

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	username := requestBody["username"]
	label := requestBody["label"]
	resp, err := auth.Register(origin, username, label)
	if err != nil {
		if resp == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)

		if err != nil {
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)
			return
		}

		respBodyStr := string(respBody)
		http.Error(w, respBodyStr, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User registered successfully"))
}

// Handles add public key requests from the web client
// It expects a POST request with a JSON body containing the username and a pubkey label
// The handler retrieves the new public key from the TKey and sends a request to the backend to add the new public key
//
// Possible responses:
// - 400 Bad Request: if the request body is invalid or cannot be parsed
// - 500 Internal Server Error: if there is an error adding the public key
// - 200 OK: if the public key is added successfully
func addPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	// Get origin from request header and replace port with 8080
	origin := r.Header.Get("Origin")

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	username := requestBody["username"]
	label := requestBody["label"]
	sessionCookie := r.Header.Get("Cookie")
	err := auth.AddPublicKey(origin, username, label, sessionCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Public key added successfully"))
}

// Handles remove public key requests from the web client
// It expects a POST request with a JSON body containing the username and the label of the public key to be removed
// The handler sends a request to the backend to remove the public key
//
// Possible responses:
// - 400 Bad Request: if the request body is invalid or cannot be parsed
// - 500 Internal Server Error: if there is an error removing the public key
// - 200 OK: if the public key is removed successfully
func removePublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	// Get origin from request header and replace port with 8080
	origin := r.Header.Get("Origin")

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	username := requestBody["username"]
	label := requestBody["label"]
	sessionCookie := r.Header.Get("Cookie")
	err := auth.RemovePublicKey(origin, username, label, sessionCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Public key removed successfully"))
}
