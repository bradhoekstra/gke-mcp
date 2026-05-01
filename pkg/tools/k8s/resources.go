package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NameHeader is the header for the name of the resource.
	NameHeader = "NAME"
	// ShortNamesHeader is  the header for the short names of the resource.
	ShortNamesHeader = "SHORTNAMES"
	// APIGroupHeader is the header for the API group of the resource.
	APIGroupHeader = "APIGROUP"
	// NamespacedHeader is the header for the namespaced of the resource.
	NamespacedHeader = "NAMESPACED"
	// KindHeader is the header for the kind of the resource.
	KindHeader = "KIND"
	// VerbsHeader is the header for the verbs of the resource.
	VerbsHeader = "VERBS"
	// CategoriesHeader is the header for the categories of the resource.
	CategoriesHeader = "CATEGORIES"
)

// APIResourceInfo is a struct to hold the information about a k8s APIResource.
type APIResourceInfo struct {
	Name       string
	ShortNames []string
	APIGroup   string
	Namespaced bool
	Kind       string
	Verbs      []string
	Categories []string
}

// NewAPIResourceInfoFromAPIResource creates an APIResourceInfo from a metav1.APIResource.
func NewAPIResourceInfoFromAPIResource(resource metav1.APIResource, apiVersion string) APIResourceInfo {
	return APIResourceInfo{
		Name:       resource.Name,
		ShortNames: resource.ShortNames,
		APIGroup:   apiVersion,
		Namespaced: resource.Namespaced,
		Kind:       resource.Kind,
		Verbs:      resource.Verbs,
		Categories: resource.Categories,
	}
}

// AuthStatus is a struct to hold the authorization status of a user for a specific action.
type AuthStatus struct {
	Warning         []string
	Allowed         bool
	Reason          string
	EvaluationError string
}

// AddWarning adds a warning(s) to the authorization status.
func (a *AuthStatus) AddWarning(warnings ...string) {
	a.Warning = append(a.Warning, warnings...)
}

// SetAllowed sets the allowed flag for the authorization status.
func (a *AuthStatus) SetAllowed(allowed bool) {
	a.Allowed = allowed
}

// SetReason sets the reason for the authorization status.
func (a *AuthStatus) SetReason(reason string) {
	a.Reason = reason
}

// SetEvaluationError sets the evaluation error for the authorization status.
func (a *AuthStatus) SetEvaluationError(evaluationError string) {
	a.EvaluationError = evaluationError
}
