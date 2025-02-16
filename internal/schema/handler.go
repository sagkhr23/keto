// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/ory/keto/internal/x"
	"github.com/ory/keto/ketoapi"
	opl "github.com/ory/keto/proto/ory/keto/opl/v1alpha1"
)

type (
	handlerDependencies interface {
		x.LoggerProvider
		x.WriterProvider
	}
	Handler struct {
		d handlerDependencies
	}
)

const RouteBase = "/opl/syntax/check"

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d}
}

func (h *Handler) RegisterSyntaxRoutes(r *x.OPLSyntaxRouter) {
	r.POST(RouteBase, h.postCheckOplSyntax)
}

func (h *Handler) RegisterSyntaxGRPC(s *grpc.Server) {
	opl.RegisterSyntaxServiceServer(s, h)
}

func (h *Handler) Check(_ context.Context, request *opl.CheckRequest) (*opl.CheckResponse, error) {
	_, parseErrors := Parse(string(request.GetContent()))
	apiErrors := make([]*opl.ParseError, len(parseErrors))
	for i, e := range parseErrors {
		apiErrors[i] = e.ToProto()
	}
	return &opl.CheckResponse{ParseErrors: apiErrors}, nil
}

// swagger:route POST /opl/syntax/check syntax checkOplSyntax
//
// # Check the syntax of an OPL file
//
// The OPL file is expected in the body of the request.
//
//		Consumes:
//		- text/plain
//
//		Produces:
//		- application/json
//
//		Schemes: http, https
//
//	 Parameters:
//	   + name: content
//	     in: body
//	     description: the OPL content to check
//	     required: true
//	     type: string
//
//	 Responses:
//		  200: postCheckOplSyntaxResponse
//		  400: genericError
//		  500: genericError
func (h *Handler) postCheckOplSyntax(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError(err.Error())))
		return
	}
	_, parseErrors := Parse(string(data))
	apiErrors := make([]*ketoapi.ParseError, len(parseErrors))
	for i, e := range parseErrors {
		apiErrors[i] = e.ToAPI()
	}
	h.d.Writer().Write(w, r, &ketoapi.CheckOPLSyntaxResponse{Errors: apiErrors})
}
