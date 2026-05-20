import { type Component } from 'vue'
import AutoNodeForm from '../components/EkAutoNodeForm.vue'

type NodeResolver = (type: string) => Component | undefined

const resolvers: NodeResolver[] = []

export function registerNodeResolver(resolver: NodeResolver) {
  resolvers.push(resolver)
}

const registry: Record<string, Component> = {
  // Register nodes here that require a custom form component instead of the auto-generated one.
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
