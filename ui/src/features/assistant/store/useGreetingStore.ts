import { defineStore } from 'pinia'
import { ref } from 'vue'
import { Storage, StorageKeys } from '@shared/utils/storage'
import { useProfileStore } from '@profile/stores/profile'

export const useGreetingStore = defineStore('assistantGreeting', () => {
  const displayedPrefix = ref('')
  const displayedMain = ref('')
  const isTyping = ref(false)
  const nextGreetingIdx = ref(Storage.get<number>(StorageKeys.ASSISTANT_NEXT_GREETING) ?? 0)

  const greetings = [
    { prefix: "Ahoy, Captain{{name}}! 🏴‍☠️", main: "Give the word and I'll burn the seas for ya." },
    { prefix: "Welcome back, {{name}}!", main: "Ready to execute your commands." },
    { prefix: "{{timeOfDay}}, My Lord", main: "Shall we automate something and let the machines work while you sleep?" },
    { prefix: "Shiver me timbers, {{name}}! 🏴‍☠️", main: "Not all treasure is silver and gold, mate. So what are we plunderin' today?" },
  ]

  const startTyping = async () => {
    if (isTyping.value) return
    isTyping.value = true
    const g = greetings[nextGreetingIdx.value]

    // Increment for next time (Round Robin)
    nextGreetingIdx.value = (nextGreetingIdx.value + 1) % greetings.length
    Storage.set(StorageKeys.ASSISTANT_NEXT_GREETING, nextGreetingIdx.value)

    displayedPrefix.value = ''
    displayedMain.value = ''

    const profileStore = useProfileStore()
    const name = profileStore.profile?.name?.trim()

    const hour = new Date().getHours()
    let timeOfDay = 'Evening'
    let dayOrNight = 'night'

    if (hour >= 0 && hour < 5) {
      timeOfDay = 'Late Night'
      dayOrNight = 'night'
    } else if (hour >= 5 && hour < 12) {
      timeOfDay = 'Morning'
      dayOrNight = 'day'
    } else if (hour >= 12 && hour < 18) {
      timeOfDay = 'Afternoon'
      dayOrNight = 'day'
    } else {
      // jam 18–23
      timeOfDay = 'Evening'
      dayOrNight = 'night'
    }

    const prefix = g.prefix
      .replace('{{name}}', name ? ` ${name}` : '')
      .replace('{{timeOfDay}}', timeOfDay)

    const main = g.main.replace('{{dayOrNight}}', dayOrNight)

    for (const char of prefix) {
      displayedPrefix.value += char
      await new Promise(r => setTimeout(r, 45))
    }

    await new Promise(r => setTimeout(r, 150))

    const segmenter = new Intl.Segmenter(undefined, { granularity: 'grapheme' })
    for (const { segment } of segmenter.segment(main)) {
      displayedMain.value += segment
      await new Promise(r => setTimeout(r, 25))
    }
    isTyping.value = false
  }

  return {
    displayedPrefix,
    displayedMain,
    isTyping,
    startTyping
  }
})
