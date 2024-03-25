package llm

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"bufio"
	"strings"
	"context"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"github.com/autogorg/autog"
)

const (
	ollamaDefaultBaseURL        = "http://localhost:11434"
	ollamaDefaultVendor         = "ollama"
	ollamaDefaultModel          = "gemma:2b"
	ollamaDefaultModelWeak      = "gemma:2b"
	ollamaDefaultModelEmbed     = "nomic-embed-text"
)

// OllamaAPIError represents an error that occurred on an API
type OllamaAPIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

// Error returns a string representation of the error
func (e OllamaAPIError) Error() string {
	return fmt.Sprintf("[%d:%s] %s", e.StatusCode, e.Type, e.Message)
}

// OllamaAPIErrorResponse is the full error response that has been returned by an API.
type OllamaAPIErrorResponse struct {
	Error OllamaAPIError `json:"error"`
}

type OllamaChatCompletionRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaOptions struct {
	//NumCtx      int     `json:num_ctx,omitempty`
	//Seed        int     `json:seed,omitempty`
	//TopP        float32 `json:top_p,omitempty`
	//MaxTokens   int     `json:num_predict,omitempty`
	Temperature float32 `json:"temperature,omitempty"`
}

type OllamaChatCompletionRequest struct {
	// Model is the name of the model to use.
	Model string `json:"model"`
	// Messages is a list of messages to use as the context for the chat completion.
	Messages []OllamaChatCompletionRequestMessage `json:"messages"`
	// the format to return a response in. Currently the only accepted value is json
	Format string `json:format,omitempty`
	// additional model parameters listed in the documentation for the Modelfile such as temperature
	Options OllamaOptions `json:"options,omitempty"`
	// the prompt template to use (overrides what is defined in the Modelfile)
	Template string `json:template,omitempty`
	// if false the response will be returned as a single response object, rather than a stream of objects
	Stream bool `json:"stream,omitempty"`
	// controls how long the model will stay loaded into memory following the request (default: 5m)
	KeepAlive int `json:"keep_alive,omitempty"`
}

type OllamaChatCompletionResponse struct {
	Model      string                                `json:"model"`
	CreatedAt  string                                `json:"created_at"`
	Done       bool                                  `json:"done"`
	Message    OllamaChatCompletionResponseMessage   `json:"message,omitempty"`
	TotalDuration      int `json:"total_duration,omitempty"`
	LoadDuration       int `json:"load_duration,omitempty"`
	PromptEvalCount    int `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int `json:"prompt_eval_duration,omitempty"`
	EvalCount          int `json:"eval_count,omitempty"`
	EvalCuration       int `json:"eval_duration,omitempty"`
}

type OllamaChatCompletionStreamResponse struct {
	Model      string                                `json:"model"`
	CreatedAt  string                                `json:"created_at"`
	Done       bool                                  `json:"done"`
	Message    OllamaChatCompletionResponseMessage   `json:"message,omitempty"`
	TotalDuration      int `json:"total_duration,omitempty"`
	LoadDuration       int `json:"load_duration,omitempty"`
	PromptEvalCount    int `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int `json:"prompt_eval_duration,omitempty"`
	EvalCount          int `json:"eval_count,omitempty"`
	EvalCuration       int `json:"eval_duration,omitempty"`
}

type OllamaEmbeddingRequest struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	// additional model parameters listed in the documentation for the Modelfile such as temperature
	// Options OllamaOptions `json:"options,omitempty"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type Ollama struct {
	ApiKey      string
	ApiBase     string
	ApiVendor   string
	ApiOrg      string
	Model       string
	ModelWeak   string
	ModelEmbedding  string
	Temperature     int
	TemperatureWeak int
	TimeOut       int
	TimeOutWeak   int
	MaxTokens     int
	MaxTokensWeak int
	Verbose       int
	VerboseLog    func(log string)
	
	httpMain  *http.Client
	httpWeak  *http.Client
	httpEmbed *http.Client
}


func (gpt *Ollama) InitLLM() error {
	/*
	// No Need!
	if len(gpt.ApiKey) <= 0 {
		return fmt.Errorf("API Key is needed!")
	}*/
	if len(gpt.ApiBase) <= 0 {
		gpt.ApiBase = ollamaDefaultBaseURL
	}
	if len(gpt.ApiVendor) <= 0 {
		gpt.ApiVendor = ollamaDefaultVendor
	}
	if len(gpt.Model) <= 0 {
		gpt.Model = ollamaDefaultModel
	}
	if len(gpt.ModelWeak) <= 0 {
		gpt.ModelWeak = ollamaDefaultModelWeak
	}
	if len(gpt.ModelEmbedding) <= 0 {
		gpt.ModelEmbedding = ollamaDefaultModelEmbed
	}
	if gpt.Temperature <= 0 {
		// TODO: Changed by model
		gpt.Temperature = 0
	}
	if gpt.TemperatureWeak <= 0 {
		// TODO: Changed by model
		gpt.TemperatureWeak = 0
	}
	if gpt.TimeOut <= 0 {
		// TODO: Changed by model
		gpt.TimeOut = 300
	}
	if gpt.TimeOutWeak <= 0 {
		// TODO: Changed by model
		gpt.TimeOutWeak = 300
	}
	if gpt.MaxTokens <= 0 {
		// TODO: Changed by model
		gpt.MaxTokens = 0
	}
	if gpt.MaxTokensWeak <= 0 {
		// TODO: Changed by model
		gpt.MaxTokensWeak = 0
	}

	gpt.httpMain = &http.Client{
		Timeout: time.Duration(gpt.TimeOut) * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	gpt.httpWeak = &http.Client{
		Timeout: time.Duration(gpt.TimeOutWeak) * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	gpt.httpEmbed = &http.Client{
		Timeout: time.Duration(gpt.TimeOutWeak) * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return nil
}

func (gpt *Ollama) ConvertMessages(msgs []autog.ChatMessage) []OllamaChatCompletionRequestMessage {
	reqmsgs := make([]OllamaChatCompletionRequestMessage, len(msgs))
	for i, msg := range msgs {
		reqmsgs[i] = OllamaChatCompletionRequestMessage{Role: msg.Role, Content: msg.Content}
	}
	return reqmsgs
}


func (gpt *Ollama) JsonBodyReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return bytes.NewBuffer(nil), nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Failed encoding json: %w", err)
	}
	return bytes.NewBuffer(raw), nil
}

func (gpt *Ollama) CreateChatCompletionRequest(weak, stream bool, msgs []autog.ChatMessage) *OllamaChatCompletionRequest{
	model       := gpt.Model
	temperature := gpt.Temperature
	maxtokens   := gpt.MaxTokens

	if weak {
		model       = gpt.ModelWeak
		temperature = gpt.TemperatureWeak
		maxtokens   = gpt.MaxTokensWeak
	}

	if maxtokens > 0 {
		return &OllamaChatCompletionRequest{
			Messages    : gpt.ConvertMessages(msgs),
			Model       : model,
			Stream      : stream,
			Options     : OllamaOptions{
				Temperature : float32(temperature) / float32(100),
				// MaxTokens   : maxtokens,
			},
		}
	}

	return &OllamaChatCompletionRequest{
		Messages    : gpt.ConvertMessages(msgs),
		Model       : model,
		Stream      : stream,
		Options     : OllamaOptions{
			Temperature : float32(temperature) / float32(100),
		},
	}
}

func (gpt *Ollama) CreateHttpRequest(cxt context.Context, method, path string, payload interface{}) (*http.Request, error) {
	jsonBody, err := gpt.JsonBodyReader(payload)
	if err != nil {
		return nil, err
	}
	url := gpt.ApiBase + path
	req, err := http.NewRequestWithContext(cxt, method, url, jsonBody)
	if err != nil {
		return nil, err
	}
	/*
	// No need for Ollama
	if len(gpt.ApiOrg) > 0 {
		req.Header.Set("OpenAI-Organization", gpt.ApiOrg)
	}
	*/
	req.Header.Set("Content-type", "application/json")
	// No need for Ollama
	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gpt.ApiKey))
	return req, nil
}

func (gpt *Ollama) CheckHttpResponseSuccess(httpRsp *http.Response) error {
	if httpRsp.StatusCode >= 200 && httpRsp.StatusCode < 300 {
		return nil
	}
	defer httpRsp.Body.Close()
	data, err := io.ReadAll(httpRsp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read from body: %w", err)
	}
	var result OllamaAPIErrorResponse
	if err := json.Unmarshal(data, &result); err != nil {
		apiError := OllamaAPIError{
			StatusCode: httpRsp.StatusCode,
			Type:       "Unexpected",
			Message:    string(data),
		}
		return apiError
	}
	result.Error.StatusCode = httpRsp.StatusCode
	return result.Error
}

func (gpt *Ollama) GetHttpResponse(httpClient *http.Client, httpReq *http.Request) (*http.Response, error) {
	httpRsp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err := gpt.CheckHttpResponseSuccess(httpRsp); err != nil {
		return nil, err
	}

	return httpRsp, nil
}

func (gpt *Ollama) GetHttpBodyObject(httpRsp *http.Response, response interface{}) error {
	defer httpRsp.Body.Close()
	if err := json.NewDecoder(httpRsp.Body).Decode(response); err != nil {
		return fmt.Errorf("Invalid json response: %w", err)
	}
	return nil
}

func (gpt *Ollama) SendMessagesInner(cxt context.Context, msgs []autog.ChatMessage, weak bool) (autog.LLMStatus, autog.ChatMessage) {
	request := gpt.CreateChatCompletionRequest(weak, false, msgs)

	if gpt.Verbose >= autog.VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/api/chat", request)
	if err != nil {
		return autog.LLM_STATUS_BED_REQUEST, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: err.Error()}
	}
	httpClient   := gpt.httpMain
	if weak {
		httpClient = gpt.httpWeak
	}
	httpRsp, err := gpt.GetHttpResponse(httpClient, httpReq)
	if err != nil {
		return autog.LLM_STATUS_BED_RESPONSE, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: err.Error()}
	}
	response := OllamaChatCompletionResponse{}
	if err := gpt.GetHttpBodyObject(httpRsp, &response); err != nil {
		return autog.LLM_STATUS_BED_MESSAGE, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: err.Error()}
	}

	if gpt.Verbose >= autog.VerboseShowReceiving {
		repstr, reperr := json.Marshal(response)
		if reperr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECV_FROM_LLVM:\n %s \n", repstr))
		}
	}

	revMsg := autog.ChatMessage{
		Role    : response.Message.Role,
		Content : response.Message.Content,
	}

	return autog.LLM_STATUS_OK, revMsg
}


func (gpt *Ollama) SendMessagesStreamInner(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader, weak bool) (autog.LLMStatus, autog.ChatMessage) {
	request := gpt.CreateChatCompletionRequest(weak, true, msgs)

	if gpt.Verbose >= autog.VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	var contentbuf *strings.Builder
	if reader != nil {
		contentbuf = reader.StreamStart()
	}
	if contentbuf == nil {
		contentbuf = &strings.Builder{}
	}

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/api/chat", request)
	if err != nil {
		if reader != nil {
			reader.StreamError(contentbuf, autog.LLM_STATUS_BED_REQUEST, err.Error())
			reader.StreamEnd(contentbuf)
		}
		return autog.LLM_STATUS_BED_REQUEST, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: err.Error()}
	}
	httpClient   := gpt.httpMain
	if weak {
		httpClient = gpt.httpWeak
	}
	httpRsp, err := gpt.GetHttpResponse(httpClient, httpReq)
	if err != nil {
		if reader != nil {
			reader.StreamError(contentbuf, autog.LLM_STATUS_BED_REQUEST, err.Error())
			reader.StreamEnd(contentbuf)
		}
		return autog.LLM_STATUS_BED_RESPONSE, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: err.Error()}
	}
	bufreader := bufio.NewReader(httpRsp.Body)
	defer httpRsp.Body.Close()

	var readErr error
	var line []byte
	for {
		line, readErr = bufreader.ReadBytes('\n')
		if readErr != nil {
			break
		}

		response := OllamaChatCompletionStreamResponse{}
		if err := json.Unmarshal(line, &response); err != nil {
			readErr = fmt.Errorf("Invalid json stream data: %v", err)
			break
		}

		delta := response.Message.Content
		if contentbuf != nil {
			contentbuf.WriteString(delta)
		}
		if reader != nil {
			reader.StreamDelta(contentbuf, delta)
		}

		if response.Done {
			break
		}
	}
	if reader != nil {
		if readErr != nil {
			reader.StreamError(contentbuf, autog.LLM_STATUS_BED_MESSAGE, readErr.Error())
		}
		reader.StreamEnd(contentbuf)
	}

	if readErr != nil {
		return autog.LLM_STATUS_BED_MESSAGE, autog.ChatMessage{Role:autog.ROLE_ASSISTANT, Content: readErr.Error()}
	}

	revMsg := autog.ChatMessage{
		Role    : autog.ROLE_ASSISTANT,
		Content : contentbuf.String(),
	}

	return autog.LLM_STATUS_OK, revMsg
}

func (gpt *Ollama) CalcTokens(cxt context.Context, content string) int {
	// TODO: a fake tokens calc, will fix it with actual tokens calc logic
	return len(content) / 6
}

func (gpt *Ollama) SendMessages(cxt context.Context, msgs []autog.ChatMessage) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesInner(cxt, msgs, false)

	if gpt.Verbose >= autog.VerboseShowReceiving {
		msgstr, err := json.Marshal(msg)
		if err == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECEIVE_FROM_LLVM STATUS(%d)\n %s \n", status, msgstr))
		}
	}

	return status, msg
}

func (gpt *Ollama) SendMessagesStream(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesStreamInner(cxt, msgs, reader, false)

	if gpt.Verbose >= autog.VerboseShowReceiving {
		msgstr, err := json.Marshal(msg)
		if err == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECEIVE_FROM_LLVM STATUS(%d)\n %s \n", status, msgstr))
		}
	}

	return status, msg
}

func (gpt *Ollama) CalcTokensByWeakModel(cxt context.Context, content string) int {
	// TODO: a fake tokens calc, will fix it with actual tokens calc logic
	return len(content) / 6
}

func (gpt *Ollama) SendMessagesByWeakModel(cxt context.Context, msgs []autog.ChatMessage) (autog.LLMStatus, autog.ChatMessage) {
	return gpt.SendMessagesInner(cxt, msgs, true)
}

func (gpt *Ollama) SendMessagesStreamByWeakModel(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader) (autog.LLMStatus, autog.ChatMessage) {
	return gpt.SendMessagesStreamInner(cxt, msgs, reader, true)
}

func (gpt *Ollama) Embedding(cxt context.Context, dimensions int, text string) (autog.Embedding, error) {
	var embed autog.Embedding
	embeddingReq := OllamaEmbeddingRequest{
		Prompt: text,
		Model: gpt.ModelEmbedding,
	}
	if dimensions > 0 {
		return embed, fmt.Errorf("dimensions is not support!")
	}
	request := embeddingReq

	if gpt.Verbose >= autog.VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	httpReq, cerr := gpt.CreateHttpRequest(cxt, "POST", "/api/embeddings", request)
	if cerr != nil {
		return embed, cerr
	}

	httpRsp, gerr := gpt.GetHttpResponse(gpt.httpEmbed, httpReq)
	if gerr != nil {
		return embed, gerr
	}

	response := OllamaEmbeddingResponse{}
	if rerr := gpt.GetHttpBodyObject(httpRsp, &response); rerr != nil {
		return embed, rerr
	}

	if gpt.Verbose >= autog.VerboseShowReceiving {
		repstr, reperr := json.Marshal(response)
		if reperr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECV_FROM_LLVM:\n %s \n", repstr))
		}
	}

	embed = autog.Embedding{}
	for _, f := range response.Embedding {
		embed = append(embed, float64(f))
	}

	return embed, nil
}

func (gpt *Ollama) Embeddings(cxt context.Context, dimensions int, texts []string) ([]autog.Embedding, error) {
	var embeds []autog.Embedding
	var err error
	embeds = make([]autog.Embedding, len(texts))
	for i, text := range texts {
		embeds[i], err = gpt.Embedding(cxt, dimensions, text)
		if err != nil {
			break
		}
	}
	return embeds, err
}