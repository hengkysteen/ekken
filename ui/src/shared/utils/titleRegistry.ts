import { reactive } from 'vue'

interface Registry {
  [id: string]: string
}

export const titleRegistry = reactive<Registry>({})

export function setTitle(id: string, title: string) {
  if (!id || !title) return
  titleRegistry[id] = title
}

export function getTitle(id: string): string | null {
  return titleRegistry[id] || null
}

export function bulkSetTitles(items: { id: string, title: string }[]) {
  items.forEach(item => {
    titleRegistry[item.id] = item.title
  })
}
