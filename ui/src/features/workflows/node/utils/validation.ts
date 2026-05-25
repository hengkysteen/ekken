/**
 * Runtime validation utilities for node configuration.
 * Validates that node configs match their schema definitions.
 */

import type { NodeDefinition, NodeField, NodeFieldType } from '@workflows/node/types/node'
import { getActionBlueprint, getActionType, getActionValue } from './node'

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
/**
 * Validate a node configuration (NodeAction) against its schema definition.
 * 
 * @param action - The NodeAction object containing fields and their values
 * @param schema - The node definition schema
 * @returns Validation result with any errors found
 */
export function validateNodeConfig(
  action: any,
  schema: NodeDefinition
): ValidationResult {
  const errors: string[] = []
  const actionType = getActionType(action)
  const actionBlueprint = getActionBlueprint(schema, actionType)

  if (!actionBlueprint) {
    errors.push(`Invalid action: ${actionType || '(empty)'}`)
    return { valid: false, errors }
  }

  const schemaFields = [...(schema.global_fields || []), ...(actionBlueprint.fields || [])]
  const schemaKeys = new Set(schemaFields.map(field => field.key))

  for (const field of action?.fields || []) {
    if (field?.key && !schemaKeys.has(field.key)) {
      errors.push(`Unknown field: ${field.key}`)
    }
  }

  for (const field of schemaFields) {
    const fieldErrors = validateField(field, getActionValue(action, field.key, field.default))
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
 * @param field - The field object (NodeField) containing its value and constraints
 * @returns Array of error messages (empty if valid)
 */
function validateField(
  field: NodeField,
  value: any
): string[] {
  const errors: string[] = []
  const hasValue = value !== undefined && value !== null && (typeof value === 'string' ? value.trim() !== '' : true)
  
  // Check required fields
  if (field.required && !hasValue) {
    // If field has a default value, it's OK - the default will be used
    if (field.default !== undefined) {
      return errors
    }
    errors.push(`Missing required field: ${field.label || field.key}`)
    return errors
  }
  
  // Skip validation if field is not present and not required
  if (!hasValue) {
    return errors
  }
  
  // Validate type
  if (!validateFieldType(value, field.type)) {
    errors.push(
      `Invalid type for "${field.label || field.key}": expected ${field.type}, got ${getValueType(value)}`
    )
  }
  
  // Validate options constraint
  if (field.options && Array.isArray(field.options) && field.options.length > 0) {
    const validValues = field.options.map((opt: any) => 
      (typeof opt === 'object' && opt !== null && 'value' in opt) ? opt.value : opt
    )
    
    if (!validValues.includes(value)) {
      errors.push(
        `Invalid value for "${field.label || field.key}": must be one of [${validValues.join(', ')}], got "${value}"`
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
  return (schema.output_handles || []).includes(handle)
}

/**
 * Get all valid output handle keys for a node.
 * 
 * @param schema - The node definition schema
 * @returns Array of valid handle keys
 */
export function getValidOutputHandles(schema: NodeDefinition): string[] {
  return schema.output_handles || []
}
