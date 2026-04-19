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

      <div class="grid gap-4 md:grid-cols-3">
        <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
          <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
            {{ t('profile.configGuide.configPathTitle') }}
          </p>
          <code class="mt-2 block break-all rounded-xl bg-white px-3 py-2 font-mono text-sm text-gray-800 dark:bg-dark-700 dark:text-gray-100">
            {{ configPath }}
          </code>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.configPathDescription') }}
          </p>
        </div>

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
            {{ t('profile.configGuide.apiKeyTitle') }}
          </p>
          <code class="mt-2 block break-all rounded-xl bg-white px-3 py-2 font-mono text-sm text-gray-800 dark:bg-dark-700 dark:text-gray-100">
            {{ placeholderApiKey }}
          </code>
          <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('profile.configGuide.apiKeyDescription') }}
          </p>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-200 p-5 dark:border-dark-700">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('profile.configGuide.stepsTitle') }}
        </h3>
        <ol class="mt-4 space-y-3">
          <li v-for="(step, index) in steps" :key="step" class="flex items-start gap-3 text-sm text-gray-600 dark:text-gray-300">
            <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ index + 1 }}
            </span>
            <span class="leading-6">{{ step }}</span>
          </li>
        </ol>
      </div>

      <div class="flex flex-wrap gap-2">
        <button
          v-for="tab in providerTabs"
          :key="tab.key"
          type="button"
          class="rounded-xl px-4 py-2 text-sm font-medium transition-all"
          :class="activeProvider === tab.key
            ? 'bg-primary-600 text-white shadow-lg shadow-primary-600/20'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700'"
          @click="activeProvider = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
        <div class="flex flex-col gap-3 border-b border-gray-200 bg-gray-900 px-4 py-3 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p class="text-sm font-semibold text-white">
              {{ currentProvider.label }}
            </p>
            <p class="mt-1 text-xs leading-5 text-gray-400">
              {{ currentProvider.description }}
            </p>
          </div>

          <button
            type="button"
            class="rounded-lg bg-white/10 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-white/20"
            @click="copyCurrentSnippet"
          >
            {{ copiedProvider === activeProvider ? t('common.copied') : t('profile.configGuide.copy') }}
          </button>
        </div>

        <pre class="overflow-x-auto bg-gray-950 p-4 text-sm leading-6 text-gray-100"><code>{{ currentProvider.snippet }}</code></pre>
      </div>

      <div class="rounded-2xl border border-primary-200 bg-primary-50 p-5 dark:border-primary-800/60 dark:bg-primary-900/20">
        <p class="text-sm font-semibold text-primary-900 dark:text-primary-200">
          {{ t('profile.configGuide.modelsTitle') }}
        </p>
        <div class="mt-3 flex flex-wrap gap-2">
          <span
            v-for="model in currentProvider.models"
            :key="model"
            class="rounded-full border border-primary-200 bg-white px-3 py-1 text-xs font-medium text-primary-700 dark:border-primary-700 dark:bg-primary-950/40 dark:text-primary-200"
          >
            {{ model }}
          </span>
        </div>
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

interface ProviderTab {
  key: GuideProvider
  label: string
  description: string
  models: string[]
  snippet: string
}

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const configPath = '~/.config/opencode/opencode.jsonc'
const placeholderApiKey = 'YOUR_API_KEY'
const activeProvider = ref<GuideProvider>('all')
const copiedProvider = ref<GuideProvider | null>(null)

const apiBaseUrl = computed(() => {
  if (typeof window === 'undefined') {
    return '/v1'
  }

  return new URL('/v1', window.location.origin).toString().replace(/\/$/, '')
})

const steps = computed(() => [
  t('profile.configGuide.steps.openFile'),
  t('profile.configGuide.steps.copySnippet'),
  t('profile.configGuide.steps.replaceKey'),
  t('profile.configGuide.steps.selectProvider')
])

const buildSnippet = (provider: Record<string, unknown>) => JSON.stringify({
  $schema: 'https://opencode.ai/config.json',
  provider
}, null, 2)

const openaiProviderConfig = computed(() => ({
  openai: {
    models: {
      'gpt-5.4': {
        name: 'gpt-5.4'
      },
      'gpt-5.4-2026-03-05': {
        name: 'gpt-5.4-2026-03-05'
      },
      'gpt-5.3-codex': {
        name: 'gpt-5.3-codex'
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
    name: 'Qwen',
    npm: '@ai-sdk/openai-compatible',
    models: {
      'qwen3.6-plus': {
        name: 'Qwen 3.6 Plus'
      },
      'qwen3.5-plus': {
        name: 'Qwen 3.5 Plus'
      },
      'qwen3.5-flash': {
        name: 'Qwen 3.5 Flash'
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
    name: 'Grok',
    npm: '@ai-sdk/openai-compatible',
    models: {
      'grok-4.20-0309': {
        name: 'Grok 4.20 0309'
      },
      'grok-4.20-0309-reasoning': {
        name: 'Grok 4.20 0309 Reasoning'
      },
      'grok-imagine-image-lite': {
        name: 'Grok Imagine Image Lite'
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
      models: ['gpt-5.4', 'qwen3.6-plus', 'grok-4.20-0309'],
      snippet: buildSnippet(fullConfig)
    },
    {
      key: 'openai',
      label: t('profile.configGuide.providers.openai.label'),
      description: t('profile.configGuide.providers.openai.description'),
      models: ['gpt-5.4', 'gpt-5.4-2026-03-05', 'gpt-5.3-codex'],
      snippet: buildSnippet(openaiProviderConfig.value)
    },
    {
      key: 'qwen',
      label: t('profile.configGuide.providers.qwen.label'),
      description: t('profile.configGuide.providers.qwen.description'),
      models: ['qwen3.6-plus', 'qwen3.5-plus', 'qwen3.5-flash'],
      snippet: buildSnippet(qwenProviderConfig.value)
    },
    {
      key: 'grok',
      label: t('profile.configGuide.providers.grok.label'),
      description: t('profile.configGuide.providers.grok.description'),
      models: ['grok-4.20-0309', 'grok-4.20-0309-reasoning', 'grok-imagine-image-lite'],
      snippet: buildSnippet(grokProviderConfig.value)
    }
  ]
})

const currentProvider = computed(() =>
  providerTabs.value.find((tab) => tab.key === activeProvider.value) ?? providerTabs.value[0]
)

const copyCurrentSnippet = async () => {
  const ok = await copyToClipboard(currentProvider.value.snippet)
  if (!ok) {
    return
  }

  const copiedKey = activeProvider.value
  copiedProvider.value = copiedKey
  window.setTimeout(() => {
    if (copiedProvider.value === copiedKey) {
      copiedProvider.value = null
    }
  }, 2000)
}
</script>
