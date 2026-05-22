<template>
  <AppPage :scrollable="false" :no-padding="true">
    <el-container style="height: 100%; border-radius: 0; overflow: hidden;">
      <AssistantSidebar :active-chat-id="activeChatId" />
      <el-container direction="vertical"
        style="background: var(--el-bg-color); min-width: 0; flex: 1; position: relative;">
        <el-header
          style="border-bottom: 1px solid var(--el-border-color-lighter); height: 60px; width: 100%; box-sizing: border-box;">
          <el-row align="middle" justify="space-between" style="height: 100%; width: 100%;">
            <el-button :icon="store.sidebarCollapsed ? ArrowRight : ArrowLeft"
              @click="store.setSidebarCollapsed(!store.sidebarCollapsed)" />
            <el-text tag="b" size="large">{{ activeChatTitle }}</el-text>
            <el-button text :icon="Operation" @click="showSettings = true">Providers</el-button>
          </el-row>
        </el-header>
        <el-main style="padding: 0; position: relative; overflow: hidden; flex: 1; min-width: 0;">
          <el-scrollbar ref="scrollBox" style="height: 100%; width: 100%;" @scroll="handleScroll">
            <el-row justify="center" style="padding: 40px 0; width: 100%; margin: 0; min-width: 0;">
              <el-col :xs="24" :sm="22" :md="20" :lg="18" :xl="16"
                style="max-width: 800px; padding: 0 20px; width: 100%; min-width: 0; overflow: hidden;">
                <div v-if="store.initialized">

                  <!-- Welcome Hero -->
                  <div v-if="isNewChat" style="margin-top: 15vh; text-align: left; min-height: 100px;">
                    <el-space direction="vertical" :size="16" alignment="flex-start">
                      <el-text tag="b" style="line-height: 1.5;">
                        {{ store.displayedPrefix }}
                      </el-text>
                      <el-text tag="b" size="large" style="line-height: 1.5;">
                        {{ store.displayedMain }}
                      </el-text>
                    </el-space>
                  </div>

                  <!-- Message List -->
                  <div v-else style="display: flex; flex-direction: column; gap: 24px;">
                    <ChatCard v-for="(msg, index) in store.messages" :key="index" :role="msg.role"
                      :content="msg.content" :thinking="msg.thinking" :provider="msg.provider" :model="msg.model"
                      :provider-logo="store.getProviderLogo(msg.provider)" :done="msg.done" :state="msg.state" />
                  </div>
                </div>
              </el-col>
            </el-row>
          </el-scrollbar>
        </el-main>
        <!-- Input Layer -->
        <div class="input-layer" :class="{ 'is-centered': isNewChat }">
          <el-row justify="center" style="width: 100%; margin: 0;">
            <el-col :xs="24" :sm="22" :md="20" :lg="18" :xl="16"
              style="max-width: 800px; padding: 0 20px; width: 100%;">
              <ChatInput ref="chatInput" v-model:provider="store.activeProvider" v-model:model="store.activeModel"
                @send="handleSendMessage" @stop="store.stopAssistantChat(activeChatId)" />
            </el-col>
          </el-row>
        </div>
      </el-container>
    </el-container>
    <AssistantProviders v-model="showSettings" @saved="handleSettingsSaved" />
  </AppPage>
</template>
<script setup lang="ts">
import { ref, watch, onMounted, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ArrowRight, Operation } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ChatCard from '../components/ChatCard.vue'
import ChatInput from '../components/ChatInput.vue'
import AssistantSidebar from '../components/AssistantSidebar.vue'
import AssistantProviders from '../components/AssistantProviders.vue'
import { useAssistantStore } from '../store/useAssistantStore'
import AppPage from '@shared/components/AppPage.vue'


const route = useRoute()
const router = useRouter()
const store = useAssistantStore()

const scrollBox = ref<any>(null)
const chatInput = ref<any>(null)
const showSettings = ref(false)
const isAutoScrollEnabled = ref(true)

const activeChatId = computed(() => route.params.id as string || '')
const isNewChat = computed(() => store.isNewChat(activeChatId.value))
const activeChatTitle = computed(() => store.getChatTitle(activeChatId.value))

const handleSettingsSaved = () => {
  if (chatInput.value) chatInput.value.refresh()
}

const scrollToBottom = async (force = false) => {
  if (!force && !isAutoScrollEnabled.value) return

  await nextTick()
  if (scrollBox.value) {
    const scrollEl = scrollBox.value.$el.querySelector('.el-scrollbar__wrap')
    if (scrollEl) {
      scrollEl.scrollTop = scrollEl.scrollHeight
    }
  }
}

const handleSendMessage = async (payload: { content: string; provider: string; model: string; agent?: string }) => {
  try {
    isAutoScrollEnabled.value = true
    await store.sendMessage(
      payload,
      activeChatId.value,
      (newId) => router.replace(`/assistant/${newId}`),
      () => scrollToBottom(false)
    )
  } catch (e) {
    ElMessage.error('Failed to send message')
  }
}

onMounted(() => {
  store.loadConversation(activeChatId.value, () => router.push('/assistant'), () => scrollToBottom(true))
  store.fetchProviderList()
  store.fetchAllTitles()
  store.syncRunningJobs()
  store.fetchAgents()
})

watch(() => route.params.id, (newId) => {
  isAutoScrollEnabled.value = true
  store.loadConversation(newId as string, () => router.push('/assistant'), () => scrollToBottom(true))
})

const handleScroll = ({ scrollTop }: { scrollTop: number }) => {
  if (!scrollBox.value) return
  const scrollEl = scrollBox.value.$el.querySelector('.el-scrollbar__wrap')
  if (!scrollEl) return

  const { scrollHeight, clientHeight } = scrollEl
  const distanceFromBottom = scrollHeight - clientHeight - scrollTop

  // If user scrolls up more than 50px from bottom, disable auto-scroll
  if (distanceFromBottom > 50) {
    isAutoScrollEnabled.value = false
  } else {
    isAutoScrollEnabled.value = true
  }
}
</script>
<style>
@import '../styles/markdown.css';
</style>
<style scoped>
.input-layer {
  width: 100%;
  padding: 0 0 32px 0;
  box-sizing: border-box;
  z-index: 10;
  background: linear-gradient(to top, var(--el-bg-color) 80%, transparent);
}

.input-layer.is-animating {
  transition: all 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

.input-layer.is-centered {
  position: absolute;
  top: 55%;
  left: 0;
  right: 0;
  transform: translateY(-50%);
  padding-bottom: 0;
  background: transparent;
}
</style>
