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
 * 
 * @export
 * @interface UpdateTriggerRequest
 */
export interface UpdateTriggerRequest {
    /**
     * 
     * @type {string}
     * @memberof UpdateTriggerRequest
     */
    'name'?: string;
    /**
     * Trigger configuration containing trigger-specific parameters and settings
     * @type {{ [key: string]: any; }}
     * @memberof UpdateTriggerRequest
     */
    'config'?: { [key: string]: any; };
    /**
     * 
     * @type {boolean}
     * @memberof UpdateTriggerRequest
     */
    'enabled'?: boolean;
}

