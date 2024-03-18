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
	"autog/autog"
)

const (
	defaultBaseURL        = "https://api.chatpp.org/v1"
	defaultVendor         = "openai"
	defaultModel          = "gpt-4-turbo-preview"
	defaultModelWeak      = "gpt-4-turbo-preview"
)

var (
	dataPrefix = []byte("data: ")
	donePrefix = []byte("[DONE]")
)

// APIError represents an error that occurred on an API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

// Error returns a string representation of the error
func (e APIError) Error() string {
	return fmt.Sprintf("[%d:%s] %s", e.StatusCode, e.Type, e.Message)
}

// APIErrorResponse is the full error response that has been returned by an API.
type APIErrorResponse struct {
	Error APIError `json:"error"`
}

type OpenaiChatCompletionRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenaiChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenaiChatCompletionRequest struct {
	// Model is the name of the model to use. If not specified, will default to gpt-3.5-turbo.
	Model string `json:"model"`
	// Messages is a list of messages to use as the context for the chat completion.
	Messages []OpenaiChatCompletionRequestMessage `json:"messages"`
	// Temperature is sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random,
	// while lower values like 0.2 will make it more focused and deterministic
	Temperature float32 `json:"temperature,omitempty"`
	// TopP is an alternative to sampling with temperature, called nucleus sampling, where the model considers the results of
	// the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP float32 `json:"top_p,omitempty"`
	// N is number of responses to generate
	N int `json:"n,omitempty"`
	// Stream is whether to stream responses back as they are generated
	Stream bool `json:"stream,omitempty"`
	// Stop is up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`
	// MaxTokens is the maximum number of tokens to return.
	MaxTokens int `json:"max_tokens,omitempty"`
	// PresencePenalty (-2, 2) penalize tokens that haven't appeared yet in the history.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`
	// FrequencyPenalty (-2, 2) penalize tokens that appear too frequently in the history.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	// LogitBias modify the probability of specific tokens appearing in the completion.
	LogitBias map[string]float32 `json:"logit_bias,omitempty"`
	// User can be used to identify an end-user
	User string `json:"user,omitempty"`
}

type OpenaiChatCompletionResponseChoice struct {
	Index        int                                 `json:"index"`
	FinishReason string                              `json:"finish_reason"`
	Message      OpenaiChatCompletionResponseMessage `json:"message"`
}

type OpenaiChatCompletionStreamResponseChoice struct {
	Index        int                                 `json:"index"`
	FinishReason string                              `json:"finish_reason"`
	Delta        OpenaiChatCompletionResponseMessage `json:"delta"`
}

type OpenaiChatCompletionsResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenaiChatCompletionResponse struct {
	ID      string                               `json:"id"`
	Object  string                               `json:"object"`
	Created int                                  `json:"created"`
	Model   string                               `json:"model"`
	Choices []OpenaiChatCompletionResponseChoice `json:"choices"`
	Usage  OpenaiChatCompletionsResponseUsage    `json:"usage"`
}

type OpenaiChatCompletionStreamResponse struct {
	ID      string                                     `json:"id"`
	Object  string                                     `json:"object"`
	Created int                                        `json:"created"`
	Model   string                                     `json:"model"`
	Choices []OpenaiChatCompletionStreamResponseChoice `json:"choices"`
	Usage   OpenaiChatCompletionsResponseUsage         `json:"usage"`
}

const (
	VerboseNone int = iota
	VerboseShowSending
	VerboseShowReceiving
)

type OpenAi struct {
	ApiKey      string
	ApiBase     string
	ApiVendor   string
	ApiOrg      string
	Model       string
	ModelWeak   string
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
}

func (gpt *OpenAi) InitLLM() error {
	if len(gpt.ApiKey) <= 0 {
		return fmt.Errorf("API Key is needed!")
	}
	if len(gpt.ApiBase) <= 0 {
		gpt.ApiBase.ApiBase = defaultBaseURL
	}
	if len(gpt.ApiBase.ApiVendor) <= 0 {
		gpt.ApiBase.ApiVendor = defaultVendor
	}
	if len(gpt.Model) <= 0 {
		gpt.Model = defaultModel
	}
	if len(gpt.ModelWeak) <= 0 {
		gpt.ModelWeak = defaultModelWeak
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

	return nil
}

func (gpt *OpenAi) ConvertMessages(msgs []autog.ChatMessage) []OpenaiChatCompletionRequestMessage {
	reqmsgs := make([]OpenaiChatCompletionRequestMessage, len(msgs))
	for i, msg := range msgs {
		reqmsgs[i] = OpenaiChatCompletionRequestMessage{Role: msg.Role, Content: msg.Content}
	}
	return reqmsgs
}

func (gpt *OpenAi) JsonBodyReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return bytes.NewBuffer(nil), nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Failed encoding json: %w", err)
	}
	return bytes.NewBuffer(raw), nil
}

func (gpt *OpenAi) CreateChatCompletionRequest(weak, stream bool, msgs []autog.ChatMessage) *OpenaiChatCompletionRequest{
	model       := gpt.Model
	temperature := gpt.Temperature
	maxtokens   := gpt.MaxTokens

	if weak {
		model       = gpt.ModelWeak
		temperature = gpt.TemperatureWeak
		maxtokens   = gpt.MaxTokensWeak
	}

	if maxtokens > 0 {
		return &OpenaiChatCompletionRequest{
			Messages    : gpt.ConvertMessages(msgs),
			Model       : model,
			Temperature : float32(temperature) / float32(100),
			Stream      : stream,
			MaxTokens   : maxtokens,
		}
	}

	return &OpenaiChatCompletionRequest{
		Messages    : gpt.ConvertMessages(msgs),
		Model       : model,
		Temperature : float32(temperature) / float32(100),
		Stream      : stream,
	}
}

func (gpt *OpenAi) CreateHttpRequest(cxt context.Context, method, path string, payload interface{}) (*http.Request, error) {
	jsonBody, err := gpt.JsonBodyReader(payload)
	if err != nil {
		return nil, err
	}
	url := gpt.ApiBase + path
	req, err := http.NewRequestWithContext(cxt, method, url, jsonBody)
	if err != nil {
		return nil, err
	}
	if len(gpt.ApiOrg) > 0 {
		req.Header.Set("OpenAI-Organization", gpt.ApiOrg)
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gpt.ApiKey))
	return req, nil
}

func (gpt *OpenAi) CheckHttpResponseSuccess(httpRsp *http.Response) error {
	if httpRsp.StatusCode >= 200 && httpRsp.StatusCode < 300 {
		return nil
	}
	defer httpRsp.Body.Close()
	data, err := io.ReadAll(httpRsp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read from body: %w", err)
	}
	var result APIErrorResponse
	if err := json.Unmarshal(data, &result); err != nil {
		apiError := APIError{
			StatusCode: httpRsp.StatusCode,
			Type:       "Unexpected",
			Message:    string(data),
		}
		return apiError
	}
	result.Error.StatusCode = httpRsp.StatusCode
	return result.Error
}

func (gpt *OpenAi) GetHttpResponse(httpClient *http.Client, httpReq *http.Request) (*http.Response, error) {
	httpRsp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err := gpt.CheckHttpResponseSuccess(httpRsp); err != nil {
		return nil, err
	}

	return httpRsp, nil
}

func (gpt *OpenAi) GetHttpBodyObject(httpRsp *http.Response, response interface{}) error {
	defer httpRsp.Body.Close()
	if err := json.NewDecoder(httpRsp.Body).Decode(response); err != nil {
		return fmt.Errorf("Invalid json response: %w", err)
	}
	return nil
}

func (gpt *OpenAi) SendMessagesInner(cxt context.Context, msgs []autog.ChatMessage, weak bool) (autog.LLMStatus, autog.ChatMessage) {
	request := gpt.CreateChatCompletionRequest(weak, false, msgs)

	if gpt.Verbose >= VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/chat/completions", request)
	if err != nil {
		return autog.LLM_STATUS_BED_REQUEST, autog.ChatMessage{Role:autog.ASSISTANT, Content: err.Error()}
	}
	httpClient   := gpt.httpMain
	if weak {
		httpClient = gpt.httpWeak
	}
	httpRsp, err := gpt.GetHttpResponse(httpClient, httpReq)
	if err != nil {
		return autog.LLM_STATUS_BED_RESPONSE, autog.ChatMessage{Role:autog.ASSISTANT, Content: err.Error()}
	}
	response := OpenaiChatCompletionResponse{}
	if err := gpt.GetHttpBodyObject(httpRsp, &response); err != nil {
		return autog.LLM_STATUS_BED_MESSAGE, autog.ChatMessage{Role:autog.ASSISTANT, Content: err.Error()}
	}

	if gpt.Verbose >= VerboseShowReceiving {
		repstr, reperr := json.Marshal(response)
		if reperr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECV_FROM_LLVM:\n %s \n", repstr))
		}
	}

	revMsg := autog.ChatMessage{
		Role    : response.Choices[0].Message.Role,
		Content : response.Choices[0].Message.Content,
	}

	return autog.LLM_STATUS_OK, revMsg
}


func (gpt *OpenAi) SendMessagesStreamInner(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader, weak bool) (autog.LLMStatus, autog.ChatMessage) {
	request := gpt.CreateChatCompletionRequest(weak, true, msgs)

	if gpt.Verbose >= VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/chat/completions", request)
	if err != nil {
		return autog.LLM_STATUS_BED_REQUEST, autog.ChatMessage{Role:autog.ASSISTANT, Content: err.Error()}
	}
	httpClient   := gpt.httpMain
	if weak {
		httpClient = gpt.httpWeak
	}
	httpRsp, err := gpt.GetHttpResponse(httpClient, httpReq)
	if err != nil {
		return autog.LLM_STATUS_BED_RESPONSE, autog.ChatMessage{Role:autog.ASSISTANT, Content: err.Error()}
	}
	bufreader := bufio.NewReader(httpRsp.Body)
	defer httpRsp.Body.Close()

	var readErr error
	var line []byte
	var contentbuf *strings.Builder
	if reader != nil {
		contentbuf = reader.StreamStart()
	}
	for {
		line, readErr = bufreader.ReadBytes('\n')
		if readErr != nil {
			break
		}

		line = bytes.TrimSpace(line)

		if !bytes.HasPrefix(line, dataPrefix) {
			continue
		}

		line = bytes.TrimPrefix(line, dataPrefix)

		if bytes.HasPrefix(line, donePrefix) {
			break
		}

		response := OpenaiChatCompletionStreamResponse{}
		if err := json.Unmarshal(line, &response); err != nil {
			readErr = fmt.Errorf("Invalid json stream data: %v", err)
			break
		}

		delta := response.Choices[0].Delta.Content
		contentbuf.WriteString(delta)
		if reader != nil {
			reader.StreamDelta(contentbuf, delta)
		}
	}
	if reader != nil {
		reader.StreamEnd(contentbuf)
	}

	if readErr != nil {
		return autog.LLM_STATUS_BED_MESSAGE, autog.ChatMessage{Role:autog.ASSISTANT, Content: readErr.Error()}
	}

	revMsg := autog.ChatMessage{
		Role    : autog.ASSISTANT,
		Content : contentbuf.String(),
	}

	return autog.LLM_STATUS_OK, revMsg
}

func (gpt *OpenAi) CalcTokens(cxt context.Context, content string) int {
	// TODO: a fake tokens calc, will fix it with actual tokens calc logic
	return len(content) / 6
}

func (gpt *OpenAi) SendMessages(cxt context.Context, msgs []autog.ChatMessage) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesInner(cxt, msgs, false)

	if gpt.Verbose >= VerboseShowReceiving {
		msgstr, err := json.Marshal(msg)
		if err == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECEIVE_FROM_LLVM STATUS(%d)\n %s \n", status, msgstr))
		}
	}

	return status, msg
}

func (gpt *OpenAi) SendMessagesStream(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesStreamInner(cxt, msgs, reader, false)

	if gpt.Verbose >= VerboseShowReceiving {
		msgstr, err := json.Marshal(msg)
		if err == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECEIVE_FROM_LLVM STATUS(%d)\n %s \n", status, msgstr))
		}
	}

	return status, msg
}

func (gpt *OpenAi) CalcTokensByWeakModel(cxt context.Context, content string) int {
	// TODO: a fake tokens calc, will fix it with actual tokens calc logic
	return len(content) / 6
}

func (gpt *OpenAi) SendMessagesByWeakModel(cxt context.Context, msgs []autog.ChatMessage) (autog.LLMStatus, autog.ChatMessage) {
	return gpt.SendMessagesInner(cxt, msgs, true)
}

func (gpt *OpenAi) SendMessagesStreamByWeakModel(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader) (autog.LLMStatus, autog.ChatMessage) {
	return gpt.SendMessagesStreamInner(cxt, msgs, reader, true)
}