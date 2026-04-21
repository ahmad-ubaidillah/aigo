# Providers

Aigo supports 60+ AI providers with OpenAI-compatible APIs.

## Supported Providers

### Major Global Providers

| Provider | Models | API Key |
|----------|-------|--------|
| **OpenAI** | GPT-4o, GPT-4o Mini, o3 | [platform.openai.com](https://platform.openai.com/api-keys) |
| **Anthropic** | Claude 3.5 Sonnet, Opus | [console.anthropic.com](https://console.anthropic.com/settings/keys) |
| **Google** | Gemini 2.0 Flash | [aistudio.google.com](https://aistudio.google.com/apikey) |
| **DeepSeek** | DeepSeek-V3, R1 | [platform.deepseek.com](https://platform.deepseek.com/api_keys) |
| **xAI** | Grok-2 | [console.x.ai](https://console.x.ai) |
| **Mistral** | Mistral Large, Codestral | [console.mistral.ai](https://console.mistral.ai/api-keys) |
| **Cohere** | Command R Plus | [dashboard.cohere.ai](https://dashboard.cohere.ai) |

### Chinese Providers

| Provider | Region | API Key |
|----------|--------|---------|
| **Moonshot (Kimi)** | China | [platform.moonshot.cn](https://platform.moonshot.cn/console/api-keys) |
| **GLM/Zhipu** | China | [open.bigmodel.cn](https://open.bigmodel.cn/usercenter/proj-mgmt/apikeys) |
| **MiniMax** | China | [platform.minimaxi.com](https://platform.minimaxi.com/user-center/basic-information/interface-key) |
| **Qwen** | China | [dashscope.console.aliyun.com](https://dashscope.console.aliyun.com/apiKey) |
| **Xiaomi MiMo** | China | [platform.xiaomimimo.com](https://platform.xiaomimimo.com/) |
| **Baidu Ernie** | China | [login.baidu.com](https://login.baidu.com) |
| **iFlytek Spark** | China | [spark-api.xf-yun.com](https://spark-api.xf-yun.com) |
| **Volcengine** | China | [console.volcengine.com](https://console.volcengine.com) |

### Aggregators & Gateway

| Provider | Models | API Key |
|----------|--------|---------|
| **OpenRouter** | 200+ models | [openrouter.ai/keys](https://openrouter.ai/keys) |
| **Together AI** | Llama, Mistral | [together.ai](https://together.ai) |
| **Fireworks AI** | Firefunction | [fireworks.ai](https://fireworks.ai) |
| **DeepInfra** | Llama, Mistral | [deepinfra.com](https://deepinfra.com) |

### Cloud & Enterprise

| Provider | Notes | API Key |
|----------|-------|---------|
| **Azure OpenAI** | Enterprise Azure | [portal.azure.com](https://portal.azure.com/) |
| **AWS Bedrock** | Claude, Llama on AWS | [console.aws.amazon.com](https://console.aws.amazon.com/bedrock) |
| **Google Cloud** | Enterprise Gemini | [console.cloud.google.com](https://console.cloud.google.com/) |

### Fast Inference

| Provider | Notes | API Key |
|----------|-------|---------|
| **Groq** | Fast Llama | [console.groq.com](https://console.groq.com/keys) |
| **Cerebras** | Fast inference | [cloud.cerebras.ai](https://cloud.cerebras.ai/) |
| **NVIDIA** | NIM models | [build.nvidia.com](https://build.nvidia.com/) |

### Local

| Provider | Notes | API Key |
|----------|-------|---------|
| **Ollama** | Local models | [ollama.com](https://ollama.com/) |
| **vLLM** | Self-hosted | http://localhost:8000 |

### Other Providers

| Provider | Models |
|----------|--------|
| **HuggingFace** | BART, Llama |
| **Perplexity** | Sonar |
| **Cloudflare** | Llama on Workers AI |
| **SambaNova** | Meta-Llama |
| **Neura** | Chat models |
| **Arcee** | Diva |
| **Novita AI** | Various |
| **AI Gateway** | Proxy |
| **KiloCode** | Gateway |
| **OpenCode Zen/Go** | Custom |

## Configuration

### Using config.yaml

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"
```

### Using Environment Variables

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# DeepSeek
export DEEPSEEK_API_KEY="sk-..."

# Google
export GOOGLE_API_KEY="AIza..."
```

## Smart Routing

Route simple queries to cheap models:

```yaml
smart_routing:
  enabled: true
  max_simple_chars: 160
  cheap_model:
    provider: "openai"
    model: "gpt-4o-mini"
```

## Default Models

| Provider | Default Model |
|-----------|-----------------|
| openai | gpt-4o |
| anthropic | claude-3-5-sonnet-20241022 |
| google | gemini-2.0-flash |
| deepseek | deepseek-chat |
| xai | grok-2 |
| mistral | mistral-small-latest |
| groq | llama-3.1-70b-versatile |
| perplexity | llama-3.1-sonar-small-online |
| qwen | qwen2.5-coder-32b |

## Adding Custom Providers

Edit `internal/providers/providers.go`:

```go
var ProviderRegistry = map[string]ProviderInfo{
  "myprovider": {
    BaseURL: "https://api.myprovider.com/v1",
    APIKeyEnv: "MYPROVIDER_API_KEY",
    APIMode: "openai",
  },
}
```

## Testing Providers

```bash
# Test OpenAI
OPENAI_API_KEY="sk-..." ./aigo "Hello"

# Test Anthropic
ANTHROPIC_API_KEY="sk-ant-..." ./aigo --provider anthropic "Hello"

# Test DeepSeek
DEEPSEEK_API_KEY="sk-..." ./aigo --provider deepseek "Hello"
```

---

For configuration details, see [Installation](./installation.md)