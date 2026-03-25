package gemini

type Config struct {
	ApiKey      string
	Model       string
	Temperature float64
	TopP        float64
}

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type GenerationConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"topP,omitempty"`
	ResponseMimeType string  `json:"response_mime_type,omitempty"`
	ResponseSchema   *Schema `json:"response_schema,omitempty"`
}

type Schema struct {
	Type        string            `json:"type"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Description string            `json:"description,omitempty"`
}

type RequestBody struct {
	SystemInstruction *Content          `json:"system_instruction,omitempty"`
	Contents          []Content         `json:"contents"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
}

type Response struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ContextResponse struct {
	Context        string `json:"context"`
	SourceLang     string `json:"sourceLang"`
	TargetLangCode string `json:"targetLangCode"`
	CleanName      string `json:"cleanName"`
}
