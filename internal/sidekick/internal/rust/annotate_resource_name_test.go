// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rust

import (
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/internal/api"
)

func TestAnnotateResourceNameField(t *testing.T) {
	model := annotateResourceNameModel(t)
	err := api.CrossReference(model)
	if err != nil {
		t.Fatal(err)
	}
	codec, err := newCodec("protobuf", map[string]string{
		"include-grpc-only-methods": "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = annotateModel(model, codec)

	tests := []struct {
		MethodID string
		Want     string
	}{
		{
			MethodID: ".test.v1.Service.Get",
			Want:     "Some(&req).map(|m| &m.name).map(|s| s.as_str())",
		},
		{
			MethodID: ".test.v1.Service.Create",
			Want:     "Some(&req).map(|m| &m.parent).map(|s| s.as_str())",
		},
		{
			MethodID: ".test.v1.Service.Update",
			Want:     "Some(&req).map(|m| &m.secret).map(|m| &m.name).map(|s| s.as_str())",
		},
	}

	for _, test := range tests {
		m, ok := model.State.MethodByID[test.MethodID]
		if !ok {
			t.Errorf("missing method %s", test.MethodID)
			continue
		}

		ann, ok := m.Codec.(*methodAnnotation)
		if !ok {
			t.Errorf("method %s missing methodAnnotation", test.MethodID)
			continue
		}

		if got := ann.ResourceNameField; got != test.Want {
			t.Errorf("Method %s: ResourceNameField = %q, want %q", test.MethodID, got, test.Want)
		}

		if got, want := ann.DefaultHost, "test.googleapis.com"; got != want {
			t.Errorf("Method %s: DefaultHost = %q, want %q", test.MethodID, got, want)
		}
	}
}

func annotateResourceNameModel(t *testing.T) *api.API {
	t.Helper()
	secret := &api.Message{
		Name: "Secret",
		ID:   ".test.v1.Secret",
		Fields: []*api.Field{
			{Name: "name", Typez: api.STRING_TYPE, ID: ".test.v1.Secret.name"},
		},
	}
	getRequest := &api.Message{
		Name: "GetRequest",
		ID:   ".test.v1.GetRequest",
		Fields: []*api.Field{
			{Name: "name", Typez: api.STRING_TYPE, ID: ".test.v1.GetRequest.name"},
		},
	}
	createRequest := &api.Message{
		Name: "CreateRequest",
		ID:   ".test.v1.CreateRequest",
		Fields: []*api.Field{
			{Name: "parent", Typez: api.STRING_TYPE, ID: ".test.v1.CreateRequest.parent"},
		},
	}
	updateRequest := &api.Message{
		Name: "UpdateRequest",
		ID:   ".test.v1.UpdateRequest",
		Fields: []*api.Field{
			{Name: "secret", Typez: api.MESSAGE_TYPE, TypezID: ".test.v1.Secret", ID: ".test.v1.UpdateRequest.secret"},
		},
	}

	service := &api.Service{
		Name:        "Service",
		ID:          ".test.v1.Service",
		Package:     "test.v1",
		DefaultHost: "test.googleapis.com",
		Methods: []*api.Method{
			{
				Name:        "Get",
				ID:          ".test.v1.Service.Get",
				InputType:   getRequest,
				InputTypeID: ".test.v1.GetRequest",
				OutputTypeID: ".test.v1.Secret",
				PathInfo: &api.PathInfo{
					Bindings: []*api.PathBinding{
						{
							Verb: "GET",
							PathTemplate: api.NewPathTemplate().
								WithLiteral("secrets").
								WithVariableNamed("name"),
						},
					},
				},
			},
			{
				Name:        "Create",
				ID:          ".test.v1.Service.Create",
				InputType:   createRequest,
				InputTypeID: ".test.v1.CreateRequest",
				OutputTypeID: ".test.v1.Secret",
				PathInfo: &api.PathInfo{
					Bindings: []*api.PathBinding{
						{
							Verb: "POST",
							PathTemplate: api.NewPathTemplate().
								WithLiteral("projects").
								WithVariableNamed("parent").
								WithLiteral("secrets"),
						},
					},
				},
			},
			{
				Name:        "Update",
				ID:          ".test.v1.Service.Update",
				InputType:   updateRequest,
				InputTypeID: ".test.v1.UpdateRequest",
				OutputTypeID: ".test.v1.Secret",
				PathInfo: &api.PathInfo{
					Bindings: []*api.PathBinding{
						{
							Verb: "PATCH",
							PathTemplate: api.NewPathTemplate().
								WithLiteral("secrets").
								WithVariableNamed("secret", "name"),
						},
					},
				},
			},
		},
	}

	model := api.NewTestAPI(
		[]*api.Message{secret, getRequest, createRequest, updateRequest},
		[]*api.Enum{},
		[]*api.Service{service})
	return model
}
