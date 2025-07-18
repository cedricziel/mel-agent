/* tslint:disable */
/* eslint-disable */
/**
 * MEL Agent API
 * AI Agents SaaS platform API with visual workflow builder
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */



/**
 * Status of a workflow step
 * @export
 * @enum {string}
 */

export const WorkflowStepStatus = {
    Pending: 'pending',
    Running: 'running',
    Completed: 'completed',
    Failed: 'failed',
    Skipped: 'skipped'
} as const;

export type WorkflowStepStatus = typeof WorkflowStepStatus[keyof typeof WorkflowStepStatus];



