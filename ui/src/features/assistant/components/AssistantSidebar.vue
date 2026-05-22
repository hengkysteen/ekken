<template>
  <el-aside :width="store.sidebarCollapsed ? '0' : '260px'" class="assistant-sidebar"
    :class="{ 'is-collapsed': store.sidebarCollapsed }">
    <el-container class="full-height">
      <!-- Header Area -->
      <el-header height="auto" class="sidebar-header">
        <el-row align="middle" justify="space-between" style="width: 100%">
          <el-text size="large" tag="b">Assistant</el-text>
          
        </el-row>
        <el-button dashed type="primary" class="new-chat-btn" @click="handleNewChat">
          New Chat
        </el-button>

      </el-header>

      <!-- Scrollable List Area -->
      <el-main class="sidebar-main">
        <el-scrollbar>
          <div class="list-container">
            <template v-if="store.conversationList.length > 0">
              <ListTile v-for="chat in store.conversationList" :key="chat.id"
                :selected="String(activeChatId) === String(chat.id)" class="chat-item"
                @click="router.push(`/assistant/${chat.id}`)" @contextmenu.prevent="onContextMenu($event, chat)">
                <template #leading>
                  <el-icon v-if="store.runningByConversation[String(chat.id)]" class="running-icon">
                    <Loading />
                  </el-icon>
                  <el-icon v-else>
                    <ChatSquare />
                  </el-icon>
                </template>
                <template #title>
                  <el-text truncated>{{ chat.title }}</el-text>
                </template>
              </ListTile>
            </template>
            <el-empty v-else description="No chats yet" :image-size="40" />
          </div>
        </el-scrollbar>
      </el-main>
    </el-container>

    <!-- Context Menu for Rename/Delete -->
    <el-dropdown ref="dropdownRef" trigger="contextmenu" virtual-triggering :virtual-ref="triggerRef"
      @command="handleCommand">
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item :command="{ type: 'rename', chat: selectedChat }">Rename</el-dropdown-item>
          <el-dropdown-item divided type="danger"
            :command="{ type: 'delete', chat: selectedChat }">Delete</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </el-aside>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ChatSquare, Loading } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { renameConversation, deleteConversation, type Conversation } from '../api'
import { setTitle } from '@shared/utils/titleRegistry'
import ListTile from '@shared/components/ListTile.vue'
import { useAssistantStore } from '../store/useAssistantStore'

const props = defineProps<{
  activeChatId: string
}>()

const store = useAssistantStore()
const router = useRouter()

// Auto-register titles to global registry
watch(() => store.conversationList, (list) => {
  list.forEach(c => setTitle(String(c.id), c.title))
}, { immediate: true })

const dropdownRef = ref()
const selectedChat = ref<Conversation | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)
const triggerRef = ref({
  getBoundingClientRect() {
    return DOMRect.fromRect({ width: 0, height: 0, x: mouseX.value, y: mouseY.value })
  }
})

const onContextMenu = (e: MouseEvent, chat: Conversation) => {
  selectedChat.value = chat; mouseX.value = e.clientX; mouseY.value = e.clientY;
  dropdownRef.value?.handleOpen?.()
}

const handleCommand = (cmd: any) => {
  if (cmd.type === 'rename') handleRename(cmd.chat)
  else if (cmd.type === 'delete') handleDelete(cmd.chat)
}

const handleRename = (chat: Conversation) => {
  ElMessageBox.prompt('New name', 'Rename', { inputValue: chat.title }).then(async ({ value }) => {
    if (value && value !== chat.title) {
      await renameConversation(chat.id, value)
      store.fetchAllTitles()
      ElMessage.success('Renamed')
    }
  })
}

const handleDelete = (chat: Conversation) => {
  ElMessageBox.confirm(`Delete "${chat.title}"?`, 'Warning', { type: 'warning' }).then(async () => {
    await deleteConversation(chat.id)
    store.fetchAllTitles()
    if (props.activeChatId === chat.id) router.push('/assistant')
  })
}

 
const handleNewChat = () => {
  router.push('/assistant')
}
</script>

<style scoped>
.assistant-sidebar {
  overflow: hidden;
  border-right: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.assistant-sidebar.is-collapsed {
  border-right-width: 0;
}

.sidebar-header {
  padding: 20px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.new-chat-btn {
  width: 100%;
  margin-top: 16px;
  border-radius: 12px;
}

.sidebar-main {
  padding: 0;
}

.list-container {
  padding: 12px 8px;
}

.chat-item {
  margin-bottom: 4px;
}

.running-icon {
  animation: spin 1s linear infinite;
  color: var(--el-color-primary);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.full-height {
  height: 100%;
}
</style>
