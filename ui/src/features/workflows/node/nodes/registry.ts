import { type Component } from 'vue'
import AutoNodeForm from '@shared/components/ek/EkAutoNodeForm.vue'
import HttpNodeForm from './http/HttpNodeForm.vue'

type NodeResolver = (type: string) => Component | undefined

const resolvers: NodeResolver[] = []

export function registerNodeResolver(resolver: NodeResolver) {
  resolvers.push(resolver)
}

const registry: Record<string, Component> = {
  'http': HttpNodeForm,
}
export function getNodeFormComponent(type: string): Component {
  if (registry[type]) return registry[type]
  for (const resolver of resolvers) {
    const comp = resolver(type)
    if (comp) return comp
  }
  return AutoNodeForm
}

export default registry
