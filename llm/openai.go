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
	"encoding/base64"
	"encoding/binary"
	"autog"
)

const (
	defaultBaseURL        = "https://api.openai.com/v1"
	defaultVendor         = "openai"
	defaultModel          = "gpt-4-turbo-preview"
	defaultModelWeak      = "gpt-4-turbo-preview"
	defaultModelEmbed     = ""
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

type OpenaiUsage struct {
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
	Usage  OpenaiUsage                           `json:"usage"`
}

type OpenaiChatCompletionStreamResponse struct {
	ID      string                                     `json:"id"`
	Object  string                                     `json:"object"`
	Created int                                        `json:"created"`
	Model   string                                     `json:"model"`
	Choices []OpenaiChatCompletionStreamResponseChoice `json:"choices"`
	Usage   OpenaiUsage                                `json:"usage"`
}

type OpenaiEmbedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse is the response from a Create embeddings request.
type OpenaiEmbeddingResponse struct {
	Object string                             `json:"object"`
	Data   []OpenaiEmbedding                  `json:"data"`
	Model  string                             `json:"model"`
	Usage  OpenaiUsage                        `json:"usage"`
}

type base64String string

func (b base64String) Decode() ([]float32, error) {
	decodedData, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}

	const sizeOfFloat32 = 4
	floats := make([]float32, len(decodedData)/sizeOfFloat32)
	for i := 0; i < len(floats); i++ {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(decodedData[i*4 : (i+1)*4]))
	}

	return floats, nil
}

// Base64Embedding is a container for base64 encoded embeddings.
type OpenaiBase64Embedding struct {
	Object    string       `json:"object"`
	Embedding base64String `json:"embedding"`
	Index     int          `json:"index"`
}

// EmbeddingResponseBase64 is the response from a Create embeddings request with base64 encoding format.
type OpenaiEmbeddingResponseBase64 struct {
	Object string                  `json:"object"`
	Data   []OpenaiBase64Embedding `json:"data"`
	Model  string                  `json:"model"`
	Usage  OpenaiUsage             `json:"usage"`
}

// ToEmbeddingResponse converts an embeddingResponseBase64 to an EmbeddingResponse.
func (r *OpenaiEmbeddingResponseBase64) ToEmbeddingResponse() (OpenaiEmbeddingResponse, error) {
	data := make([]OpenaiEmbedding, len(r.Data))

	for i, base64Embedding := range r.Data {
		embedding, err := base64Embedding.Embedding.Decode()
		if err != nil {
			return OpenaiEmbeddingResponse{}, err
		}

		data[i] = OpenaiEmbedding{
			Object:    base64Embedding.Object,
			Embedding: embedding,
			Index:     base64Embedding.Index,
		}
	}

	return OpenaiEmbeddingResponse{
		Object: r.Object,
		Model:  r.Model,
		Data:   data,
		Usage:  r.Usage,
	}, nil
}


type OpenaiEmbeddingRequestConverter interface {
	// Needs to be of type EmbeddingRequestStrings or EmbeddingRequestTokens
	Convert() OpenaiEmbeddingRequest
}

// EmbeddingEncodingFormat is the format of the embeddings data.
// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
// If not specified OpenAI will use "float".
type OpenaiEmbeddingEncodingFormat string

const (
	OpenaiEmbeddingEncodingFormatFloat  OpenaiEmbeddingEncodingFormat = "float"
	OpenaiEmbeddingEncodingFormatBase64 OpenaiEmbeddingEncodingFormat = "base64"
)

type OpenaiEmbeddingRequest struct {
	Input          any                           `json:"input"`
	Model          string                        `json:"model"`
	User           string                        `json:"user"`
	EncodingFormat OpenaiEmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
}

func (r OpenaiEmbeddingRequest) Convert() OpenaiEmbeddingRequest {
	return r
}

// EmbeddingRequestStrings is the input to a create embeddings request with a slice of strings.
type OpenaiEmbeddingRequestStrings struct {
	// Input is a slice of strings for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input []string `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model string `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat OpenaiEmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
}

func (r OpenaiEmbeddingRequestStrings) Convert() OpenaiEmbeddingRequest {
	return OpenaiEmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
		Dimensions:     r.Dimensions,
	}
}


type OpenaiEmbeddingRequestTokens struct {
	// Input is a slice of slices of ints ([][]int) for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input [][]int `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model string `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat OpenaiEmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
}

func (r OpenaiEmbeddingRequestTokens) Convert() OpenaiEmbeddingRequest {
	return EmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
		Dimensions:     r.Dimensions,
	}
}

type OpenAi struct {
	ApiKey      string
	ApiBase     string
	ApiVendor   string
	ApiOrg      string
	Model       string
	ModelWeak   string
	ModelEmbed  string
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
		gpt.ApiBase = defaultBaseURL
	}
	if len(gpt.ApiVendor) <= 0 {
		gpt.ApiVendor = defaultVendor
	}
	if len(gpt.Model) <= 0 {
		gpt.Model = defaultModel
	}
	if len(gpt.ModelWeak) <= 0 {
		gpt.ModelWeak = defaultModelWeak
	}
	if len(gpt.ModelEmbed) <= 0 {
		gpt.ModelEmbed = defaultModelEmbed
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

	if gpt.Verbose >= autog.VerboseShowSending {
		reqstr, reqerr := json.Marshal(request)
		if reqerr == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("SEND_TO_LLVM:\n %s \n", reqstr))
		}
	}

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/chat/completions", request)
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
	response := OpenaiChatCompletionResponse{}
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
		Role    : response.Choices[0].Message.Role,
		Content : response.Choices[0].Message.Content,
	}

	return autog.LLM_STATUS_OK, revMsg
}


func (gpt *OpenAi) SendMessagesStreamInner(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader, weak bool) (autog.LLMStatus, autog.ChatMessage) {
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

	httpReq, err := gpt.CreateHttpRequest(cxt, "POST", "/chat/completions", request)
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
		if contentbuf != nil {
			contentbuf.WriteString(delta)
		}
		if reader != nil {
			reader.StreamDelta(contentbuf, delta)
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

func (gpt *OpenAi) CalcTokens(cxt context.Context, content string) int {
	// TODO: a fake tokens calc, will fix it with actual tokens calc logic
	return len(content) / 6
}

func (gpt *OpenAi) SendMessages(cxt context.Context, msgs []autog.ChatMessage) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesInner(cxt, msgs, false)

	if gpt.Verbose >= autog.VerboseShowReceiving {
		msgstr, err := json.Marshal(msg)
		if err == nil && gpt.VerboseLog != nil {
			gpt.VerboseLog(fmt.Sprintf("RECEIVE_FROM_LLVM STATUS(%d)\n %s \n", status, msgstr))
		}
	}

	return status, msg
}

func (gpt *OpenAi) SendMessagesStream(cxt context.Context, msgs []autog.ChatMessage, reader autog.StreamReader) (autog.LLMStatus, autog.ChatMessage) {
	status, msg := gpt.SendMessagesStreamInner(cxt, msgs, reader, false)

	if gpt.Verbose >= autog.VerboseShowReceiving {
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

func (gpt *OpenAi) Embedding(cxt context.Context, texts []string) (autog.Embedding, error) {
	var embed Embedding
	embeddingReq := openai.EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: openai.AdaSearchQuery,
	}
	baseReq := embeddingReq.Convert()
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/embeddings", string(baseReq.Model)), withBody(baseReq))
	if err != nil {
		return
	}

	if baseReq.EncodingFormat != EmbeddingEncodingFormatBase64 {
		err = c.sendRequest(req, &res)
		return
	}

	base64Response := &EmbeddingResponseBase64{}
	err = c.sendRequest(req, base64Response)
	if err != nil {
		return
	}

	res, err = base64Response.ToEmbeddingResponse()
	return embed, nil
}