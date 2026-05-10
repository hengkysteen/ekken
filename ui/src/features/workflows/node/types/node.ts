/**
 * Node type definitions for Ekken workflow system.
 * These types define the contract between frontend and backend for node metadata and configuration.
 */

/**
 * Supported field types for node configuration
 */
export type NodeFieldType = string

/**
 * Visual tone for node output handles
 */
export type NodeOutputTone = 'success' | 'error' | 'warning' | 'info' | 'neutral'

/**
 * Configuration field definition for a node.
 * Defines the schema for a single configuration parameter.
 */
export interface NodeField {
  /** Unique identifier for this field */
  key: string

  /** Data type of this field */
  type: NodeFieldType

  /** Whether this field is required */
  required?: boolean

  /** Human-readable label for this field */
  label: string

  /** Valid options for this field (for dropdown/select inputs) */
  options?: any

  /** Default value for this field */
  default?: any
}

/**
 * Layout item for dynamic form rendering.
 */
export interface Form {
  /** Reference to field key or unique UI element key */
  key: string

  /** Flex span for grid layout (1-24) */
  flex: number

  /** UI component to use for rendering */
  component?: string

  /** Specific options for the UI component */
  form_options?: any
}

/**
 * Action variant with its own set of fields.
 * Used for nodes that have different configuration schemas based on the action.
 */
export interface NodeAction {
  /** Unique action key (e.g., "login", "send") */
  key: string

  /** Human-readable label for this action */
  label: string

  /** Description of what this action does */
  description: string

  /** Configuration fields specific to this action */
  fields: NodeField[]

  /** Layout definition for dynamic form rendering */
  form?: Form[][]

  /** Whether this action produces a response that can be used in the workflow */
  has_response: boolean

  /** Compatibility properties for older components */
  auto_layout?: any
  has_output?: boolean
}

/**
 * Output handle definition for a node.
 * Defines a connection point that can be used to route workflow execution.
 */
export interface NodeOutput {
  /** Unique identifier for this output handle */
  key: string

  /** Human-readable label for this output */
  label: string

  /** Visual style/tone for this output */
  tone: NodeOutputTone
}

/**
 * Complete node definition from the catalog.
 * This is the metadata that describes a node type and how to configure it.
 */
/**
 * Common metadata shared between node specifications and node instances.
 */
export interface NodeMetadata {
  /** Unique type identifier for this node */
  type: string

  /** Specification version for this node */
  version?: string

  /** Categorization tags for organizing nodes in the UI */
  tags?: string[]

  /** Human-readable label for this node */
  label: string

  /** Icon URL or identifier for visual representation */
  icon?: string
}

/**
 * Complete node definition from the catalog.
 * This is the metadata that describes a node type and how to configure it.
 */
export interface NodeDefinition extends NodeMetadata {
  /** Parent node type (for nested nodes) */
  parent?: string

  /** Detailed description of what this node does */
  description: string

  /** Configuration fields for the parent node (when this is a nested node) */
  parent_config?: NodeField[]

  /** Whether this node can contain nested child nodes */
  supports_nested_nodes: boolean

  /** Per-action field groups */
  actions: NodeAction[]

  /** Shared fields across all actions */
  global_fields?: NodeField[]

  /** Layout definition for global fields */
  global_form?: Form[][]

  /** Default action when using Actions pattern */
  default_action?: string

  /** Output handles for routing workflow execution */
  outputs: NodeOutput[]
}

/**
 * Runtime node instance in a workflow
 */
export interface WorkflowNode extends NodeMetadata {
  /** Unique instance ID */
  id: string

  /** Configuration values for this node instance */
  config: Record<string, any>

  /** Variable name to save this node's output to */
  response_var?: string

  /** Source type of the node (catalog or mynodes) */
  sourceType?: 'catalog' | 'mynodes'

  /** Custom name (if this is a mynode) */
  name?: string

  /** Position in the visual editor */
  position?: { x: number; y: number }

  /** Child nodes (for nested workflows) */
  nodes?: WorkflowNode[]

  /** Edges between child nodes (for nested workflows) */
  edges?: Array<{ source: string; sourceHandle: string; target: string }>

  /** Positions of child nodes (for nested workflows) */
  positions?: Record<string, { x: number; y: number }>
}



/**
 * Props that every node form component receives.
 * All node form components (internal and plugin) MUST accept these props.
 */
export interface NodeFormProps {
  /** The node instance being configured */
  node: {
    /** Unique instance ID */
    id: string

    /** Node type identifier */
    type: string

    /** Node data including config and metadata */
    data: NodeMetadata & {
      /** Node type (may differ from type for legacy reasons) */
      nodeType: string

      /** Current configuration values */
      config: Record<string, any>

      /** Variable name to save output to */
      response_var?: string
    }
  }
}

/**
 * Interface that node form components MUST implement.
 * Components must expose a getData() method that returns the current config.
 */
export interface NodeFormComponent {
  /** Props received by the component */
  $props: NodeFormProps

  /**
   * Returns the current configuration state.
   * This method MUST be exposed via expose() in the component setup.
   * The returned object should match the schema defined in the node's fields.
   *
   * @returns Configuration object with values for each field
   */
  getData(): Record<string, any>
}
