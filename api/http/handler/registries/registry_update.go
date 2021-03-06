package registries

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

type registryUpdatePayload struct {
	Name               *string
	URL                *string
	Authentication     *bool
	Username           *string
	Password           *string
	UserAccessPolicies portainer.UserAccessPolicies
	TeamAccessPolicies portainer.TeamAccessPolicies
}

func (payload *registryUpdatePayload) Validate(r *http.Request) error {
	return nil
}

// PUT request on /api/registries/:id
func (handler *Handler) registryUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	var payload registryUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	registry, err := handler.DataStore.Registry().Registry(portainer.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a registry with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a registry with the specified identifier inside the database", err}
	}

	if payload.Name != nil {
		registry.Name = *payload.Name
	}

	if payload.URL != nil {
		registries, err := handler.DataStore.Registry().Registries()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
		}
		for _, r := range registries {
			if r.ID != registry.ID && hasSameURL(&r, registry) {
				return &httperror.HandlerError{http.StatusConflict, "Another registry with the same URL already exists", errors.New("A registry is already defined for this URL")}
			}
		}

		registry.URL = *payload.URL
	}

	if payload.Authentication != nil {
		if *payload.Authentication {
			registry.Authentication = true

			if payload.Username != nil {
				registry.Username = *payload.Username
			}

			if payload.Password != nil {
				registry.Password = *payload.Password
			}

		} else {
			registry.Authentication = false
			registry.Username = ""
			registry.Password = ""
		}
	}

	if payload.UserAccessPolicies != nil {
		registry.UserAccessPolicies = payload.UserAccessPolicies
	}

	if payload.TeamAccessPolicies != nil {
		registry.TeamAccessPolicies = payload.TeamAccessPolicies
	}

	err = handler.DataStore.Registry().UpdateRegistry(registry.ID, registry)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist registry changes inside the database", err}
	}

	return response.JSON(w, registry)
}

func hasSameURL(r1, r2 *portainer.Registry) bool {
	if r1.Type != portainer.GitlabRegistry || r2.Type != portainer.GitlabRegistry {
		return r1.URL == r2.URL
	}

	return r1.URL == r2.URL && r1.Gitlab.ProjectPath == r2.Gitlab.ProjectPath
}
