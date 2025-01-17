/*
 * fave
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 0.0.0-prealpha
 * Contact: sabyasachi@datafund.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package fave_api

// Get the nearest documents from the collection by vector
type NearestDocumentsByVectorRequest struct {
	// The vector to search for
	Vector []float32 `json:"vector,omitempty"`
	// Name of the collection
	Name string `json:"name,omitempty"`
	Distance float32 `json:"distance,omitempty"`
	Limit float64 `json:"limit,omitempty"`
}
