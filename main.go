// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains the sample code for the GenerateContent API.
package main

/*
# For Vertex AI API
export GOOGLE_GENAI_USE_VERTEXAI=true
export GOOGLE_CLOUD_PROJECT={YOUR_PROJECT_ID}
export GOOGLE_CLOUD_LOCATION={YOUR_LOCATION}

# For Gemini AI API
export GOOGLE_GENAI_USE_VERTEXAI=false
export GOOGLE_API_KEY={YOUR_API_KEY}

go run samples/generate_text.go --model=gemini-1.5-pro-002
*/

import (
	"context"
	"flag"
	"fmt"
	"log"

	genai "google.golang.org/genai"

	"net/http"

	"github.com/gorilla/websocket"
)

var (
	model string
)

func init() {
	flag.StringVar(&model, "model", "gemini-2.0-flash-exp", "the model name, e.g. gemini-2.0-flash-exp")
}

//var prompt = flag.String("prompt", "say something nice to me", "the prompt to use")

func generateText(ctx context.Context, model, prompt string) *genai.GenerateContentResponse {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if client.ClientConfig.Backend == genai.BackendVertexAI {
		log.Println("Calling VertexAI.GenerateContent API...")
	} else {
		log.Println("Calling GeminiAI.GenerateContent API...")
	}
	// No configs are being used in this sample, explicitly set it to nil for clarity.
	var config *genai.GenerateContentConfig = nil
	// Call the GenerateContent method.
	result, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), config)
	if err != nil {
		log.Fatal(err)
	}
	// Marshal the result to JSON and pretty-print it to a byte array.
	//response, err := json.MarshalIndent(*result, "", "  ")
	//if err != nil {
	//	log.Fatal(err)
	//}
	// Log the output.
	//fmt.Println(string(response))
	return result
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

func reader(conn *websocket.Conn) {
	ctx := context.Background()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Client says:", string(p))

		// Generate
		result := generateText(ctx, model, string(p))
		text := result.Candidates[0].Content.Parts[0].Text
		b := []byte(text)
		log.Printf("Gemini says: %s", text)

		if err := conn.WriteMessage(messageType, b); err != nil {
			log.Println(err)
			return
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	reader(ws)
	defer ws.Close()
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {
	//ctx := context.Background()
	//flag.Parse()
	//generateText(ctx)

	model = "gemini-2.0-flash-exp"

	log.Println("Hello, WebSockets!")
	http.HandleFunc("/", home)
	http.HandleFunc("/ws", wsEndpoint)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
