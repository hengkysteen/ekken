import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  fetchConversations,
  fetchConversationById,
  chatWithProvider,
  fetchAssistantJob,
  fetchAssistantJobs,
  openAssistantStream,
  stopChat,
  type Conversation
} from '../api'
import { Storage, StorageKeys } from '@shared/utils/storage'

export interface Message {
  role: 'user' | 'assistant'
  content: string
  thinking?: string
  provider?: string
  model?: string
  agent?: string
  done?: boolean
  state?: string
}

export const useAssistantStore = defineStore('assistant', () => {
  // State
  const messages = ref<Message[]>([])
  const conversationList = ref<Conversation[]>([])
  const providerList = ref<any[]>([])
  const activeConversationId = ref('')
  const runningByConversation = ref<Record<string, boolean>>({})
  const pendingNewChatSending = ref(false)
  const isSending = computed(() => {
    if (!activeConversationId.value) return pendingNewChatSending.value
    return !!runningByConversation.value[activeConversationId.value]
  })
  const initialized = ref(false)
  const activeProvider = ref('')
  const activeModel = ref('')
  const agents = ref<any[]>([])
  const selectedAgent = ref('chat')
  const selectedThinking = ref('low')
  const displayedPrefix = ref('')
  const displayedMain = ref('')
  const enableTransition = ref(false)
  const isTyping = ref(false)
  const sidebarCollapsed = ref(Storage.get<boolean>(StorageKeys.ASSISTANT_SIDEBAR_COLLAPSED) ?? false)
  const nextGreetingIdx = ref(Storage.get<number>(StorageKeys.ASSISTANT_NEXT_GREETING) ?? 0)
  let streamController: AbortController | null = null
  let streamConversationId = ''

  const greetings = [
    { prefix: "Ahoy, Captain! 🏴‍☠️", main: "Give the word and I'll burn the seas for ya." },

    { prefix: "Hi!", main: "Let's craft some powerful **node plugins** to expand your system's capabilities. 🔌" },
    { prefix: "System Online.", main: "Ready to help you plan, architect, and bring your **modular ideas** to life. 🏗️" },
    { prefix: "Welcome back!", main: "Need help with **logic nodes** or mapping out a new project structure? ✨" },
    { prefix: "Hello there!", main: "My processors are ready. Let's optimize your **automation** and build something epic. 🚀" }
  ]

  // Getters
  const isNewChat = (chatId: string) => !chatId && messages.value.length === 0

  const getChatTitle = (chatId: string) => {
    if (!chatId) return 'Assistant'
    const chat = conversationList.value.find(c => String(c.id) === String(chatId))
    return chat?.title || 'Assistant'
  }

  const getProviderLogo = (providerId?: string) => {
    if (!providerId) return ''
    return providerList.value.find(p => p.provider_id === providerId)?.logo || ''
  }

  // Actions
  const fetchAllTitles = async () => {
    try {
      conversationList.value = await fetchConversations()
      await syncRunningJobs()
    } catch (e) {
      console.error('Failed to fetch conversations:', e)
    }
  }

  const syncRunningJobs = async () => {
    try {
      const jobs = await fetchAssistantJobs()
      const next: Record<string, boolean> = {}
      for (const job of jobs) {
        if (job.running) next[job.conversation_id] = true
      }
      runningByConversation.value = next
    } catch (e) {
      console.error('Failed to fetch assistant jobs:', e)
    }
  }

  const fetchProviderList = async () => {
    try {
      const res = await fetch('/api/assistant/providers')
      const json = await res.json()
      providerList.value = json.data || []
    } catch (e) {
      console.error('Failed to fetch providers:', e)
    }
  }

  const fetchAgents = async () => {
    try {
      const res = await fetch('/api/assistant/agents')
      const json = await res.json()
      if (json.data?.length) {
        agents.value = json.data
        const names = json.data.map((a: any) => a.name)
        if (!selectedAgent.value || !names.includes(selectedAgent.value)) {
          selectedAgent.value = names[0]
        }
      }
    } catch (e) {
      console.error('Failed to fetch agents:', e)
    }
  }

  const startTyping = async () => {
    if (isTyping.value) return
    isTyping.value = true;
    const g = greetings[nextGreetingIdx.value];

    // Increment for next time (Round Robin)
    nextGreetingIdx.value = (nextGreetingIdx.value + 1) % greetings.length;
    Storage.set(StorageKeys.ASSISTANT_NEXT_GREETING, nextGreetingIdx.value);

    displayedPrefix.value = '';
    displayedMain.value = '';

    for (const char of g.prefix) {
      displayedPrefix.value += char;
      await new Promise(r => setTimeout(r, 45));
    }

    await new Promise(r => setTimeout(r, 150));

    const segmenter = new Intl.Segmenter(undefined, { granularity: 'grapheme' });
    for (const { segment } of segmenter.segment(g.main)) {
      displayedMain.value += segment;
      await new Promise(r => setTimeout(r, 25));
    }
    isTyping.value = false;
  }

  const closeActiveStream = () => {
    if (streamController) {
      streamController.abort()
      streamController = null
      streamConversationId = ''
    }
  }

  const ensureAssistantMessage = (provider?: string, model?: string, agent?: string) => {
    const last = messages.value[messages.value.length - 1]
    if (last?.role === 'assistant' && !last.done) {
      last.content = ''
      last.thinking = ''
      last.provider = provider || last.provider
      last.model = model || last.model
      last.agent = agent || last.agent
      return messages.value.length - 1
    }

    messages.value.push({
      role: 'assistant',
      content: '',
      thinking: '',
      provider,
      model,
      agent,
      done: false
    })
    return messages.value.length - 1
  }

  const consumeAssistantStream = async (
    conversationId: string,
    assistantMsgIdx: number,
    scrollToBottom: () => void
  ) => {
    closeActiveStream()
    streamController = new AbortController()
    streamConversationId = conversationId
    runningByConversation.value[conversationId] = true

    const response = await openAssistantStream(conversationId, streamController.signal)
    const reader = response.body?.getReader()
    if (!reader) return

    const decoder = new TextDecoder()
    let buffer = ''
    let fullContent = ''
    let aborted = false

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const events = buffer.split('\n\n')
        buffer = events.pop() || ''

        for (const event of events) {
          const dataLines = event
            .split('\n')
            .map(line => line.trim())
            .filter(line => line.startsWith('data:'))
            .map(line => line.slice(5).trim())

          for (const dataStr of dataLines) {
            if (!dataStr || dataStr === '[DONE]') {
              runningByConversation.value[conversationId] = false
              continue
            }

            try {
              const parsed = JSON.parse(dataStr)
              const delta = parsed.message?.content || ''
              const thinkingDelta = parsed.message?.thinking || ''
              const state = parsed.message?.state || ''
              const msg = messages.value[assistantMsgIdx]
              if (!msg) continue

              msg.state = state || undefined

              if (thinkingDelta) {
                msg.thinking = (msg.thinking || '') + thinkingDelta
                scrollToBottom()
              }

              if (delta) {
                fullContent += delta
                msg.content = fullContent
                scrollToBottom()
              }

              if (parsed.done) {
                runningByConversation.value[conversationId] = false
              }
            } catch (e) {
              console.error('Failed to parse assistant stream event:', e)
            }
          }
        }
      }
    } catch (e: any) {
      if (e?.name === 'AbortError') {
        aborted = true
      } else {
        throw e
      }
    } finally {
      if (streamConversationId === conversationId) {
        streamController = null
        streamConversationId = ''
      }
      if (!aborted) {
        runningByConversation.value[conversationId] = false
        if (activeConversationId.value === conversationId) {
          const msg = messages.value[assistantMsgIdx]
          if (msg) {
            msg.done = true
            msg.state = undefined
          }
        }
        fetchAllTitles()
        syncRunningJobs()
      }
    }
  }

  const loadConversation = async (id: string, onNotFound: () => void, scrollToBottom: () => void) => {
    activeConversationId.value = id || ''
    if (id && streamConversationId === id) {
      initialized.value = true
      scrollToBottom()
      return
    }
    closeActiveStream()
    messages.value = []
    enableTransition.value = false

    if (!id) {
      initialized.value = true
      startTyping()
      return
    }

    try {
      const detail = await fetchConversationById(id)
      messages.value = detail.messages.map(m => ({
        role: m.role,
        content: m.content,
        thinking: m.thinking,
        provider: m.provider,
        model: m.model,
        agent: m.agent,
        done: true
      }))

      if (messages.value.length > 0) {
        // Restore last used provider/model
        const lastMsg = messages.value[messages.value.length - 1]
        if (lastMsg.provider && lastMsg.model) {
          activeProvider.value = lastMsg.provider
          activeModel.value = lastMsg.model
        }

        // Restore last used agent
        for (let i = messages.value.length - 1; i >= 0; i--) {
          const msg = messages.value[i]
          if (msg.agent) {
            selectedAgent.value = msg.agent
            break
          }
        }
      }

      const job = await fetchAssistantJob(id)
      runningByConversation.value[id] = job.running
      if (job.running) {
        const assistantMsgIdx = ensureAssistantMessage(activeProvider.value, activeModel.value, selectedAgent.value)
        consumeAssistantStream(id, assistantMsgIdx, scrollToBottom).catch(e => {
          if (e?.name !== 'AbortError') console.error('Assistant stream failed:', e)
        })
      }
    } catch (e: any) {
      if (e.message?.toLowerCase().includes('not found')) {
        onNotFound()
      }
    } finally {
      initialized.value = true
      scrollToBottom()
    }
  }

  const setSidebarCollapsed = (value: boolean) => {
    sidebarCollapsed.value = value
    Storage.set(StorageKeys.ASSISTANT_SIDEBAR_COLLAPSED, value)
  }

  const sendMessage = async (
    payload: { content: string; provider: string; model: string; agent?: string },
    chatId: string,
    onRedirect: (newId: string) => void,
    scrollToBottom: () => void
  ) => {
    if (!payload.content.trim()) return
    if (chatId && runningByConversation.value[chatId]) return
    if (!chatId && pendingNewChatSending.value) return

    if (isNewChat(chatId)) enableTransition.value = true
    activeConversationId.value = chatId || ''
    pendingNewChatSending.value = !chatId

    let currentChatId = chatId

    try {
      messages.value.push({
        role: 'user',
        content: payload.content,
        provider: payload.provider,
        model: payload.model,
        done: true
      })
      scrollToBottom()

      const job = await chatWithProvider(
        payload.provider,
        payload.model,
        [{ role: 'user', content: payload.content }],
        currentChatId,
        payload.agent,
        // selectedThinking.value
      )

      currentChatId = job.conversation_id
      activeConversationId.value = currentChatId
      pendingNewChatSending.value = false
      runningByConversation.value[currentChatId] = job.running
      syncRunningJobs()

      const assistantMsgIdx = ensureAssistantMessage(payload.provider, payload.model, payload.agent)
      const streamPromise = consumeAssistantStream(currentChatId, assistantMsgIdx, scrollToBottom)

      if (!chatId) {
        onRedirect(currentChatId)
        fetchAllTitles()
      }

      await streamPromise
    } catch (e: any) {
      pendingNewChatSending.value = false
      if (currentChatId) runningByConversation.value[currentChatId] = false
      throw e
    }
  }

  const stopAssistantChat = async (chatId: string) => {
    if (!chatId) return
    try {
      await stopChat(chatId)
      runningByConversation.value[chatId] = false
      syncRunningJobs()
    } catch (e) {
      console.error('Failed to stop chat:', e)
    }
  }

  return {
    messages,
    conversationList,
    providerList,
    runningByConversation,
    isSending,
    initialized,
    activeProvider,
    activeModel,
    displayedPrefix,
    displayedMain,
    enableTransition,
    isNewChat,
    getChatTitle,
    getProviderLogo,
    fetchAllTitles,
    syncRunningJobs,
    fetchProviderList,
    fetchAgents,
    loadConversation,
    sendMessage,
    startTyping,
    setSidebarCollapsed,
    sidebarCollapsed,
    isTyping,
    agents,
    selectedAgent,
    selectedThinking,
    stopAssistantChat
  }
})
