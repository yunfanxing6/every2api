<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.configGuide.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('profile.configGuide.description') }}
      </p>
    </div>

    <div class="space-y-6 px-6 py-6">
      <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
        {{ t('profile.configGuide.intro') }}
      </p>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
          <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
            {{ t('profile.configGuide.endpointTitle') }}
          </p>
          <code class="mt-2 block break-all rounded-xl bg-white px-3 py-2 font-mono text-sm text-gray-800 dark:bg-dark-700 dark:text-gray-100">
            {{ apiBaseUrl }}
          </code>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.endpointDescription') }}
          </p>
        </div>

        <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
          <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
            {{ t('profile.configGuide.authHeaderTitle') }}
          </p>
          <code class="mt-2 block break-all rounded-xl bg-white px-3 py-2 font-mono text-sm text-gray-800 dark:bg-dark-700 dark:text-gray-100">
            {{ authHeader }}
          </code>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.authHeaderDescription') }}
          </p>
        </div>

        <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
          <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
            {{ t('profile.configGuide.apiKeyTitle') }}
          </p>
          <code class="mt-2 block break-all rounded-xl bg-white px-3 py-2 font-mono text-sm text-gray-800 dark:bg-dark-700 dark:text-gray-100">
            {{ placeholderApiKey }}
          </code>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.apiKeyDescription') }}
          </p>
        </div>

        <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
          <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
            {{ t('profile.configGuide.modelReferenceTitle') }}
          </p>
          <div class="mt-2 flex flex-wrap gap-2">
            <span
              v-for="model in quickModels"
              :key="model"
              class="rounded-full bg-white px-3 py-1 font-mono text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-100"
            >
              {{ model }}
            </span>
          </div>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.modelReferenceDescription') }}
          </p>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-200 p-5 dark:border-dark-700">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('profile.configGuide.stepsTitle') }}
        </h3>
        <ol class="mt-4 space-y-3">
          <li
            v-for="(step, index) in steps"
            :key="step"
            class="flex items-start gap-3 text-sm text-gray-600 dark:text-gray-300"
          >
            <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ index + 1 }}
            </span>
            <span class="leading-6">{{ step }}</span>
          </li>
        </ol>
      </div>

      <div class="flex flex-wrap gap-2">
        <button
          v-for="client in clientTabs"
          :key="client.key"
          type="button"
          class="rounded-xl px-4 py-2 text-sm font-medium transition-all"
          :class="activeClient === client.key
            ? 'bg-primary-600 text-white shadow-lg shadow-primary-600/20'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700'"
          @click="activeClient = client.key"
        >
          {{ client.label }}
        </button>
      </div>

      <div class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
        <div class="border-b border-gray-200 bg-gray-900 px-4 py-3 dark:border-dark-700">
          <p class="text-sm font-semibold text-white">
            {{ currentClient.label }}
          </p>
          <p class="mt-1 text-xs leading-5 text-gray-400">
            {{ currentClient.description }}
          </p>
        </div>

        <div v-if="activeClient === 'opencode'" class="border-b border-gray-200 bg-gray-50 px-4 py-4 dark:border-dark-700 dark:bg-dark-900/30">
          <div class="flex flex-wrap gap-2">
            <button
              v-for="provider in providerTabs"
              :key="provider.key"
              type="button"
              class="rounded-lg px-3 py-2 text-sm font-medium transition-all"
              :class="activeProvider === provider.key
                ? 'bg-primary-600 text-white shadow-lg shadow-primary-600/20'
                : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700'"
              @click="activeProvider = provider.key"
            >
              {{ provider.label }}
            </button>
          </div>
          <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-gray-400">
            {{ currentProvider.description }}
          </p>
        </div>

        <div class="space-y-4 p-4">
          <div
            v-for="file in currentFiles"
            :key="file.id"
            class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700"
          >
            <div class="flex flex-col gap-3 border-b border-gray-200 bg-gray-900 px-4 py-3 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
              <span class="text-xs font-mono text-gray-300">{{ file.path }}</span>
              <button
                type="button"
                class="rounded-lg bg-white/10 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-white/20"
                @click="copyContent(file.id, file.content)"
              >
                {{ copiedItem === file.id ? t('common.copied') : t('profile.configGuide.copy') }}
              </button>
            </div>
            <pre class="overflow-x-auto bg-gray-950 p-4 text-sm leading-6 text-gray-100"><code>{{ file.content }}</code></pre>
          </div>
        </div>
      </div>

      <div class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
        <div class="flex flex-col gap-3 border-b border-gray-200 bg-gray-900 px-4 py-3 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p class="text-sm font-semibold text-white">
              {{ t('profile.configGuide.aiPromptTitle') }}
            </p>
            <p class="mt-1 text-xs leading-5 text-gray-400">
              {{ t('profile.configGuide.aiPromptDescription') }}
            </p>
          </div>

          <button
            type="button"
            class="rounded-lg bg-white/10 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-white/20"
            @click="copyContent(`prompt-${activeClient}`, currentPrompt)"
          >
            {{ copiedItem === `prompt-${activeClient}` ? t('common.copied') : t('profile.configGuide.aiPromptCopy') }}
          </button>
        </div>

        <pre class="overflow-x-auto bg-gray-950 p-4 text-sm leading-6 text-gray-100"><code>{{ currentPrompt }}</code></pre>
      </div>

      <div class="rounded-2xl border border-primary-200 bg-primary-50 p-5 dark:border-primary-800/60 dark:bg-primary-900/20">
        <p class="text-sm font-semibold text-primary-900 dark:text-primary-200">
          {{ t('profile.configGuide.modelsTitle') }}
        </p>
        <div class="mt-3 flex flex-wrap gap-2">
          <span
            v-for="model in currentModels"
            :key="model"
            class="rounded-full border border-primary-200 bg-white px-3 py-1 text-xs font-medium text-primary-700 dark:border-primary-700 dark:bg-primary-950/40 dark:text-primary-200"
          >
            {{ model }}
          </span>
        </div>
      </div>

      <div v-if="currentClient.note" class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/30">
        <p class="text-sm leading-6 text-gray-700 dark:text-gray-300">
          {{ currentClient.note }}
        </p>
      </div>

      <div class="rounded-2xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800/60 dark:bg-amber-900/20">
        <p class="text-sm leading-6 text-amber-800 dark:text-amber-200">
          {{ t('profile.configGuide.mergedHint') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'

type GuideProvider = 'all' | 'openai' | 'qwen' | 'grok'
type GuideClient = 'opencode' | 'codex' | 'claude' | 'curl' | 'python' | 'javascript'

interface ProviderTab {
  key: GuideProvider
  label: string
  description: string
  models: string[]
  snippet: string
}

interface ClientTab {
  key: GuideClient
  label: string
  description: string
  note?: string
  models: string[]
}

interface ExampleFile {
  id: string
  path: string
  content: string
}

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const configPath = '~/.config/opencode/opencode.jsonc'
const placeholderApiKey = 'YOUR_API_KEY'
const authHeader = 'Authorization: Bearer YOUR_API_KEY'
const quickModels = ['gpt-5.4', 'qwen3.6-plus', 'grok-4.20-0309']
const genericModels = ['gpt-5.4', 'qwen3.6-plus', 'grok-4.20-0309', 'gpt-5.3-codex', 'qwen3.6-plus:thinking', 'grok-4.20-0309-reasoning']

const activeClient = ref<GuideClient>('opencode')
const activeProvider = ref<GuideProvider>('all')
const copiedItem = ref<string | null>(null)

const apiBaseUrl = computed(() => {
  if (typeof window === 'undefined') {
    return '/v1'
  }

  return new URL('/v1', window.location.origin).toString().replace(/\/$/, '')
})

const steps = computed(() => [
  t('profile.configGuide.steps.createKey'),
  t('profile.configGuide.steps.chooseClient'),
  t('profile.configGuide.steps.copyExample'),
  t('profile.configGuide.steps.replaceValues'),
  t('profile.configGuide.steps.testRequest')
])

const buildSnippet = (provider: Record<string, unknown>) => JSON.stringify({
  $schema: 'https://opencode.ai/config.json',
  provider
}, null, 2)

const openAIVariants = {
  low: {},
  medium: {},
  high: {},
  xhigh: {}
}

const openaiProviderConfig = computed(() => ({
  openai: {
    models: {
      'gpt-5.4': {
        name: 'GPT-5.4',
        limit: {
          context: 1050000,
          output: 128000
        },
        options: {
          store: false
        },
        variants: openAIVariants
      },
      'gpt-5.4-2026-03-05': {
        name: 'GPT-5.4 2026-03-05',
        limit: {
          context: 1050000,
          output: 128000
        },
        options: {
          store: false
        },
        variants: openAIVariants
      },
      'gpt-5.3-codex': {
        name: 'GPT-5.3 Codex',
        limit: {
          context: 400000,
          output: 128000
        },
        options: {
          store: false
        },
        variants: openAIVariants
      },
      'gpt-5.3-codex-spark': {
        name: 'GPT-5.3 Codex Spark',
        limit: {
          context: 128000,
          output: 32000
        },
        options: {
          store: false
        },
        variants: openAIVariants
      }
    },
    options: {
      baseURL: apiBaseUrl.value,
      apiKey: placeholderApiKey
    }
  }
}))

const qwenProviderConfig = computed(() => ({
  qwen: {
    name: 'Qwen 3.6 Plus',
    npm: '@ai-sdk/openai-compatible',
    models: {
      'qwen3.6-plus': {
        name: 'Qwen 3.6 Plus (Auto)',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.6-plus:auto': {
        name: 'Qwen 3.6 Plus (Auto)',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.6-plus:fast': {
        name: 'Qwen 3.6 Plus (Fast)',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.6-plus:thinking': {
        name: 'Qwen 3.6 Plus (Thinking)',
        attachment: false,
        reasoning: true,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.5-plus': {
        name: 'Qwen 3.5 Plus',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.5-flash': {
        name: 'Qwen 3.5 Flash',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'qwen3.5-omni-plus': {
        name: 'Qwen 3.5 Omni Plus',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 1000000,
          output: 65536
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      }
    },
    options: {
      baseURL: apiBaseUrl.value,
      apiKey: placeholderApiKey,
      timeout: 600000
    }
  }
}))

const grokProviderConfig = computed(() => ({
  grok: {
    name: 'Grok2API',
    npm: '@ai-sdk/openai-compatible',
    models: {
      'grok-4.20-0309-non-reasoning': {
        name: 'Grok 4.20 0309 Non-Reasoning (Fast)',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 256000,
          output: 64000
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'grok-4.20-0309': {
        name: 'Grok 4.20 0309 (Auto)',
        attachment: false,
        reasoning: false,
        tool_call: true,
        temperature: true,
        limit: {
          context: 256000,
          output: 64000
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'grok-4.20-0309-reasoning': {
        name: 'Grok 4.20 0309 Reasoning (Expert)',
        attachment: false,
        reasoning: true,
        tool_call: true,
        temperature: true,
        limit: {
          context: 256000,
          output: 64000
        },
        modalities: {
          input: ['text'],
          output: ['text']
        }
      },
      'grok-imagine-image-lite': {
        name: 'Grok Imagine Image Lite',
        attachment: false,
        reasoning: false,
        tool_call: false,
        temperature: false,
        limit: {
          context: 32000,
          output: 4096
        },
        modalities: {
          input: ['text'],
          output: ['image']
        }
      }
    },
    options: {
      baseURL: apiBaseUrl.value,
      apiKey: placeholderApiKey,
      timeout: 600000
    }
  }
}))

const providerTabs = computed<ProviderTab[]>(() => {
  const fullConfig = {
    ...openaiProviderConfig.value,
    ...qwenProviderConfig.value,
    ...grokProviderConfig.value
  }

  return [
    {
      key: 'all',
      label: t('profile.configGuide.providers.all.label'),
      description: t('profile.configGuide.providers.all.description'),
      models: ['gpt-5.4', 'gpt-5.3-codex', 'qwen3.6-plus', 'qwen3.6-plus:thinking', 'grok-4.20-0309', 'grok-imagine-image-lite'],
      snippet: buildSnippet(fullConfig)
    },
    {
      key: 'openai',
      label: t('profile.configGuide.providers.openai.label'),
      description: t('profile.configGuide.providers.openai.description'),
      models: ['gpt-5.4', 'gpt-5.4-2026-03-05', 'gpt-5.3-codex', 'gpt-5.3-codex-spark'],
      snippet: buildSnippet(openaiProviderConfig.value)
    },
    {
      key: 'qwen',
      label: t('profile.configGuide.providers.qwen.label'),
      description: t('profile.configGuide.providers.qwen.description'),
      models: ['qwen3.6-plus', 'qwen3.6-plus:auto', 'qwen3.6-plus:fast', 'qwen3.6-plus:thinking', 'qwen3.5-plus', 'qwen3.5-flash', 'qwen3.5-omni-plus'],
      snippet: buildSnippet(qwenProviderConfig.value)
    },
    {
      key: 'grok',
      label: t('profile.configGuide.providers.grok.label'),
      description: t('profile.configGuide.providers.grok.description'),
      models: ['grok-4.20-0309-non-reasoning', 'grok-4.20-0309', 'grok-4.20-0309-reasoning', 'grok-imagine-image-lite'],
      snippet: buildSnippet(grokProviderConfig.value)
    }
  ]
})

const currentProvider = computed(() =>
  providerTabs.value.find((tab) => tab.key === activeProvider.value) ?? providerTabs.value[0]
)

const clientTabs = computed<ClientTab[]>(() => [
  {
    key: 'opencode',
    label: t('profile.configGuide.clients.opencode.label'),
    description: t('profile.configGuide.clients.opencode.description'),
    note: t('profile.configGuide.clients.opencode.note'),
    models: currentProvider.value.models
  },
  {
    key: 'codex',
    label: t('profile.configGuide.clients.codex.label'),
    description: t('profile.configGuide.clients.codex.description'),
    note: t('profile.configGuide.clients.codex.note'),
    models: genericModels
  },
  {
    key: 'claude',
    label: t('profile.configGuide.clients.claude.label'),
    description: t('profile.configGuide.clients.claude.description'),
    note: t('profile.configGuide.clients.claude.note'),
    models: genericModels
  },
  {
    key: 'curl',
    label: t('profile.configGuide.clients.curl.label'),
    description: t('profile.configGuide.clients.curl.description'),
    note: t('profile.configGuide.clients.curl.note'),
    models: genericModels
  },
  {
    key: 'python',
    label: t('profile.configGuide.clients.python.label'),
    description: t('profile.configGuide.clients.python.description'),
    note: t('profile.configGuide.clients.python.note'),
    models: genericModels
  },
  {
    key: 'javascript',
    label: t('profile.configGuide.clients.javascript.label'),
    description: t('profile.configGuide.clients.javascript.description'),
    note: t('profile.configGuide.clients.javascript.note'),
    models: genericModels
  }
])

const currentClient = computed(() =>
  clientTabs.value.find((tab) => tab.key === activeClient.value) ?? clientTabs.value[0]
)

const currentFiles = computed<ExampleFile[]>(() => {
  if (activeClient.value === 'opencode') {
    return [
      {
        id: `opencode-${activeProvider.value}`,
        path: configPath,
        content: currentProvider.value.snippet
      }
    ]
  }

  if (activeClient.value === 'codex') {
    return [
      {
        id: 'codex-config',
        path: '~/.codex/config.toml',
        content: `model_provider = "OpenAI"
model = "gpt-5.4"
review_model = "gpt-5.4"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${apiBaseUrl.value}"
wire_api = "responses"
requires_openai_auth = true`
      },
      {
        id: 'codex-auth',
        path: '~/.codex/auth.json',
        content: `{
  "OPENAI_API_KEY": "${placeholderApiKey}"
}`
      }
    ]
  }

  if (activeClient.value === 'claude') {
    return [
      {
        id: 'claude-terminal',
        path: 'Terminal',
        content: `export ANTHROPIC_BASE_URL="${apiBaseUrl.value}"
export ANTHROPIC_AUTH_TOKEN="${placeholderApiKey}"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
      },
      {
        id: 'claude-settings',
        path: '~/.claude/settings.json',
        content: `{
  "env": {
    "ANTHROPIC_BASE_URL": "${apiBaseUrl.value}",
    "ANTHROPIC_AUTH_TOKEN": "${placeholderApiKey}",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  }
}`
      }
    ]
  }

  if (activeClient.value === 'curl') {
    return [
      {
        id: 'curl-example',
        path: 'Terminal',
        content: `curl "${apiBaseUrl.value}/chat/completions" \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${placeholderApiKey}" \\
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello"}
    ],
    "stream": false
  }'`
      }
    ]
  }

  if (activeClient.value === 'python') {
    return [
      {
        id: 'python-example',
        path: 'example.py',
        content: `from openai import OpenAI

client = OpenAI(
    base_url="${apiBaseUrl.value}",
    api_key="${placeholderApiKey}",
)

response = client.chat.completions.create(
    model="gpt-5.4",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello"},
    ],
)

print(response.choices[0].message.content)`
      }
    ]
  }

  return [
    {
      id: 'javascript-example',
      path: 'example.mjs',
      content: `import OpenAI from "openai"

const client = new OpenAI({
  baseURL: "${apiBaseUrl.value}",
  apiKey: "${placeholderApiKey}",
})

const response = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [
    { role: "system", content: "You are a helpful assistant." },
    { role: "user", content: "Hello" }
  ]
})

console.log(response.choices[0].message.content)`
    }
  ]
})

const currentModels = computed(() =>
  activeClient.value === 'opencode' ? currentProvider.value.models : currentClient.value.models
)

const currentPrompt = computed(() => {
  const promptFiles = currentFiles.value
    .map((file) => `Path: ${file.path}\n\n\`\`\`\n${file.content}\n\`\`\``)
    .join('\n\n')

  const mergeRule =
    activeClient.value === 'opencode'
      ? t('profile.configGuide.promptRules.mergeProviders')
      : t('profile.configGuide.promptRules.keepExactFiles')

  return `${t('profile.configGuide.promptHeader', { client: currentClient.value.label })}

- Base URL: ${apiBaseUrl.value}
- API Key placeholder: ${placeholderApiKey}
- Example models: ${quickModels.join(' / ')}

${t('profile.configGuide.promptReferenceTitle')}
1. ${mergeRule}
2. ${t('profile.configGuide.promptRules.replaceKey')}
3. ${t('profile.configGuide.promptRules.outputComplete')}
4. ${t('profile.configGuide.promptRules.noPseudo')}

${promptFiles}`
})

const copyContent = async (key: string, content: string) => {
  const ok = await copyToClipboard(content)
  if (!ok) {
    return
  }

  copiedItem.value = key
  window.setTimeout(() => {
    if (copiedItem.value === key) {
      copiedItem.value = null
    }
  }, 2000)
}
</script>
