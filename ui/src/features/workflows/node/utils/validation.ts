/**
 * Runtime validation utilities for node configuration.
 * Validates that node configs match their schema definitions.
 */

import type { NodeDefinition, NodeField, NodeFieldType } from '@workflows/node/types/node'

/**
 * Result of a validation operation
 */
export interface ValidationResult {
  /** Whether the validation passed */
  valid: boolean
  
  /** List of validation error messages */
  errors: string[]
}

/**
 * Validate a node configuration against its schema definition.
 * Checks required fields, types, and option constraints.
 * 
 * @param config - The configuration object to validate
 * @param schema - The node definition schema
 * @returns Validation result with any errors found
 * 
 * @example
 * ```typescript
 * const result = validateNodeConfig(
 *   { action: 'sleep', duration_ms: 1000 },
 *   nodeDefinition
 * )
 * if (!result.valid) {
 *   console.error(result.errors)
 * }
 * ```
 */
export function validateNodeConfig(
  config: Record<string, any>,
  schema: NodeDefinition
): ValidationResult {
  const errors: string[] = []

  // Determine which fields to validate
  const fieldsToValidate: NodeField[] = []

  // Validate global_fields
  if (schema.global_fields) {
    fieldsToValidate.push(...schema.global_fields)
  }

  // Validate action-specific fields
  const action = config.action || schema.default_action
  if (action && schema.actions) {
    const actionDef = schema.actions.find(a => a.key === action)
    if (actionDef) {
      fieldsToValidate.push(...actionDef.fields)
    }
  }

  // Validate each field
  for (const field of fieldsToValidate) {
    const fieldErrors = validateField(config, field)
    errors.push(...fieldErrors)
  }

  return {
    valid: errors.length === 0,
    errors
  }
}

/**
 * Validate a single field against its schema definition.
 * 
 * @param config - The full configuration object
 * @param field - The field schema to validate against
 * @returns Array of error messages (empty if valid)
 */
function validateField(
  config: Record<string, any>,
  field: NodeField
): string[] {
  const errors: string[] = []
  const value = config[field.key]
  const hasValue = field.key in config
  
  // Check required fields
  if (field.required && !hasValue) {
    // If field has a default value, it's OK - the default will be used
    if (field.default !== undefined) {
      return errors
    }
    errors.push(`Missing required field: ${field.key}`)
    return errors
  }
  
  // Skip validation if field is not present and not required
  if (!hasValue) {
    return errors
  }
  
  // Validate type
  if (!validateFieldType(value, field.type)) {
    errors.push(
      `Invalid type for "${field.key}": expected ${field.type}, got ${getValueType(value)}`
    )
  }
  
  // Validate options constraint
  if (field.options && Array.isArray(field.options) && field.options.length > 0) {
    const validValues = field.options.map((opt: any) => 
      (typeof opt === 'object' && opt !== null && 'value' in opt) ? opt.value : opt
    )
    
    if (!validValues.includes(value)) {
      errors.push(
        `Invalid value for "${field.key}": must be one of [${validValues.join(', ')}], got "${value}"`
      )
    }
  }
  
  return errors
}

/**
 * Check if a value matches the expected field type.
 * 
 * @param value - The value to check
 * @param type - The expected field type
 * @returns True if the value matches the type
 */
function validateFieldType(value: any, type: NodeFieldType): boolean {
  switch (type) {
    case 'string':
      return typeof value === 'string'
    
    case 'number':
      return typeof value === 'number' && !isNaN(value)
    
    case 'boolean':
      return typeof value === 'boolean'
    
    case 'json':
      return typeof value === 'object' && value !== null && !Array.isArray(value)
    
    case 'array':
      return Array.isArray(value)
    
    default:
      // Unknown type, allow it
      return true
  }
}

/**
 * Get a human-readable type name for a value.
 * 
 * @param value - The value to get the type of
 * @returns Type name string
 */
function getValueType(value: any): string {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

/**
 * Validate that a handle key exists in the node's output definitions.
 * 
 * @param handle - The handle key to validate
 * @param schema - The node definition schema
 * @returns True if the handle is valid
 */
export function validateOutputHandle(
  handle: string,
  schema: NodeDefinition
): boolean {
  return schema.outputs.some(output => output.key === handle)
}

/**
 * Get all valid output handle keys for a node.
 * 
 * @param schema - The node definition schema
 * @returns Array of valid handle keys
 */
export function getValidOutputHandles(schema: NodeDefinition): string[] {
  return schema.outputs.map(output => output.key)
}
