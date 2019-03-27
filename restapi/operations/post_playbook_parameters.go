// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
)

// NewPostPlaybookParams creates a new PostPlaybookParams object
// no default values defined in spec.
func NewPostPlaybookParams() PostPlaybookParams {

	return PostPlaybookParams{}
}

// PostPlaybookParams contains all the bound params for the post playbook operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostPlaybook
type PostPlaybookParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: body
	*/
	Manifest string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostPlaybookParams() beforehand.
func (o *PostPlaybookParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body string
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("manifest", "body"))
			} else {
				res = append(res, errors.NewParseError("manifest", "body", "", err))
			}
		} else {
			// no validation required on inline body
			o.Manifest = body
		}
	} else {
		res = append(res, errors.Required("manifest", "body"))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}