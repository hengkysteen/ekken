import type { Workflow } from '@workflows/workflow/api'

export function useWorkflowFile() {
  function exportWorkflowToFile(workflow: Workflow, filename: string) {
    const blob = new Blob([JSON.stringify(workflow, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${filename}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  function importWorkflowFromJson(jsonString: string): Workflow {
    const imported = JSON.parse(jsonString) as Workflow
    if (!imported.nodes) {
      throw new Error('Invalid workflow format: missing nodes')
    }
    return imported
  }

  return {
    exportWorkflowToFile,
    importWorkflowFromJson
  }
}
