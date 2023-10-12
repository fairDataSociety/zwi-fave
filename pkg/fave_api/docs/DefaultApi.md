# {{classname}}

All URIs are relative to */v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**FaveAddDocuments**](DefaultApi.md#FaveAddDocuments) | **Post** /documents | 
[**FaveCreateCollection**](DefaultApi.md#FaveCreateCollection) | **Post** /collections | 
[**FaveDeleteCollection**](DefaultApi.md#FaveDeleteCollection) | **Delete** /collections/{collection} | 
[**FaveGetCollections**](DefaultApi.md#FaveGetCollections) | **Get** /collections | 
[**FaveGetDocuments**](DefaultApi.md#FaveGetDocuments) | **Get** /documents | 
[**FaveGetNearestDocuments**](DefaultApi.md#FaveGetNearestDocuments) | **Post** /nearest-documents | 
[**FaveGetNearestDocumentsByVector**](DefaultApi.md#FaveGetNearestDocumentsByVector) | **Post** /nearest-documents-by-vector | 
[**FaveRoot**](DefaultApi.md#FaveRoot) | **Get** / | 

# **FaveAddDocuments**
> OkResponse FaveAddDocuments(ctx, body)


Add documents into a collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddDocumentsRequest**](AddDocumentsRequest.md)|  | 

### Return type

[**OkResponse**](OKResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveCreateCollection**
> OkResponse FaveCreateCollection(ctx, body)


Create a new collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Collection**](Collection.md)|  | 

### Return type

[**OkResponse**](OKResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveDeleteCollection**
> OkResponse FaveDeleteCollection(ctx, collection)


Delete a collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **collection** | **string**| Collection name | 

### Return type

[**OkResponse**](OKResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveGetCollections**
> []Collection FaveGetCollections(ctx, )


Get all collections.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]Collection**](Collection.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveGetDocuments**
> Document FaveGetDocuments(ctx, property, value, collection)


Retrieve a document based on query parameters

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **property** | **string**| The property to filter the document by | 
  **value** | **string**| The value of the property to filter the document by | 
  **collection** | **string**| The collection to use for this query | 

### Return type

[**Document**](Document.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveGetNearestDocuments**
> NearestDocumentsResponse FaveGetNearestDocuments(ctx, body)


Get nearest documents for a collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NearestDocumentsRequest**](NearestDocumentsRequest.md)|  | 

### Return type

[**NearestDocumentsResponse**](NearestDocumentsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveGetNearestDocumentsByVector**
> NearestDocumentsResponse FaveGetNearestDocumentsByVector(ctx, body)


Get nearest documents for a collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NearestDocumentsByVectorRequest**](NearestDocumentsByVectorRequest.md)|  | 

### Return type

[**NearestDocumentsResponse**](NearestDocumentsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **FaveRoot**
> FaveRoot(ctx, )


Home. Discover the REST API

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

