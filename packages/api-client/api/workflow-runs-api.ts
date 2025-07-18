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


import type { Configuration } from '../configuration';
import type { AxiosPromise, AxiosInstance, RawAxiosRequestConfig } from 'axios';
import globalAxios from 'axios';
// Some imports not used depending on template conditions
// @ts-ignore
import { DUMMY_BASE_URL, assertParamExists, setApiKeyToObject, setBasicAuthToObject, setBearerAuthToObject, setOAuthToObject, setSearchParams, serializeDataIfNeeded, toPathString, createRequestFunction } from '../common';
// @ts-ignore
import { BASE_PATH, COLLECTION_FORMATS, type RequestArgs, BaseAPI, RequiredError, operationServerMap } from '../base';
// @ts-ignore
import type { WorkflowRun } from '../models';
// @ts-ignore
import type { WorkflowRunList } from '../models';
// @ts-ignore
import type { WorkflowRunStatus } from '../models';
// @ts-ignore
import type { WorkflowStep } from '../models';
/**
 * WorkflowRunsApi - axios parameter creator
 * @export
 */
export const WorkflowRunsApiAxiosParamCreator = function (configuration?: Configuration) {
    return {
        /**
         * 
         * @summary Get workflow run details
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        getWorkflowRun: async (id: string, options: RawAxiosRequestConfig = {}): Promise<RequestArgs> => {
            // verify required parameter 'id' is not null or undefined
            assertParamExists('getWorkflowRun', 'id', id)
            const localVarPath = `/api/workflow-runs/{id}`
                .replace(`{${"id"}}`, encodeURIComponent(String(id)));
            // use dummy base URL string because the URL constructor only accepts absolute URLs.
            const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
            let baseOptions;
            if (configuration) {
                baseOptions = configuration.baseOptions;
            }

            const localVarRequestOptions = { method: 'GET', ...baseOptions, ...options};
            const localVarHeaderParameter = {} as any;
            const localVarQueryParameter = {} as any;

            // authentication ApiKeyAuth required
            await setApiKeyToObject(localVarHeaderParameter, "X-API-Key", configuration)

            // authentication BearerAuth required
            // http bearer authentication required
            await setBearerAuthToObject(localVarHeaderParameter, configuration)


    
            setSearchParams(localVarUrlObj, localVarQueryParameter);
            let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
            localVarRequestOptions.headers = {...localVarHeaderParameter, ...headersFromBaseOptions, ...options.headers};

            return {
                url: toPathString(localVarUrlObj),
                options: localVarRequestOptions,
            };
        },
        /**
         * 
         * @summary Get workflow run steps
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        getWorkflowRunSteps: async (id: string, options: RawAxiosRequestConfig = {}): Promise<RequestArgs> => {
            // verify required parameter 'id' is not null or undefined
            assertParamExists('getWorkflowRunSteps', 'id', id)
            const localVarPath = `/api/workflow-runs/{id}/steps`
                .replace(`{${"id"}}`, encodeURIComponent(String(id)));
            // use dummy base URL string because the URL constructor only accepts absolute URLs.
            const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
            let baseOptions;
            if (configuration) {
                baseOptions = configuration.baseOptions;
            }

            const localVarRequestOptions = { method: 'GET', ...baseOptions, ...options};
            const localVarHeaderParameter = {} as any;
            const localVarQueryParameter = {} as any;

            // authentication ApiKeyAuth required
            await setApiKeyToObject(localVarHeaderParameter, "X-API-Key", configuration)

            // authentication BearerAuth required
            // http bearer authentication required
            await setBearerAuthToObject(localVarHeaderParameter, configuration)


    
            setSearchParams(localVarUrlObj, localVarQueryParameter);
            let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
            localVarRequestOptions.headers = {...localVarHeaderParameter, ...headersFromBaseOptions, ...options.headers};

            return {
                url: toPathString(localVarUrlObj),
                options: localVarRequestOptions,
            };
        },
        /**
         * 
         * @summary List workflow runs
         * @param {string} [workflowId] 
         * @param {WorkflowRunStatus} [status] 
         * @param {number} [page] 
         * @param {number} [limit] 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        listWorkflowRuns: async (workflowId?: string, status?: WorkflowRunStatus, page?: number, limit?: number, options: RawAxiosRequestConfig = {}): Promise<RequestArgs> => {
            const localVarPath = `/api/workflow-runs`;
            // use dummy base URL string because the URL constructor only accepts absolute URLs.
            const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
            let baseOptions;
            if (configuration) {
                baseOptions = configuration.baseOptions;
            }

            const localVarRequestOptions = { method: 'GET', ...baseOptions, ...options};
            const localVarHeaderParameter = {} as any;
            const localVarQueryParameter = {} as any;

            // authentication ApiKeyAuth required
            await setApiKeyToObject(localVarHeaderParameter, "X-API-Key", configuration)

            // authentication BearerAuth required
            // http bearer authentication required
            await setBearerAuthToObject(localVarHeaderParameter, configuration)

            if (workflowId !== undefined) {
                localVarQueryParameter['workflow_id'] = workflowId;
            }

            if (status !== undefined) {
                localVarQueryParameter['status'] = status;
            }

            if (page !== undefined) {
                localVarQueryParameter['page'] = page;
            }

            if (limit !== undefined) {
                localVarQueryParameter['limit'] = limit;
            }


    
            setSearchParams(localVarUrlObj, localVarQueryParameter);
            let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
            localVarRequestOptions.headers = {...localVarHeaderParameter, ...headersFromBaseOptions, ...options.headers};

            return {
                url: toPathString(localVarUrlObj),
                options: localVarRequestOptions,
            };
        },
    }
};

/**
 * WorkflowRunsApi - functional programming interface
 * @export
 */
export const WorkflowRunsApiFp = function(configuration?: Configuration) {
    const localVarAxiosParamCreator = WorkflowRunsApiAxiosParamCreator(configuration)
    return {
        /**
         * 
         * @summary Get workflow run details
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        async getWorkflowRun(id: string, options?: RawAxiosRequestConfig): Promise<(axios?: AxiosInstance, basePath?: string) => AxiosPromise<WorkflowRun>> {
            const localVarAxiosArgs = await localVarAxiosParamCreator.getWorkflowRun(id, options);
            const localVarOperationServerIndex = configuration?.serverIndex ?? 0;
            const localVarOperationServerBasePath = operationServerMap['WorkflowRunsApi.getWorkflowRun']?.[localVarOperationServerIndex]?.url;
            return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
        },
        /**
         * 
         * @summary Get workflow run steps
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        async getWorkflowRunSteps(id: string, options?: RawAxiosRequestConfig): Promise<(axios?: AxiosInstance, basePath?: string) => AxiosPromise<Array<WorkflowStep>>> {
            const localVarAxiosArgs = await localVarAxiosParamCreator.getWorkflowRunSteps(id, options);
            const localVarOperationServerIndex = configuration?.serverIndex ?? 0;
            const localVarOperationServerBasePath = operationServerMap['WorkflowRunsApi.getWorkflowRunSteps']?.[localVarOperationServerIndex]?.url;
            return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
        },
        /**
         * 
         * @summary List workflow runs
         * @param {string} [workflowId] 
         * @param {WorkflowRunStatus} [status] 
         * @param {number} [page] 
         * @param {number} [limit] 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        async listWorkflowRuns(workflowId?: string, status?: WorkflowRunStatus, page?: number, limit?: number, options?: RawAxiosRequestConfig): Promise<(axios?: AxiosInstance, basePath?: string) => AxiosPromise<WorkflowRunList>> {
            const localVarAxiosArgs = await localVarAxiosParamCreator.listWorkflowRuns(workflowId, status, page, limit, options);
            const localVarOperationServerIndex = configuration?.serverIndex ?? 0;
            const localVarOperationServerBasePath = operationServerMap['WorkflowRunsApi.listWorkflowRuns']?.[localVarOperationServerIndex]?.url;
            return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
        },
    }
};

/**
 * WorkflowRunsApi - factory interface
 * @export
 */
export const WorkflowRunsApiFactory = function (configuration?: Configuration, basePath?: string, axios?: AxiosInstance) {
    const localVarFp = WorkflowRunsApiFp(configuration)
    return {
        /**
         * 
         * @summary Get workflow run details
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        getWorkflowRun(id: string, options?: RawAxiosRequestConfig): AxiosPromise<WorkflowRun> {
            return localVarFp.getWorkflowRun(id, options).then((request) => request(axios, basePath));
        },
        /**
         * 
         * @summary Get workflow run steps
         * @param {string} id 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        getWorkflowRunSteps(id: string, options?: RawAxiosRequestConfig): AxiosPromise<Array<WorkflowStep>> {
            return localVarFp.getWorkflowRunSteps(id, options).then((request) => request(axios, basePath));
        },
        /**
         * 
         * @summary List workflow runs
         * @param {string} [workflowId] 
         * @param {WorkflowRunStatus} [status] 
         * @param {number} [page] 
         * @param {number} [limit] 
         * @param {*} [options] Override http request option.
         * @throws {RequiredError}
         */
        listWorkflowRuns(workflowId?: string, status?: WorkflowRunStatus, page?: number, limit?: number, options?: RawAxiosRequestConfig): AxiosPromise<WorkflowRunList> {
            return localVarFp.listWorkflowRuns(workflowId, status, page, limit, options).then((request) => request(axios, basePath));
        },
    };
};

/**
 * WorkflowRunsApi - object-oriented interface
 * @export
 * @class WorkflowRunsApi
 * @extends {BaseAPI}
 */
export class WorkflowRunsApi extends BaseAPI {
    /**
     * 
     * @summary Get workflow run details
     * @param {string} id 
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     * @memberof WorkflowRunsApi
     */
    public getWorkflowRun(id: string, options?: RawAxiosRequestConfig) {
        return WorkflowRunsApiFp(this.configuration).getWorkflowRun(id, options).then((request) => request(this.axios, this.basePath));
    }

    /**
     * 
     * @summary Get workflow run steps
     * @param {string} id 
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     * @memberof WorkflowRunsApi
     */
    public getWorkflowRunSteps(id: string, options?: RawAxiosRequestConfig) {
        return WorkflowRunsApiFp(this.configuration).getWorkflowRunSteps(id, options).then((request) => request(this.axios, this.basePath));
    }

    /**
     * 
     * @summary List workflow runs
     * @param {string} [workflowId] 
     * @param {WorkflowRunStatus} [status] 
     * @param {number} [page] 
     * @param {number} [limit] 
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     * @memberof WorkflowRunsApi
     */
    public listWorkflowRuns(workflowId?: string, status?: WorkflowRunStatus, page?: number, limit?: number, options?: RawAxiosRequestConfig) {
        return WorkflowRunsApiFp(this.configuration).listWorkflowRuns(workflowId, status, page, limit, options).then((request) => request(this.axios, this.basePath));
    }
}

